package elasticproc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer"
	"github.com/ElrondNetwork/elastic-indexer-go/process/elasticproc/accounts"
	"github.com/ElrondNetwork/elastic-indexer-go/process/elasticproc/block"
	"github.com/ElrondNetwork/elastic-indexer-go/process/elasticproc/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/process/elasticproc/logsevents"
	"github.com/ElrondNetwork/elastic-indexer-go/process/elasticproc/miniblocks"
	"github.com/ElrondNetwork/elastic-indexer-go/process/elasticproc/operations"
	"github.com/ElrondNetwork/elastic-indexer-go/process/elasticproc/statistics"
	"github.com/ElrondNetwork/elastic-indexer-go/process/elasticproc/tags"
	"github.com/ElrondNetwork/elastic-indexer-go/process/elasticproc/transactions"
	"github.com/ElrondNetwork/elastic-indexer-go/process/elasticproc/validators"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/stretchr/testify/require"
)

func newElasticsearchProcessor(elasticsearchWriter DatabaseClientHandler, arguments *ArgElasticProcessor) *elasticProcessor {
	return &elasticProcessor{
		elasticClient:     elasticsearchWriter,
		enabledIndexes:    arguments.EnabledIndexes,
		blockProc:         arguments.BlockProc,
		transactionsProc:  arguments.TransactionsProc,
		miniblocksProc:    arguments.MiniblocksProc,
		accountsProc:      arguments.AccountsProc,
		validatorsProc:    arguments.ValidatorsProc,
		statisticsProc:    arguments.StatisticsProc,
		logsAndEventsProc: arguments.LogsAndEventsProc,
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
		ValidatorsProc:    vp,
		StatisticsProc:    statistics.NewStatisticsProcessor(),
		TransactionsProc:  &mock.DBTransactionProcessorStub{},
		MiniblocksProc:    mp,
		AccountsProc:      acp,
		BlockProc:         bp,
		LogsAndEventsProc: lp,
		OperationsProc:    op,
	}
}

func newTestTxPool() map[string]coreData.TransactionHandlerWithGasUsedAndFee {
	txPool := map[string]coreData.TransactionHandlerWithGasUsedAndFee{
		"tx1": outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
			Nonce:     uint64(1),
			Value:     big.NewInt(1),
			RcvAddr:   []byte("receiver_address1"),
			SndAddr:   []byte("sender_address1"),
			GasPrice:  uint64(10000),
			GasLimit:  uint64(1000),
			Data:      []byte("tx_data1"),
			Signature: []byte("signature1"),
		}, 0, big.NewInt(0)),
		"tx2": outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
			Nonce:     uint64(2),
			Value:     big.NewInt(2),
			RcvAddr:   []byte("receiver_address2"),
			SndAddr:   []byte("sender_address2"),
			GasPrice:  uint64(10000),
			GasLimit:  uint64(1000),
			Data:      []byte("tx_data2"),
			Signature: []byte("signature2"),
		}, 0, big.NewInt(0)),
		"tx3": outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
			Nonce:     uint64(3),
			Value:     big.NewInt(3),
			RcvAddr:   []byte("receiver_address3"),
			SndAddr:   []byte("sender_address3"),
			GasPrice:  uint64(10000),
			GasLimit:  uint64(1000),
			Data:      []byte("tx_data3"),
			Signature: []byte("signature3"),
		}, 0, big.NewInt(0)),
	}

	return txPool
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
	header := &dataBlock.Header{Nonce: 1}
	signerIndexes := []uint64{0, 1}
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localErr
		},
	}
	arguments.BlockProc, _ = block.NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})
	elasticDatabase := newElasticsearchProcessor(dbWriter, arguments)

	err := elasticDatabase.SaveHeader([]byte("hh"), header, signerIndexes, &dataBlock.Body{}, nil, outport.HeaderGasConsumption{}, 1)
	require.Equal(t, localErr, err)
}

func TestElasticseachDatabaseSaveHeader_CheckRequestBody(t *testing.T) {
	header := &dataBlock.Header{
		Nonce: 1,
	}
	signerIndexes := []uint64{0, 1}

	miniBlock := &dataBlock.MiniBlock{
		Type: dataBlock.TxBlock,
	}
	blockBody := &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			miniBlock,
		},
	}

	arguments := createMockElasticProcessorArgs()

	mbHash, _ := core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, miniBlock)
	hexEncodedHash := hex.EncodeToString(mbHash)

	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req *esapi.IndexRequest) error {
			require.Equal(t, dataindexer.BlockIndex, req.Index)

			var bl data.Block
			blockBytes, _ := ioutil.ReadAll(req.Body)
			_ = json.Unmarshal(blockBytes, &bl)
			require.Equal(t, header.Nonce, bl.Nonce)
			require.Equal(t, hexEncodedHash, bl.MiniBlocksHashes[0])
			require.Equal(t, signerIndexes, bl.Validators)

			return nil
		},
	}

	arguments.BlockProc, _ = block.NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})
	elasticDatabase := newElasticsearchProcessor(dbWriter, arguments)
	err := elasticDatabase.SaveHeader([]byte("hh"), header, signerIndexes, blockBody, nil, outport.HeaderGasConsumption{}, 1)
	require.Nil(t, err)
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
	txPool := newTestTxPool()

	args := &transactions.ArgsTransactionProcessor{
		AddressPubkeyConverter: mock.NewPubkeyConverterMock(32),
		Hasher:                 &mock.HasherMock{},
		Marshalizer:            &mock.MarshalizerMock{},
	}
	txDbProc, _ := transactions.NewTransactionsProcessor(args)
	arguments.TransactionsProc = txDbProc

	elasticDatabase := newElasticsearchProcessor(dbWriter, arguments)
	pool := &outport.Pool{Txs: txPool}
	err := elasticDatabase.SaveTransactions(body, header, pool, nil, false, 3)
	require.Equal(t, localErr, err)
}

func TestElasticProcessor_SaveValidatorsRating(t *testing.T) {
	docID := "0_1"
	localErr := errors.New("localErr")

	blsKey := "bls"

	arguments := createMockElasticProcessorArgs()
	arguments.DBClient = &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localErr
		},
	}

	arguments.ValidatorsProc, _ = validators.NewValidatorsProcessor(mock.NewPubkeyConverterMock(32), 0)
	elasticProc, _ := NewElasticProcessor(arguments)

	err := elasticProc.SaveValidatorsRating(
		docID,
		[]*data.ValidatorRatingInfo{
			{
				PublicKey: blsKey,
				Rating:    100,
			},
		},
	)
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
	err := elasticProc.SaveMiniblocks(header, body)
	require.Equal(t, localErr, err)
}

func TestElasticsearch_saveShardValidatorsPubKeys_RequestError(t *testing.T) {
	shardID := uint32(0)
	epoch := uint32(0)
	valPubKeys := [][]byte{[]byte("key1"), []byte("key2")}
	localErr := errors.New("localErr")
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req *esapi.IndexRequest) error {
			return localErr
		},
	}
	arguments.ValidatorsProc, _ = validators.NewValidatorsProcessor(mock.NewPubkeyConverterMock(32), 0)
	elasticDatabase := newElasticsearchProcessor(dbWriter, arguments)

	err := elasticDatabase.SaveShardValidatorsPubKeys(shardID, epoch, valPubKeys)
	require.Equal(t, localErr, err)
}

func TestElasticsearch_saveShardValidatorsPubKeys(t *testing.T) {
	shardID := uint32(0)
	epoch := uint32(0)
	valPubKeys := [][]byte{[]byte("key1"), []byte("key2")}
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req *esapi.IndexRequest) error {
			require.Equal(t, fmt.Sprintf("%d_%d", shardID, epoch), req.DocumentID)
			return nil
		},
	}
	elasticDatabase := newElasticsearchProcessor(dbWriter, arguments)

	err := elasticDatabase.SaveShardValidatorsPubKeys(shardID, epoch, valPubKeys)
	require.Nil(t, err)
}

func TestElasticsearch_saveRoundInfo(t *testing.T) {
	roundInfo := &data.RoundInfo{
		Index: 1, ShardId: 0, BlockWasProposed: true,
	}
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req *esapi.IndexRequest) error {
			require.Equal(t, strconv.FormatUint(uint64(roundInfo.ShardId), 10)+"_"+strconv.FormatUint(roundInfo.Index, 10), req.DocumentID)
			return nil
		},
	}
	elasticDatabase := newElasticsearchProcessor(dbWriter, arguments)

	err := elasticDatabase.SaveRoundsInfo([]*data.RoundInfo{roundInfo})
	require.Nil(t, err)
}

func TestElasticsearch_saveRoundInfoRequestError(t *testing.T) {
	roundInfo := &data.RoundInfo{}
	localError := errors.New("local err")
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localError
		},
	}
	elasticDatabase := newElasticsearchProcessor(dbWriter, arguments)

	err := elasticDatabase.SaveRoundsInfo([]*data.RoundInfo{roundInfo})
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
			require.Contains(t, []string{dataindexer.TransactionsIndex, dataindexer.OperationsIndex}, index)
			require.True(t, strings.Contains(bodyStr, expectedHashes[0]))
			require.True(t, strings.Contains(bodyStr, expectedHashes[1]))
			called = true
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

	body := &dataBlock.Body{}
	metaHeader := &dataBlock.MetaBlock{}

	err = elasticSearchProc.SaveHeader([]byte("hh"), metaHeader, nil, body, nil, outport.HeaderGasConsumption{}, 0)
	require.Nil(t, err)
	require.True(t, called)
}

func TestElasticProcessor_SaveTransactionNoDataShouldNotDoRequest(t *testing.T) {
	called := false
	arguments := createMockElasticProcessorArgs()
	arguments.TransactionsProc = &mock.DBTransactionProcessorStub{
		PrepareTransactionsForDatabaseCalled: func(body *dataBlock.Body, header coreData.HeaderHandler, pool *outport.Pool) *data.PreparedResults {
			return &data.PreparedResults{
				Transactions: nil,
				ScResults:    nil,
				Receipts:     nil,
				AlteredAccts: nil,
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

	err := elasticSearchProc.SaveTransactions(&dataBlock.Body{}, &dataBlock.Header{}, &outport.Pool{}, nil, false, 3)
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
	alteredAccounts := data.NewAlteredAccounts()
	tagsCount := tags.NewTagsCount()
	err := elasticSearchProc.indexAlteredAccounts(100, alteredAccounts, nil, nil, buffSlice, tagsCount, 0)
	require.Nil(t, err)
	require.True(t, called)
}
