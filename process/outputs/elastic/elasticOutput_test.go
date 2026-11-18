package elastic

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataprocessor"
	"github.com/multiversx/mx-chain-es-indexer-go/process/outputs"
	"github.com/multiversx/mx-chain-es-indexer-go/process/outputs/elastic/templatesAndPolicies"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/accounts"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/block"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/converters"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/logsevents"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/miniblocks"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/operations"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/statistics"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/tags"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/transactions"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/validators"
	"github.com/stretchr/testify/require"
)

func newElasticsearchProcessor(elasticsearchWriter DatabaseClientHandler, arguments *ArgElasticOutput) *elasticOutput {
	return &elasticOutput{
		elasticClient:       elasticsearchWriter,
		enabledIndexes:      arguments.EnabledIndexes,
		blockProc:           arguments.BlockProc,
		transactionsProc:    arguments.TransactionsProc,
		miniblocksProc:      arguments.MiniblocksProc,
		accountsProc:        arguments.AccountsProc,
		validatorsProc:      arguments.ValidatorsProc,
		statisticsProc:      arguments.StatisticsProc,
		logsAndEventsProc:   arguments.LogsAndEventsProc,
		numWritesInParallel: 1,
	}
}

// prepareBlock prepares the data of the block in the same way the data processor does it before calling the outputs
func prepareBlock(t *testing.T, arguments *ArgElasticOutput, obh *outport.OutportBlockWithHeader) *data.PreparedBlock {
	transactionsProc, ok := arguments.TransactionsProc.(dataprocessor.TransactionsHandler)
	require.True(t, ok)
	logsAndEventsProc, ok := arguments.LogsAndEventsProc.(dataprocessor.LogsAndEventsHandler)
	require.True(t, ok)
	operationsProc, ok := arguments.OperationsProc.(dataprocessor.OperationsHandler)
	require.True(t, ok)

	dataProcessor, err := dataprocessor.NewDataProcessor(dataprocessor.ArgDataProcessor{
		TransactionsProc:  transactionsProc,
		LogsAndEventsProc: logsAndEventsProc,
		OperationsProc:    operationsProc,
		Outputs:           []outputs.OutputWithPolicy{{Name: "test", Output: &mock.OutputStub{}}},
	})
	require.Nil(t, err)

	return dataProcessor.PrepareBlock(obh)
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

func createMockElasticProcessorArgs() *ArgElasticOutput {
	balanceConverter, _ := converters.NewBalanceConverter(10)

	acp, _ := accounts.NewAccountsProcessor(&mock.PubkeyConverterMock{}, balanceConverter)
	bp, _ := block.NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.PubkeyConverterMock{})
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

	return &ArgElasticOutput{
		DBClient: &mock.DatabaseWriterStub{},
		EnabledIndexes: map[string]struct{}{
			dataindexer.BlockIndex: {}, dataindexer.TransactionsIndex: {}, dataindexer.MiniblocksIndex: {}, dataindexer.ValidatorsIndex: {}, dataindexer.RoundsIndex: {}, dataindexer.AccountsIndex: {}, dataindexer.RatingIndex: {}, dataindexer.AccountsHistoryIndex: {},
		},
		ValidatorsProc:      vp,
		StatisticsProc:      statistics.NewStatisticsProcessor(),
		TransactionsProc:    &mock.DBTransactionProcessorStub{},
		MiniblocksProc:      mp,
		AccountsProc:        acp,
		BlockProc:           bp,
		LogsAndEventsProc:   lp,
		OperationsProc:      op,
		MappingsHandler:     templatesAndPolicies.NewTemplatesAndPolicyReader(false, "", []string{dataindexer.BlockIndex}, nil),
		NumWritesInParallel: 1,
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

func TestNewElasticOutput(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local error")
	tests := []struct {
		name  string
		args  func() *ArgElasticOutput
		exErr error
	}{
		{
			name: "NilArguments",
			args: func() *ArgElasticOutput {
				return nil
			},
			exErr: dataindexer.ErrNilElasticOutputArguments,
		},
		{
			name: "NilEnabledIndexesMap",
			args: func() *ArgElasticOutput {
				arguments := createMockElasticProcessorArgs()
				arguments.EnabledIndexes = nil
				return arguments
			},
			exErr: dataindexer.ErrNilEnabledIndexesMap,
		},
		{
			name: "NilDatabaseClient",
			args: func() *ArgElasticOutput {
				arguments := createMockElasticProcessorArgs()
				arguments.DBClient = nil
				return arguments
			},
			exErr: dataindexer.ErrNilDatabaseClient,
		},
		{
			name: "NilStatisticProc",
			args: func() *ArgElasticOutput {
				arguments := createMockElasticProcessorArgs()
				arguments.StatisticsProc = nil
				return arguments
			},
			exErr: dataindexer.ErrNilStatisticHandler,
		},
		{
			name: "NilBlockProc",
			args: func() *ArgElasticOutput {
				arguments := createMockElasticProcessorArgs()
				arguments.BlockProc = nil
				return arguments
			},
			exErr: dataindexer.ErrNilBlockHandler,
		},
		{
			name: "NilAccountsProc",
			args: func() *ArgElasticOutput {
				arguments := createMockElasticProcessorArgs()
				arguments.AccountsProc = nil
				return arguments
			},
			exErr: dataindexer.ErrNilAccountsHandler,
		},
		{
			name: "NilMiniblocksProc",
			args: func() *ArgElasticOutput {
				arguments := createMockElasticProcessorArgs()
				arguments.MiniblocksProc = nil
				return arguments
			},
			exErr: dataindexer.ErrNilMiniblocksHandler,
		},
		{
			name: "NilValidatorsProc",
			args: func() *ArgElasticOutput {
				arguments := createMockElasticProcessorArgs()
				arguments.ValidatorsProc = nil
				return arguments
			},
			exErr: dataindexer.ErrNilValidatorsHandler,
		},
		{
			name: "NilTxsProc",
			args: func() *ArgElasticOutput {
				arguments := createMockElasticProcessorArgs()
				arguments.TransactionsProc = nil
				return arguments
			},
			exErr: dataindexer.ErrNilTransactionsHandler,
		},
		{
			name: "NilMappingsHandler",
			args: func() *ArgElasticOutput {
				arguments := createMockElasticProcessorArgs()
				arguments.MappingsHandler = nil
				return arguments
			},
			exErr: dataindexer.ErrNilMappingsHandler,
		},
		{
			name: "InitError",
			args: func() *ArgElasticOutput {
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
			_, err := NewElasticOutput(tt.args())
			require.True(t, errors.Is(err, tt.exErr))
		})
	}
}

func TestNewElasticOutput_ShouldWork(t *testing.T) {
	args := createMockElasticProcessorArgs()
	args.DBClient = &mock.DatabaseWriterStub{}

	elasticProc, err := NewElasticOutput(args)
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

	args.BlockProc, _ = block.NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.PubkeyConverterMock{})

	elasticProc, err := NewElasticOutput(args)
	require.NoError(t, err)

	err = elasticProc.RemoveHeader(&dataBlock.Header{})
	require.Nil(t, err)
	require.True(t, called)
}

func TestElasticProcessor_RemoveMiniblocks(t *testing.T) {
	called := false

	mb2 := &dataBlock.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
	} // should be removed
	mb3 := &dataBlock.MiniBlock{
		ReceiverShardID: 1,
		SenderShardID:   1,
	} // should be removed

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

	elasticProc, err := NewElasticOutput(args)
	require.NoError(t, err)

	header := &dataBlock.Header{
		ShardID: 1,
		MiniBlockHeaders: []dataBlock.MiniBlockHeader{
			{
				Type: dataBlock.PeerBlock,
				Hash: []byte("hash1"),
			},
			{
				Hash:            mbHash2,
				ReceiverShardID: 0,
				SenderShardID:   1,
			},
			{
				Hash:            mbHash3,
				ReceiverShardID: 1,
				SenderShardID:   1,
			},
			{
				Hash:            []byte("hash4"),
				ReceiverShardID: 1,
				SenderShardID:   0,
			},
		},
	}
	err = elasticProc.RemoveMiniblocks(header)
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
	arguments.BlockProc, _ = block.NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.PubkeyConverterMock{})
	elasticDatabase := newElasticsearchProcessor(dbWriter, arguments)

	err := elasticDatabase.SaveHeader(createEmptyOutportBlockWithHeader())
	require.NotNil(t, err)
	require.Equal(t, localErr.Error(), err.Error())
}

func TestElasticseachSaveTransactions(t *testing.T) {
	localErr := errors.New("localErr")
	arguments := createMockElasticProcessorArgs()
	bulkRequestCalls := 0
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			bulkRequestCalls++
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
	err := elasticDatabase.SaveTransactions(prepareBlock(t, arguments, outportBlock))
	require.NotNil(t, err)
	require.Equal(t, localErr.Error(), err.Error())
	require.Equal(t, 1, bulkRequestCalls)
}

func TestElasticProcessor_SaveTransactionsIndexesHeaderRoundInValues(t *testing.T) {
	var bulkRequests []string
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, _ string) error {
			bulkRequests = append(bulkRequests, buff.String())
			return nil
		},
	}
	arguments := createMockElasticProcessorArgs()
	arguments.EnabledIndexes[dataindexer.ValuesIndex] = struct{}{}
	arguments.TransactionsProc = &mock.DBTransactionProcessorStub{
		PrepareTransactionsForDatabaseCalled: func(_ []*dataBlock.MiniBlock, _ *data.HeaderData, _ *outport.TransactionPool) *data.PreparedResults {
			return &data.PreparedResults{ScResults: []*data.ScResult{{}}}
		},
		SerializeScResultsCalled: func(_ []*data.ScResult, buffers *data.BufferSlice, _ string) error {
			return buffers.PutData([]byte("transaction-meta\n"), []byte("transaction-data"))
		},
	}
	arguments.EnabledIndexes[dataindexer.ScResultsIndex] = struct{}{}
	obh := createEmptyOutportBlockWithHeader()
	obh.Header = &dataBlock.Header{ShardID: 2, Round: 42, Nonce: 17}
	obh.BlockData.TimestampMs = 123000

	processor := newElasticsearchProcessor(dbWriter, arguments)
	err := processor.SaveTransactions(prepareBlock(t, arguments, obh))

	require.NoError(t, err)
	require.Len(t, bulkRequests, 2)
	require.Contains(t, bulkRequests[0], "transaction-data")
	require.NotContains(t, bulkRequests[0], `"_index":"values"`)
	require.Contains(t, bulkRequests[1], `"_index":"values", "_id" : "shard_2"`)
	require.Contains(t, bulkRequests[1], `{"key":"shard_2","value":"42","timestampMs":123000}`)
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
	elasticProc, _ := NewElasticOutput(arguments)

	err := elasticProc.SaveValidatorsRating(&outport.ValidatorsRating{
		ShardID:              0,
		Epoch:                1,
		ValidatorsRatingInfo: []*outport.ValidatorRatingInfo{{}},
	})
	require.NotNil(t, err)
	require.Equal(t, localErr.Error(), err.Error())
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
	elasticProc, _ := NewElasticOutput(arguments)

	header := &dataBlock.Header{}
	body := &dataBlock.Body{MiniBlocks: dataBlock.MiniBlockSlice{
		{SenderShardID: 0, ReceiverShardID: 1},
	}}

	ob := createEmptyOutportBlockWithHeader()
	ob.Header = header
	ob.BlockData.Body = body

	err := elasticProc.SaveMiniblocks(ob)
	require.NotNil(t, err)
	require.Equal(t, localErr.Error(), err.Error())
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
	require.NotNil(t, err)
	require.Equal(t, localErr.Error(), err.Error())
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
					`{"query": {"bool": {"must": [{"match": {"shardID": {"query": 4294967295,"operator": "AND"}}},{"match": {"timestampMs": {"query": "0","operator": "AND"}}}]}}}`,
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

	err := elasticSearchProc.RemoveTransactions(header, blk, 0)
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

func TestElasticProcessor_SaveTransactionsIndexesRoundWhenNoTransactions(t *testing.T) {
	called := false
	var bulkRequest string
	arguments := createMockElasticProcessorArgs()
	arguments.TransactionsProc = &mock.DBTransactionProcessorStub{
		PrepareTransactionsForDatabaseCalled: func(mbs []*dataBlock.MiniBlock, headerData *data.HeaderData, pool *outport.TransactionPool) *data.PreparedResults {
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
			bulkRequest += buff.String()
			return nil
		},
	}

	elasticSearchProc := newElasticsearchProcessor(dbWriter, arguments)
	elasticSearchProc.enabledIndexes[dataindexer.ValuesIndex] = struct{}{}
	elasticSearchProc.enabledIndexes[dataindexer.ScResultsIndex] = struct{}{}

	obh := createEmptyOutportBlockWithHeader()
	obh.Header = &dataBlock.Header{ShardID: 4, Round: 83}
	err := elasticSearchProc.SaveTransactions(prepareBlock(t, arguments, obh))
	require.Nil(t, err)
	require.True(t, called)
	require.Contains(t, bulkRequest, `{"key":"shard_4","value":"83"}`)
}

func TestElasticProcessor_IndexShardRoundWhenValuesIndexDisabled(t *testing.T) {
	bulkRequestCalled := false
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(_ *bytes.Buffer, _ string) error {
			bulkRequestCalled = true
			return nil
		},
	}
	processor := newElasticsearchProcessor(dbWriter, createMockElasticProcessorArgs())

	err := processor.indexShardRound(4, 83, 0)

	require.NoError(t, err)
	require.False(t, bulkRequestCalled)
}

func TestElasticProcessor_RevertShardRound(t *testing.T) {
	prevHash := []byte("prev-hash")
	expectedPrevHashHex := hex.EncodeToString(prevHash)

	t.Run("updates values index with previous block round", func(t *testing.T) {
		var bulkRequest string
		dbWriter := &mock.DatabaseWriterStub{
			DoMultiGetCalled: func(ids []string, index string, withSource bool, response interface{}) error {
				require.Equal(t, []string{expectedPrevHashHex}, ids)
				require.Equal(t, dataindexer.BlockIndex, index)
				require.True(t, withSource)

				responseBlocks := response.(*data.ResponseBlocks)
				responseBlocks.Docs = []data.ResponseBlockDB{
					{
						Found:  true,
						Source: data.Block{Round: 100, TimestampMs: 12345},
					},
				}
				return nil
			},
			DoBulkRequestCalled: func(buff *bytes.Buffer, _ string) error {
				bulkRequest += buff.String()
				return nil
			},
		}

		arguments := createMockElasticProcessorArgs()
		arguments.EnabledIndexes[dataindexer.ValuesIndex] = struct{}{}
		processor := newElasticsearchProcessor(dbWriter, arguments)

		err := processor.RevertShardRound(&dataBlock.Header{
			ShardID:  1,
			Round:    101,
			PrevHash: prevHash,
		})
		require.NoError(t, err)
		require.Contains(t, bulkRequest, `"_index":"values", "_id" : "shard_1"`)
		require.Contains(t, bulkRequest, `{"key":"shard_1","value":"100","timestampMs":12345}`)
	})

	t.Run("skips when values index is disabled", func(t *testing.T) {
		multiGetCalled := false
		dbWriter := &mock.DatabaseWriterStub{
			DoMultiGetCalled: func(_ []string, _ string, _ bool, _ interface{}) error {
				multiGetCalled = true
				return nil
			},
		}

		processor := newElasticsearchProcessor(dbWriter, createMockElasticProcessorArgs())
		err := processor.RevertShardRound(&dataBlock.Header{
			ShardID:  1,
			Round:    101,
			PrevHash: prevHash,
		})
		require.NoError(t, err)
		require.False(t, multiGetCalled)
	})

	t.Run("skips when previous block is not found", func(t *testing.T) {
		bulkRequestCalled := false
		dbWriter := &mock.DatabaseWriterStub{
			DoMultiGetCalled: func(_ []string, _ string, _ bool, response interface{}) error {
				responseBlocks := response.(*data.ResponseBlocks)
				responseBlocks.Docs = []data.ResponseBlockDB{{Found: false}}
				return nil
			},
			DoBulkRequestCalled: func(_ *bytes.Buffer, _ string) error {
				bulkRequestCalled = true
				return nil
			},
		}

		arguments := createMockElasticProcessorArgs()
		arguments.EnabledIndexes[dataindexer.ValuesIndex] = struct{}{}
		processor := newElasticsearchProcessor(dbWriter, arguments)

		err := processor.RevertShardRound(&dataBlock.Header{
			ShardID:  1,
			Round:    101,
			PrevHash: prevHash,
		})
		require.NoError(t, err)
		require.False(t, bulkRequestCalled)
	})
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
	err := elasticSearchProc.indexAlteredAccounts(nil, nil, buffSlice, tagsCount, 0, 0)
	require.Nil(t, err)
	require.True(t, called)
}

func TestElasticProcessor_DoBulkRequests_SuccessfulParallelProcessing(t *testing.T) {
	arguments := createMockElasticProcessorArgs()
	arguments.NumWritesInParallel = 3

	processedBuffers := make(map[string]bool)
	processedBuffersMutex := &sync.Mutex{}

	arguments.DBClient = &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			processedBuffersMutex.Lock()
			processedBuffers[buff.String()] = true
			processedBuffersMutex.Unlock()
			return nil
		},
	}

	elasticProc, _ := NewElasticOutput(arguments)

	// Create 5 test buffers
	buffers := make([]*bytes.Buffer, 5)
	for i := 0; i < 5; i++ {
		buffers[i] = bytes.NewBufferString(fmt.Sprintf("test-buffer-%d", i))
	}

	err := elasticProc.doBulkRequests("test-index", buffers, 0)
	require.Nil(t, err)

	// Verify all buffers were processed
	processedBuffersMutex.Lock()
	defer processedBuffersMutex.Unlock()
	require.Equal(t, 5, len(processedBuffers))
	for i := 0; i < 5; i++ {
		require.True(t, processedBuffers[fmt.Sprintf("test-buffer-%d", i)])
	}
}

func TestElasticProcessor_DoBulkRequests_WorkerCancellationOnError(t *testing.T) {
	arguments := createMockElasticProcessorArgs()
	arguments.NumWritesInParallel = 3

	processedBuffers := make(map[string]bool)
	processedBuffersMutex := &sync.Mutex{}
	expectedErr := errors.New("expected error")

	arguments.DBClient = &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			if buff.String() == "test-buffer-1" {
				return expectedErr
			}

			processedBuffersMutex.Lock()
			processedBuffers[buff.String()] = true
			processedBuffersMutex.Unlock()

			return nil
		},
	}

	elasticProc, _ := NewElasticOutput(arguments)

	buffers := make([]*bytes.Buffer, 5)
	for i := 0; i < 5; i++ {
		buffers[i] = bytes.NewBufferString(fmt.Sprintf("test-buffer-%d", i))
	}

	err := elasticProc.doBulkRequests("test-index", buffers, 0)
	require.NotNil(t, err)
	require.Equal(t, "expected error", err.Error())

	processedBuffersMutex.Lock()
	defer processedBuffersMutex.Unlock()
	require.True(t, len(processedBuffers) > 0)
	require.True(t, processedBuffers["test-buffer-0"])
	require.False(t, processedBuffers["test-buffer-1"])
	require.True(t, processedBuffers["test-buffer-2"])
	require.True(t, processedBuffers["test-buffer-3"])
	require.True(t, processedBuffers["test-buffer-4"])
}

func TestElasticProcessor_DoBulkRequests_DatabaseError(t *testing.T) {
	arguments := createMockElasticProcessorArgs()
	arguments.NumWritesInParallel = 20

	expectedErr := errors.New("database connection error")
	processedBuffers := make(map[string]int)
	processedBuffersMutex := &sync.Mutex{}

	arguments.DBClient = &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			split := strings.Split(buff.String(), "-")
			bufferNum, _ := strconv.Atoi(split[2])
			if bufferNum%3 == 0 {
				return expectedErr
			}

			processedBuffersMutex.Lock()
			processedBuffers[buff.String()]++
			processedBuffersMutex.Unlock()
			return nil
		},
	}

	elasticProc, _ := NewElasticOutput(arguments)
	buffers := make([]*bytes.Buffer, 6)
	for i := 0; i < 6; i++ {
		buffers[i] = bytes.NewBufferString(fmt.Sprintf("test-buffer-%d", i+1))
	}

	err := elasticProc.doBulkRequests("test-index", buffers, 0)
	require.NotNil(t, err)
	require.Equal(t, "database connection error\ndatabase connection error", err.Error())

	processedBuffersMutex.Lock()
	defer processedBuffersMutex.Unlock()
	require.True(t, len(processedBuffers) > 0)
	require.Equal(t, 1, processedBuffers["test-buffer-1"])
	require.Equal(t, 1, processedBuffers["test-buffer-2"])
	require.Equal(t, 0, processedBuffers["test-buffer-3"])
	require.Equal(t, 1, processedBuffers["test-buffer-4"])
	require.Equal(t, 1, processedBuffers["test-buffer-5"])
	require.Equal(t, 0, processedBuffers["test-buffer-6"])
}
