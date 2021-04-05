package process

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	errorsGo "errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/errors"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elastic-indexer-go/process/accounts"
	"github.com/ElrondNetwork/elastic-indexer-go/process/block"
	"github.com/ElrondNetwork/elastic-indexer-go/process/miniblocks"
	"github.com/ElrondNetwork/elastic-indexer-go/process/statistics"
	"github.com/ElrondNetwork/elastic-indexer-go/process/transactions"
	"github.com/ElrondNetwork/elastic-indexer-go/process/validators"
	"github.com/ElrondNetwork/elrond-go/core"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/stretchr/testify/require"
)

func newTestElasticSearchDatabase(elasticsearchWriter DatabaseClientHandler, arguments *ArgElasticProcessor) *elasticProcessor {
	return &elasticProcessor{
		elasticClient:  elasticsearchWriter,
		enabledIndexes: arguments.EnabledIndexes,
		blockProc:      arguments.BlockProc,
		txsProc:        arguments.TxsProc,
		miniblocksProc: arguments.MiniblocksProc,
		accountsProc:   arguments.AccountsProc,
		validatorsProc: arguments.ValidatorsProc,
		statisticsProc: arguments.StatisticsProc,
	}
}

func createMockElasticProcessorArgs() *ArgElasticProcessor {
	acp, _ := accounts.NewAccountsProcessor(0, &mock.MarshalizerMock{}, &mock.PubkeyConverterMock{}, &mock.AccountsStub{})
	bp, _ := block.NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})
	mp, _ := miniblocks.NewMiniblocksProcessor(0, &mock.HasherMock{}, &mock.MarshalizerMock{})
	vp, _ := validators.NewValidatorsProcessor(mock.NewPubkeyConverterMock(32))

	return &ArgElasticProcessor{
		DBClient: &mock.DatabaseWriterStub{},
		EnabledIndexes: map[string]struct{}{
			blockIndex: {}, txIndex: {}, miniblocksIndex: {}, tpsIndex: {}, validatorsIndex: {}, roundIndex: {}, accountsIndex: {}, ratingIndex: {}, accountsHistoryIndex: {},
		},
		ValidatorsProc: vp,
		StatisticsProc: statistics.NewStatisticsProcessor(),
		TxsProc:        &mock.DBTransactionProcessorStub{},
		MiniblocksProc: mp,
		AccountsProc:   acp,
		BlockProc:      bp,
	}
}

func newTestTxPool() map[string]nodeData.TransactionHandler {
	txPool := map[string]nodeData.TransactionHandler{
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

	expectedErr := errorsGo.New("local error")
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
			exErr: errors.ErrNilElasticProcessorArguments,
		},
		{
			name: "NilEnabledIndexesMap",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.EnabledIndexes = nil
				return arguments
			},
			exErr: errors.ErrNilEnabledIndexesMap,
		},
		{
			name: "NilDatabaseClient",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.DBClient = nil
				return arguments
			},
			exErr: errors.ErrNilDatabaseClient,
		},
		{
			name: "NilStatisticProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.StatisticsProc = nil
				return arguments
			},
			exErr: errors.ErrNilStatisticHandler,
		},
		{
			name: "NilBlockProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.BlockProc = nil
				return arguments
			},
			exErr: errors.ErrNilBlockHandler,
		},
		{
			name: "NilAccountsProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.AccountsProc = nil
				return arguments
			},
			exErr: errors.ErrNilAccountsHandler,
		},
		{
			name: "NilMiniblocksProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.MiniblocksProc = nil
				return arguments
			},
			exErr: errors.ErrNilMiniblocksHandler,
		},
		{
			name: "NilValidatorsProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.ValidatorsProc = nil
				return arguments
			},
			exErr: errors.ErrNilValidatorsHandler,
		},
		{
			name: "NilTxsProc",
			args: func() *ArgElasticProcessor {
				arguments := createMockElasticProcessorArgs()
				arguments.TxsProc = nil
				return arguments
			},
			exErr: errors.ErrNilTransactionsHandler,
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
			require.True(t, errorsGo.Is(err, tt.exErr))
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

	args.MiniblocksProc, _ = miniblocks.NewMiniblocksProcessor(0, &mock.HasherMock{}, &mock.MarshalizerMock{})

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
	localErr := errorsGo.New("localErr")
	header := &dataBlock.Header{Nonce: 1}
	signerIndexes := []uint64{0, 1}
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req *esapi.IndexRequest) error {
			return localErr
		},
	}
	arguments.BlockProc, _ = block.NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)

	err := elasticDatabase.SaveHeader(header, signerIndexes, &dataBlock.Body{}, nil, 1)
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
			require.Equal(t, blockIndex, req.Index)

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
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	err := elasticDatabase.SaveHeader(header, signerIndexes, blockBody, nil, 1)
	require.Nil(t, err)
}

func TestElasticseachSaveTransactions(t *testing.T) {
	localErr := errorsGo.New("localErr")
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
	arguments.TxsProc = txDbProc

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	pool := &indexer.Pool{Txs: txPool}
	err := elasticDatabase.SaveTransactions(body, header, pool, map[string]bool{})
	require.Equal(t, localErr, err)
}

func TestElasticProcessor_SaveValidatorsRating(t *testing.T) {
	docID := "0_1"
	localErr := errorsGo.New("localErr")

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
	localErr := errorsGo.New("localErr")

	arguments := createMockElasticProcessorArgs()
	arguments.DBClient = &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localErr
		},
		DoMultiGetCalled: func(hashes []string, index string) (map[string]interface{}, error) {
			return nil, nil
		},
	}

	arguments.MiniblocksProc, _ = miniblocks.NewMiniblocksProcessor(0, &mock.HasherMock{}, &mock.MarshalizerMock{})
	elasticProc, _ := NewElasticProcessor(arguments)

	header := &dataBlock.Header{}
	body := &dataBlock.Body{MiniBlocks: dataBlock.MiniBlockSlice{
		{SenderShardID: 0, ReceiverShardID: 1},
	}}
	mbsInDB, err := elasticProc.SaveMiniblocks(header, body)
	require.Equal(t, localErr, err)
	require.Equal(t, 0, len(mbsInDB))
}

func TestElasticsearch_saveShardValidatorsPubKeys_RequestError(t *testing.T) {
	shardID := uint32(0)
	epoch := uint32(0)
	valPubKeys := [][]byte{[]byte("key1"), []byte("key2")}
	localErr := errorsGo.New("localErr")
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req *esapi.IndexRequest) error {
			return localErr
		},
	}
	arguments.ValidatorsProc, _ = validators.NewValidatorsProcessor(mock.NewPubkeyConverterMock(32))
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)

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
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)

	err := elasticDatabase.SaveShardValidatorsPubKeys(shardID, epoch, valPubKeys)
	require.Nil(t, err)
}

func TestElasticsearch_saveShardStatistics_reqError(t *testing.T) {
	tpsBenchmark := &testscommon.TpsBenchmarkMock{}
	metaBlock := &dataBlock.MetaBlock{
		TxCount: 2, Nonce: 1,
		ShardInfo: []dataBlock.ShardData{{HeaderHash: []byte("hash")}},
	}
	tpsBenchmark.UpdateWithShardStats(metaBlock)

	localError := errorsGo.New("local err")
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localError
		},
	}

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)

	err := elasticDatabase.SaveShardStatistics(tpsBenchmark)
	require.Equal(t, localError, err)
}

func TestElasticsearch_saveShardStatistics(t *testing.T) {
	tpsBenchmark := &testscommon.TpsBenchmarkMock{}
	metaBlock := &dataBlock.MetaBlock{
		TxCount: 2, Nonce: 1,
		ShardInfo: []dataBlock.ShardData{{HeaderHash: []byte("hash")}},
	}
	tpsBenchmark.UpdateWithShardStats(metaBlock)

	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			require.Equal(t, tpsIndex, index)
			return nil
		},
	}
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)

	err := elasticDatabase.SaveShardStatistics(tpsBenchmark)
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
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)

	err := elasticDatabase.SaveRoundsInfo([]*data.RoundInfo{roundInfo})
	require.Nil(t, err)
}

func TestElasticsearch_saveRoundInfoRequestError(t *testing.T) {
	roundInfo := &data.RoundInfo{}
	localError := errorsGo.New("local err")
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localError
		},
	}
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)

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
			require.Equal(t, txIndex, index)
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

	arguments.TxsProc = txDbProc

	elasticSearchProc := newTestElasticSearchDatabase(dbWriter, arguments)

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
