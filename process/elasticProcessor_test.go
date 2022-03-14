package process

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"
	"testing"

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elastic-indexer-go/process/accounts"
	"github.com/ElrondNetwork/elastic-indexer-go/process/block"
	"github.com/ElrondNetwork/elastic-indexer-go/process/indicesCreator"
	"github.com/ElrondNetwork/elastic-indexer-go/process/logsevents"
	"github.com/ElrondNetwork/elastic-indexer-go/process/miniblocks"
	"github.com/ElrondNetwork/elastic-indexer-go/process/operations"
	"github.com/ElrondNetwork/elastic-indexer-go/process/statistics"
	"github.com/ElrondNetwork/elastic-indexer-go/process/transactions"
	"github.com/ElrondNetwork/elastic-indexer-go/process/validators"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/stretchr/testify/require"
)

func newElasticsearchProcessor(elasticsearchWriter DatabaseClientRequestsHandler, arguments *ArgElasticProcessor) *elasticProcessor {
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

	acp, _ := accounts.NewAccountsProcessor(&mock.MarshalizerMock{}, &mock.PubkeyConverterMock{}, &mock.AccountsStub{}, balanceConverter)
	bp, _ := block.NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})
	mp, _ := miniblocks.NewMiniblocksProcessor(0, &mock.HasherMock{}, &mock.MarshalizerMock{}, false)
	vp, _ := validators.NewValidatorsProcessor(mock.NewPubkeyConverterMock(32))
	args := &logsevents.ArgsLogsAndEventsProcessor{
		ShardCoordinator: &mock.ShardCoordinatorMock{},
		PubKeyConverter:  &mock.PubkeyConverterMock{},
		Marshalizer:      &mock.MarshalizerMock{},
		BalanceConverter: balanceConverter,
		Hasher:           &mock.HasherMock{},
		TxFeeCalculator:  &mock.EconomicsHandlerStub{},
	}
	lp, _ := logsevents.NewLogsAndEventsProcessor(args)
	op, _ := operations.NewOperationsProcessor(false, &mock.ShardCoordinatorMock{})
	esClient := &mock.DatabaseWriterStub{}
	ic, _ := indicesCreator.NewIndicesCreator(esClient)

	return &ArgElasticProcessor{
		DBClient: esClient,
		EnabledIndexes: map[string]struct{}{
			data.BlockIndex: {}, data.TransactionsIndex: {}, data.MiniblocksIndex: {}, data.ValidatorsIndex: {}, data.RoundsIndex: {}, data.AccountsIndex: {}, data.RatingIndex: {}, data.AccountsHistoryIndex: {},
		},
		ValidatorsProc:    vp,
		StatisticsProc:    statistics.NewStatisticsProcessor(),
		TransactionsProc:  &mock.DBTransactionProcessorStub{},
		MiniblocksProc:    mp,
		AccountsProc:      acp,
		BlockProc:         bp,
		LogsAndEventsProc: lp,
		OperationsProc:    op,
		IndicesCreator:    ic,
	}
}

func newTestTxPool() map[string]coreData.TransactionHandler {
	txPool := map[string]coreData.TransactionHandler{
		"tx1": &transaction.Transaction{
			Nonce:     uint64(1),
			Value:     big.NewInt(1),
			RcvAddr:   []byte("receiver_address1"),
			SndAddr:   []byte("sender_address1"),
			GasPrice:  uint64(10000),
			GasLimit:  uint64(1000),
			Data:      []byte("tx_data1"),
			Signature: []byte("signature1"),
		},
		"tx2": &transaction.Transaction{
			Nonce:     uint64(2),
			Value:     big.NewInt(2),
			RcvAddr:   []byte("receiver_address2"),
			SndAddr:   []byte("sender_address2"),
			GasPrice:  uint64(10000),
			GasLimit:  uint64(1000),
			Data:      []byte("tx_data2"),
			Signature: []byte("signature2"),
		},
		"tx3": &transaction.Transaction{
			Nonce:     uint64(3),
			Value:     big.NewInt(3),
			RcvAddr:   []byte("receiver_address3"),
			SndAddr:   []byte("sender_address3"),
			GasPrice:  uint64(10000),
			GasLimit:  uint64(1000),
			Data:      []byte("tx_data3"),
			Signature: []byte("signature3"),
		},
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
			exErr: elasticIndexer.ErrNilElasticProcessorArguments,
		},
		{
			name: "NilEnabledIndexesMap",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.EnabledIndexes = nil
				return arguments
			},
			exErr: elasticIndexer.ErrNilEnabledIndexesMap,
		},
		{
			name: "NilDatabaseClient",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.DBClient = nil
				return arguments
			},
			exErr: elasticIndexer.ErrNilDatabaseClient,
		},
		{
			name: "NilStatisticProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.StatisticsProc = nil
				return arguments
			},
			exErr: elasticIndexer.ErrNilStatisticHandler,
		},
		{
			name: "NilBlockProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.BlockProc = nil
				return arguments
			},
			exErr: elasticIndexer.ErrNilBlockHandler,
		},
		{
			name: "NilAccountsProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.AccountsProc = nil
				return arguments
			},
			exErr: elasticIndexer.ErrNilAccountsHandler,
		},
		{
			name: "NilMiniblocksProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.MiniblocksProc = nil
				return arguments
			},
			exErr: elasticIndexer.ErrNilMiniblocksHandler,
		},
		{
			name: "NilValidatorsProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.ValidatorsProc = nil
				return arguments
			},
			exErr: elasticIndexer.ErrNilValidatorsHandler,
		},
		{
			name: "NilTxsProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.TransactionsProc = nil
				return arguments
			},
			exErr: elasticIndexer.ErrNilTransactionsHandler,
		},
		{
			name: "NilIndicesCreator",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.IndicesCreator = nil
				return arguments
			},
			exErr: elasticIndexer.ErrNilIndicesCreatorHandler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewElasticProcessor(tt.args())
			require.True(t, errors.Is(err, tt.exErr))
		})
	}
}

func TestElasticProcessor_RemoveHeader(t *testing.T) {
	called := false

	args := createMockElasticProcessorArgs()
	args.DBClient = &mock.DatabaseWriterStub{
		DoBulkRemoveCalled: func(index string, hashes []string) error {
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
		DoBulkRemoveCalled: func(index string, hashes []string) error {
			called = true
			require.Equal(t, hashes[0], hex.EncodeToString(mbHash2))
			require.Equal(t, hashes[1], hex.EncodeToString(mbHash3))
			return nil
		},
	}

	args.MiniblocksProc, _ = miniblocks.NewMiniblocksProcessor(0, &mock.HasherMock{}, &mock.MarshalizerMock{}, false)

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
		DoRequestCalled: func(req *esapi.IndexRequest) error {
			return localErr
		},
	}
	arguments.BlockProc, _ = block.NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})
	elasticDatabase := newElasticsearchProcessor(dbWriter, arguments)

	err := elasticDatabase.SaveHeader(header, signerIndexes, &dataBlock.Body{}, nil, indexer.HeaderGasConsumption{}, 1)
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
			require.Equal(t, data.BlockIndex, req.Index)

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
	err := elasticDatabase.SaveHeader(header, signerIndexes, blockBody, nil, indexer.HeaderGasConsumption{}, 1)
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
		AddressPubkeyConverter: &mock.PubkeyConverterMock{},
		TxFeeCalculator:        &mock.EconomicsHandlerStub{},
		ShardCoordinator:       &mock.ShardCoordinatorMock{},
		Hasher:                 &mock.HasherMock{},
		Marshalizer:            &mock.MarshalizerMock{},
		IsInImportMode:         false,
	}
	txDbProc, _ := transactions.NewTransactionsProcessor(args)
	arguments.TransactionsProc = txDbProc

	elasticDatabase := newElasticsearchProcessor(dbWriter, arguments)
	pool := &indexer.Pool{Txs: txPool}
	err := elasticDatabase.SaveTransactions(body, header, pool)
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

	arguments.ValidatorsProc, _ = validators.NewValidatorsProcessor(mock.NewPubkeyConverterMock(32))
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

	arguments.MiniblocksProc, _ = miniblocks.NewMiniblocksProcessor(0, &mock.HasherMock{}, &mock.MarshalizerMock{}, false)
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
	arguments.ValidatorsProc, _ = validators.NewValidatorsProcessor(mock.NewPubkeyConverterMock(32))
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
		DoBulkRemoveCalled: func(index string, hashes []string) error {
			require.Equal(t, data.TransactionsIndex, index)
			require.Equal(t, expectedHashes, expectedHashes)
			called = true
			return nil
		},
	}

	args := &transactions.ArgsTransactionProcessor{
		AddressPubkeyConverter: &mock.PubkeyConverterMock{},
		TxFeeCalculator:        &mock.EconomicsHandlerStub{},
		ShardCoordinator:       &mock.ShardCoordinatorMock{},
		Hasher:                 &mock.HasherMock{},
		Marshalizer:            &mock.MarshalizerMock{},
		IsInImportMode:         false,
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
		DoRequestCalled: func(req *esapi.IndexRequest) error {
			called = true
			return nil
		},
	}

	elasticSearchProc := newElasticsearchProcessor(dbWriter, arguments)
	elasticSearchProc.selfShardID = core.MetachainShardId
	elasticSearchProc.enabledIndexes[data.EpochInfoIndex] = struct{}{}

	shardHeader := &dataBlock.Header{}
	err := elasticSearchProc.indexEpochInfoData(shardHeader)
	require.True(t, errors.Is(err, elasticIndexer.ErrHeaderTypeAssertion))

	metaHeader := &dataBlock.MetaBlock{}

	err = elasticSearchProc.indexEpochInfoData(metaHeader)
	require.Nil(t, err)
	require.True(t, called)
}

func TestElasticProcessor_SaveTransactionsIndexSCResults(t *testing.T) {
	called := false
	arguments := createMockElasticProcessorArgs()
	arguments.TransactionsProc = &mock.DBTransactionProcessorStub{
		PrepareTransactionsForDatabaseCalled: func(body *dataBlock.Body, header coreData.HeaderHandler, pool *indexer.Pool) *data.PreparedResults {
			return &data.PreparedResults{
				Transactions: nil,
				ScResults:    nil,
				Receipts:     nil,
				AlteredAccts: nil,
			}
		},
		SerializeScResultsCalled: func(scrs []*data.ScResult) ([]*bytes.Buffer, error) {
			return []*bytes.Buffer{{}}, nil
		},
	}
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			if index == data.ScResultsIndex {
				called = true
			}
			return nil
		},
		DoRequestCalled: func(req *esapi.IndexRequest) error {
			return nil
		},
	}

	elasticSearchProc := newElasticsearchProcessor(dbWriter, arguments)
	elasticSearchProc.enabledIndexes[data.ScResultsIndex] = struct{}{}

	err := elasticSearchProc.SaveTransactions(&dataBlock.Body{}, &dataBlock.Header{}, &indexer.Pool{})
	require.Nil(t, err)
	require.True(t, called)
}

func TestElasticProcessor_SaveTransactionsIndexReceipts(t *testing.T) {
	called := false
	arguments := createMockElasticProcessorArgs()
	arguments.TransactionsProc = &mock.DBTransactionProcessorStub{
		PrepareTransactionsForDatabaseCalled: func(body *dataBlock.Body, header coreData.HeaderHandler, pool *indexer.Pool) *data.PreparedResults {
			return &data.PreparedResults{
				Transactions: nil,
				ScResults:    nil,
				Receipts:     nil,
				AlteredAccts: nil,
			}
		},
		SerializeReceiptsCalled: func(recs []*data.Receipt) ([]*bytes.Buffer, error) {
			return []*bytes.Buffer{{}}, nil
		},
	}
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			if index == data.ReceiptsIndex {
				called = true
			}
			return nil
		},
		DoRequestCalled: func(req *esapi.IndexRequest) error {
			return nil
		},
	}

	elasticSearchProc := newElasticsearchProcessor(dbWriter, arguments)
	elasticSearchProc.enabledIndexes[data.ReceiptsIndex] = struct{}{}

	err := elasticSearchProc.SaveTransactions(&dataBlock.Body{}, &dataBlock.Header{}, &indexer.Pool{})
	require.Nil(t, err)
	require.True(t, called)
}

func TestElasticProcessor_IndexAlteredAccounts(t *testing.T) {
	called := false
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			called = true
			return nil
		},
		DoMultiGetCalled: func(ids []string, index string, withSource bool, response interface{}) error {
			return nil
		},
	}
	arguments := createMockElasticProcessorArgs()
	arguments.AccountsProc = &mock.DBAccountsHandlerStub{
		SerializeAccountsHistoryCalled: func(accounts map[string]*data.AccountBalanceHistory) ([]*bytes.Buffer, error) {
			return []*bytes.Buffer{{}}, nil
		},
	}
	elasticSearchProc := newElasticsearchProcessor(dbWriter, arguments)
	elasticSearchProc.enabledIndexes[data.AccountsESDTIndex] = struct{}{}
	elasticSearchProc.enabledIndexes[data.AccountsESDTHistoryIndex] = struct{}{}

	alteredAccounts := data.NewAlteredAccounts()
	err := elasticSearchProc.indexAlteredAccounts(100, alteredAccounts, nil)
	require.Nil(t, err)
	require.True(t, called)
}
