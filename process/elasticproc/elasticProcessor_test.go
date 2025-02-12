package elasticproc

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/accounts"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/block"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/logsevents"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/miniblocks"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/operations"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/statistics"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/tags"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/transactions"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/validators"
)

func newElasticsearchProcessor(elasticsearchWriter DatabaseClientHandler, arguments *ArgElasticProcessor) *elasticProcessor {
	return &elasticProcessor{
		elasticClient:      elasticsearchWriter,
		enabledIndexes:     arguments.EnabledIndexes,
		blockProc:          arguments.BlockProc,
		transactionsProc:   arguments.TransactionsProc,
		miniblocksProc:     arguments.MiniblocksProc,
		accountsProc:       arguments.AccountsProc,
		validatorsProc:     arguments.ValidatorsProc,
		statisticsProc:     arguments.StatisticsProc,
		logsAndEventsProc:  arguments.LogsAndEventsProc,
		indexTokensHandler: arguments.IndexTokensHandler,
	}
}

func createEmptyOutportBlockWithHeader() *outport.OutportBlockWithHeader {
	signerIndexes := []uint64{0, 1}
	header := &dataBlock.Header{Nonce: 1}
	return &outport.OutportBlockWithHeader{
		Header: header,
		OutportBlock: &outport.OutportBlock{
			BlockData: &outport.BlockData{
				Body: &dataBlock.Body{},
			},
			SignersIndexes:       signerIndexes,
			HeaderGasConsumption: &outport.HeaderGasConsumption{},
			TransactionPool:      &outport.TransactionPool{},
		},
	}
}

func createMockElasticProcessorArgs() *ArgElasticProcessor {
	balanceConverter, _ := converters.NewBalanceConverter(10)

	acp, _ := accounts.NewAccountsProcessor(&mock.PubkeyConverterMock{}, balanceConverter)
	bp, _ := block.NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})
	mp, _ := miniblocks.NewMiniblocksProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})
	vp, _ := validators.NewValidatorsProcessor(mock.NewPubkeyConverterMock(32), 0)
	args := logsevents.ArgsLogsAndEventsProcessor{
		PubKeyConverter:  &mock.PubkeyConverterMock{},
		Marshalizer:      &mock.MarshalizerMock{},
		BalanceConverter: balanceConverter,
		Hasher:           &mock.HasherMock{},
	}
	lp, _ := logsevents.NewLogsAndEventsProcessor(args)
	op, _ := operations.NewOperationsProcessor()

	return &ArgElasticProcessor{
		DBClient: &mock.DatabaseWriterStub{},
		EnabledIndexes: map[string]struct{}{
			dataindexer.BlockIndex: {}, dataindexer.TransactionsIndex: {}, dataindexer.MiniblocksIndex: {}, dataindexer.ValidatorsIndex: {}, dataindexer.RoundsIndex: {}, dataindexer.AccountsIndex: {}, dataindexer.RatingIndex: {}, dataindexer.AccountsHistoryIndex: {},
		},
		ValidatorsProc:     vp,
		StatisticsProc:     statistics.NewStatisticsProcessor(),
		TransactionsProc:   &mock.DBTransactionProcessorStub{},
		MiniblocksProc:     mp,
		AccountsProc:       acp,
		BlockProc:          bp,
		LogsAndEventsProc:  lp,
		OperationsProc:     op,
		IndexTokensHandler: &IndexTokenHandlerMock{},
	}
}

func newTestBlockBody() *dataBlock.Body {
	return &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{
				TxHashes: [][]byte{
					[]byte("tx1"),
					[]byte("tx2"),
				},
				ReceiverShardID: 2,
				SenderShardID:   2,
			},
			{
				TxHashes: [][]byte{
					[]byte("tx3"),
				},
				ReceiverShardID: 4,
				SenderShardID:   1,
			},
		},
	}
}

func TestNewElasticProcessor(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local error")
	tests := []struct {
		name  string
		args  func() *ArgElasticProcessor
		exErr error
	}{
		{
			name: "NilArguments",
			args: func() *ArgElasticProcessor {
				return nil
			},
			exErr: dataindexer.ErrNilElasticProcessorArguments,
		},
		{
			name: "NilEnabledIndexesMap",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.EnabledIndexes = nil
				return arguments
			},
			exErr: dataindexer.ErrNilEnabledIndexesMap,
		},
		{
			name: "NilDatabaseClient",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.DBClient = nil
				return arguments
			},
			exErr: dataindexer.ErrNilDatabaseClient,
		},
		{
			name: "NilStatisticProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.StatisticsProc = nil
				return arguments
			},
			exErr: dataindexer.ErrNilStatisticHandler,
		},
		{
			name: "NilBlockProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.BlockProc = nil
				return arguments
			},
			exErr: dataindexer.ErrNilBlockHandler,
		},
		{
			name: "NilAccountsProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.AccountsProc = nil
				return arguments
			},
			exErr: dataindexer.ErrNilAccountsHandler,
		},
		{
			name: "NilMiniblocksProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.MiniblocksProc = nil
				return arguments
			},
			exErr: dataindexer.ErrNilMiniblocksHandler,
		},
		{
			name: "NilValidatorsProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.ValidatorsProc = nil
				return arguments
			},
			exErr: dataindexer.ErrNilValidatorsHandler,
		},
		{
			name: "NilTxsProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.TransactionsProc = nil
				return arguments
			},
			exErr: dataindexer.ErrNilTransactionsHandler,
		},
		{
			name: "InitError",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.DBClient = &mock.DatabaseWriterStub{
					CheckAndCreateIndexCalled: func(index string) error {
						return expectedErr
					},
				}
				return arguments
			},
			exErr: expectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewElasticProcessor(tt.args())
			require.True(t, errors.Is(err, tt.exErr))
		})
	}
}

func TestNewElasticProcessorWithKibana(t *testing.T) {
	args := createMockElasticProcessorArgs()
	args.UseKibana = true
	args.DBClient = &mock.DatabaseWriterStub{}

	elasticProc, err := NewElasticProcessor(args)
	require.NoError(t, err)
	require.NotNil(t, elasticProc)
}

func TestElasticProcessor_RemoveHeader(t *testing.T) {
	called := false

	args := createMockElasticProcessorArgs()
	args.DBClient = &mock.DatabaseWriterStub{
		DoQueryRemoveCalled: func(index string, body *bytes.Buffer) error {
			called = true
			return nil
		},
	}

	args.BlockProc, _ = block.NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	elasticProc, err := NewElasticProcessor(args)
	require.NoError(t, err)

	err = elasticProc.RemoveHeader(&dataBlock.Header{})
	require.Nil(t, err)
	require.True(t, called)
}

func TestElasticProcessor_RemoveMiniblocks(t *testing.T) {
	called := false

	mb1 := &dataBlock.MiniBlock{
		Type: dataBlock.PeerBlock,
	}
	mb2 := &dataBlock.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
	} // should be removed
	mb3 := &dataBlock.MiniBlock{
		ReceiverShardID: 1,
		SenderShardID:   1,
	} // should be removed
	mb4 := &dataBlock.MiniBlock{
		ReceiverShardID: 1,
		SenderShardID:   0,
	} // should NOT be removed

	args := createMockElasticProcessorArgs()

	mbHash2, _ := core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, mb2)
	mbHash3, _ := core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, mb3)

	args.DBClient = &mock.DatabaseWriterStub{
		DoQueryRemoveCalled: func(index string, body *bytes.Buffer) error {
			called = true
			bodyStr := body.String()
			require.True(t, strings.Contains(bodyStr, hex.EncodeToString(mbHash2)))
			require.True(t, strings.Contains(bodyStr, hex.EncodeToString(mbHash3)))
			return nil
		},
	}

	args.MiniblocksProc, _ = miniblocks.NewMiniblocksProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	elasticProc, err := NewElasticProcessor(args)
	require.NoError(t, err)

	header := &dataBlock.Header{
		ShardID: 1,
		MiniBlockHeaders: []dataBlock.MiniBlockHeader{
			{
				Hash: []byte("hash1"),
			},
			{
				Hash: []byte("hash2"),
			},
			{
				Hash: []byte("hash3"),
			},
			{
				Hash: []byte("hash4"),
			},
		},
	}
	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			mb1, mb2, mb3, mb4,
		},
	}
	err = elasticProc.RemoveMiniblocks(header, body)
	require.Nil(t, err)
	require.True(t, called)
}

func TestElasticseachDatabaseSaveHeader_RequestError(t *testing.T) {
	localErr := errors.New("localErr")

	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localErr
		},
	}
	arguments.BlockProc, _ = block.NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})
	elasticDatabase := newElasticsearchProcessor(dbWriter, arguments)

	err := elasticDatabase.SaveHeader(createEmptyOutportBlockWithHeader())
	require.Equal(t, localErr, err)
}

func TestElasticseachSaveTransactions(t *testing.T) {
	localErr := errors.New("localErr")
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localErr
		},
	}

	body := newTestBlockBody()
	header := &dataBlock.Header{Nonce: 1, TxCount: 2}

	bc, _ := converters.NewBalanceConverter(18)
	args := &transactions.ArgsTransactionProcessor{
		AddressPubkeyConverter: mock.NewPubkeyConverterMock(32),
		Hasher:                 &mock.HasherMock{},
		Marshalizer:            &mock.MarshalizerMock{},
		BalanceConverter:       bc,
		TxHashExtractor:        transactions.NewTxHashExtractor(),
		RewardTxData:           &mock.RewardTxDataMock{},
	}
	txDbProc, _ := transactions.NewTransactionsProcessor(args)
	arguments.TransactionsProc = txDbProc

	outportBlock := createEmptyOutportBlockWithHeader()
	outportBlock.Header = header
	outportBlock.BlockData.Body = body
	outportBlock.TransactionPool.Transactions = map[string]*outport.TxInfo{
		hex.EncodeToString([]byte("tx1")): {Transaction: &transaction.Transaction{}, FeeInfo: &outport.FeeInfo{}},
	}

	elasticDatabase := newElasticsearchProcessor(dbWriter, arguments)
	err := elasticDatabase.SaveTransactions(outportBlock)
	require.Equal(t, localErr, err)
}

func TestElasticProcessor_SaveValidatorsRating(t *testing.T) {
	localErr := errors.New("localErr")

	arguments := createMockElasticProcessorArgs()
	arguments.DBClient = &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localErr
		},
	}

	arguments.ValidatorsProc, _ = validators.NewValidatorsProcessor(mock.NewPubkeyConverterMock(32), 0)
	elasticProc, _ := NewElasticProcessor(arguments)

	err := elasticProc.SaveValidatorsRating(&outport.ValidatorsRating{
		ShardID:              0,
		Epoch:                1,
		ValidatorsRatingInfo: []*outport.ValidatorRatingInfo{{}},
	})
	require.Equal(t, localErr, err)
}

func TestElasticProcessor_SaveMiniblocks(t *testing.T) {
	localErr := errors.New("localErr")

	arguments := createMockElasticProcessorArgs()
	arguments.DBClient = &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localErr
		},
		DoMultiGetCalled: func(hashes []string, index string, withSource bool, response interface{}) error {
			return nil
		},
	}

	arguments.MiniblocksProc, _ = miniblocks.NewMiniblocksProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})
	elasticProc, _ := NewElasticProcessor(arguments)

	header := &dataBlock.Header{}
	body := &dataBlock.Body{MiniBlocks: dataBlock.MiniBlockSlice{
		{SenderShardID: 0, ReceiverShardID: 1},
	}}
	err := elasticProc.SaveMiniblocks(header, body.MiniBlocks)
	require.Equal(t, localErr, err)
}

func TestElasticsearch_saveShardValidatorsPubKeys_RequestError(t *testing.T) {
	shardID := uint32(0)
	epoch := uint32(0)
	valPubKeys := [][]byte{[]byte("key1"), []byte("key2")}
	localErr := errors.New("localErr")
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localErr
		},
	}
	arguments.ValidatorsProc, _ = validators.NewValidatorsProcessor(mock.NewPubkeyConverterMock(32), 0)
	elasticDatabase := newElasticsearchProcessor(dbWriter, arguments)

	err := elasticDatabase.SaveShardValidatorsPubKeys(&outport.ValidatorsPubKeys{
		Epoch: epoch,
		ShardValidatorsPubKeys: map[uint32]*outport.PubKeys{
			shardID: {Keys: valPubKeys},
		},
	})
	require.Equal(t, localErr, err)
}

func TestElasticsearch_saveRoundInfoRequestError(t *testing.T) {
	roundInfo := &outport.RoundInfo{}
	localError := errors.New("local err")
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localError
		},
	}
	elasticDatabase := newElasticsearchProcessor(dbWriter, arguments)

	err := elasticDatabase.SaveRoundsInfo(&outport.RoundsInfo{RoundsInfo: []*outport.RoundInfo{roundInfo}})
	require.Equal(t, localError, err)

}

func TestElasticProcessor_RemoveTransactions(t *testing.T) {
	arguments := createMockElasticProcessorArgs()

	called := false
	txsHashes := [][]byte{[]byte("txHas1"), []byte("txHash2")}
	expectedHashes := []string{hex.EncodeToString(txsHashes[0]), hex.EncodeToString(txsHashes[1])}
	dbWriter := &mock.DatabaseWriterStub{
		DoQueryRemoveCalled: func(index string, body *bytes.Buffer) error {
			bodyStr := body.String()
			require.Contains(t, []string{dataindexer.TransactionsIndex, dataindexer.OperationsIndex, dataindexer.LogsIndex, dataindexer.EventsIndex}, index)
			if index != dataindexer.EventsIndex {
				require.True(t, strings.Contains(bodyStr, expectedHashes[0]))
				require.True(t, strings.Contains(bodyStr, expectedHashes[1]))
				called = true
			} else {
				require.Equal(t,
					`{"query": {"bool": {"must": [{"match": {"shardID": {"query": 4294967295,"operator": "AND"}}},{"match": {"timestamp": {"query": "0","operator": "AND"}}}]}}}`,
					body.String(),
				)
			}

			return nil
		},
	}

	args := &transactions.ArgsTransactionProcessor{
		AddressPubkeyConverter: mock.NewPubkeyConverterMock(32),
		Hasher:                 &mock.HasherMock{},
		Marshalizer:            &mock.MarshalizerMock{},
	}
	txDbProc, _ := transactions.NewTransactionsProcessor(args)

	arguments.TransactionsProc = txDbProc

	elasticSearchProc := newElasticsearchProcessor(dbWriter, arguments)

	header := &dataBlock.Header{ShardID: core.MetachainShardId, MiniBlockHeaders: []dataBlock.MiniBlockHeader{{}}}
	blk := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				TxHashes:        txsHashes,
				Type:            dataBlock.RewardsBlock,
				SenderShardID:   core.MetachainShardId,
				ReceiverShardID: 0,
			},
			{
				Type: dataBlock.TxBlock,
			},
		},
	}

	err := elasticSearchProc.RemoveTransactions(header, blk)
	require.Nil(t, err)
	require.True(t, called)
}

func TestElasticProcessor_IndexEpochInfoData(t *testing.T) {
	called := false
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			called = true
			return nil
		},
	}

	elasticSearchProc := newElasticsearchProcessor(dbWriter, arguments)
	elasticSearchProc.enabledIndexes[dataindexer.EpochInfoIndex] = struct{}{}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	shardHeader := &dataBlock.Header{
		ShardID: core.MetachainShardId,
	}
	err := elasticSearchProc.indexEpochInfoData(shardHeader, buffSlice)
	require.True(t, errors.Is(err, dataindexer.ErrHeaderTypeAssertion))

	err = elasticSearchProc.SaveHeader(createEmptyOutportBlockWithHeader())
	require.Nil(t, err)
	require.True(t, called)
}

func TestElasticProcessor_SaveTransactionNoDataShouldNotDoRequest(t *testing.T) {
	called := false
	arguments := createMockElasticProcessorArgs()
	arguments.TransactionsProc = &mock.DBTransactionProcessorStub{
		PrepareTransactionsForDatabaseCalled: func(mbs []*dataBlock.MiniBlock, header coreData.HeaderHandler, pool *outport.TransactionPool) *data.PreparedResults {
			return &data.PreparedResults{
				Transactions: nil,
				ScResults:    nil,
				Receipts:     nil,
			}
		},
		SerializeScResultsCalled: func(scrs []*data.ScResult, _ *data.BufferSlice, _ string) error {
			return nil
		},
	}
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			called = true
			return nil
		},
	}

	elasticSearchProc := newElasticsearchProcessor(dbWriter, arguments)
	elasticSearchProc.enabledIndexes[dataindexer.ScResultsIndex] = struct{}{}

	err := elasticSearchProc.SaveTransactions(createEmptyOutportBlockWithHeader())
	require.Nil(t, err)
	require.False(t, called)
}

func TestElasticProcessor_IndexAlteredAccounts(t *testing.T) {
	called := false
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return nil
		},
		DoMultiGetCalled: func(ids []string, index string, withSource bool, response interface{}) error {
			return nil
		},
	}
	arguments := createMockElasticProcessorArgs()
	arguments.AccountsProc = &mock.DBAccountsHandlerStub{
		SerializeAccountsHistoryCalled: func(accounts map[string]*data.AccountBalanceHistory, _ *data.BufferSlice, _ string) error {
			called = true
			return nil
		},
	}
	elasticSearchProc := newElasticsearchProcessor(dbWriter, arguments)
	elasticSearchProc.enabledIndexes[dataindexer.AccountsESDTIndex] = struct{}{}
	elasticSearchProc.enabledIndexes[dataindexer.AccountsESDTHistoryIndex] = struct{}{}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	tagsCount := tags.NewTagsCount()
	err := elasticSearchProc.indexAlteredAccounts(100, nil, nil, buffSlice, tagsCount, 0)
	require.Nil(t, err)
	require.True(t, called)
}
