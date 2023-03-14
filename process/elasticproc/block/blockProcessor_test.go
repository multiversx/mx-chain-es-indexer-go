package block

import (
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	indexer "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestNewBlockProcessor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		argsFunc func() (hashing.Hasher, marshal.Marshalizer)
		exErr    error
	}{
		{
			name: "NilMarshalizer",
			argsFunc: func() (hashing.Hasher, marshal.Marshalizer) {
				return &mock.HasherMock{}, nil
			},
			exErr: indexer.ErrNilMarshalizer,
		},
		{
			name: "NilHasher",
			argsFunc: func() (hashing.Hasher, marshal.Marshalizer) {
				return nil, &mock.MarshalizerMock{}
			},
			exErr: indexer.ErrNilHasher,
		},
		{
			name: "ShouldWork",
			argsFunc: func() (hashing.Hasher, marshal.Marshalizer) {
				return &mock.HasherMock{}, &mock.MarshalizerMock{}
			},
			exErr: nil,
		},
	}

	for _, tt := range tests {
		_, err := NewBlockProcessor(tt.argsFunc())
		require.Equal(t, tt.exErr, err)
	}
}

func TestBlockProcessor_PrepareBlockForDBShouldWork(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	outportBlockWithHeader := &outport.OutportBlockWithHeader{
		Header: &dataBlock.Header{},
		OutportBlock: outport.OutportBlock{
			BlockData: &outport.BlockData{
				HeaderHash: []byte("hash"),
				Body: &dataBlock.Body{
					MiniBlocks: dataBlock.MiniBlockSlice{
						{
							ReceiverShardID: 1,
						},
						{
							ReceiverShardID: 2,
						},
					},
				},
			},
			SignersIndexes:       []uint64{0, 1, 2},
			TransactionPool:      &outport.TransactionPool{},
			HeaderGasConsumption: &outport.HeaderGasConsumption{},
		},
	}
	dbBlock, err := bp.PrepareBlockForDB(outportBlockWithHeader)
	require.Nil(t, err)

	expectedBlock := &data.Block{
		Hash:                  "68617368",
		Validators:            []uint64{0x0, 0x1, 0x2},
		EpochStartBlock:       false,
		SearchOrder:           0x3fc,
		MiniBlocksHashes:      []string{"c57392e53257b4861f5e406349a8deb89c6dbc2127564ee891a41a188edbf01a", "28fda294dc987e5099d75e53cd6f87a9a42b96d55242a634385b5d41175c0c21"},
		NotarizedBlocksHashes: []string(nil),
		Size:                  104,
		AccumulatedFees:       "0",
		DeveloperFees:         "0",
	}
	require.Equal(t, expectedBlock, dbBlock)
}

func TestBlockProcessor_PrepareBlockForDBNilHeader(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	outportBlockWithHeader := &outport.OutportBlockWithHeader{}
	dbBlock, err := bp.PrepareBlockForDB(outportBlockWithHeader)
	require.Equal(t, indexer.ErrNilHeaderHandler, err)
	require.Nil(t, dbBlock)
}

func TestBlockProcessor_PrepareBlockForDBNilBody(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	outportBlockWithHeader := &outport.OutportBlockWithHeader{
		Header: &dataBlock.MetaBlock{},
	}
	dbBlock, err := bp.PrepareBlockForDB(outportBlockWithHeader)
	require.Equal(t, indexer.ErrNilBlockBody, err)
	require.Nil(t, dbBlock)
}

func TestBlockProcessor_PrepareBlockForDBMarshalFailHeader(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local error")
	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return nil, expectedErr
		},
	})

	outportBlockWithHeader := &outport.OutportBlockWithHeader{
		Header: &dataBlock.Header{},
		OutportBlock: outport.OutportBlock{
			BlockData: &outport.BlockData{
				HeaderHash: []byte("hash"),
				Body:       &dataBlock.Body{},
			},
			SignersIndexes:       []uint64{0, 1, 2},
			TransactionPool:      &outport.TransactionPool{},
			HeaderGasConsumption: &outport.HeaderGasConsumption{},
		},
	}
	dbBlock, err := bp.PrepareBlockForDB(outportBlockWithHeader)
	require.Equal(t, expectedErr, err)
	require.Nil(t, dbBlock)
}

func TestBlockProcessor_PrepareBlockForDBMarshalFailBlock(t *testing.T) {
	t.Parallel()

	count := 0
	expectedErr := errors.New("local error")
	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			defer func() {
				count++
			}()
			if count > 0 {
				return nil, expectedErr
			}
			return nil, nil
		},
	})

	outportBlockWithHeader := &outport.OutportBlockWithHeader{
		Header: &dataBlock.Header{},
		OutportBlock: outport.OutportBlock{
			BlockData: &outport.BlockData{
				HeaderHash: []byte("hash"),
				Body:       &dataBlock.Body{},
			},
			SignersIndexes:       []uint64{0, 1, 2},
			TransactionPool:      &outport.TransactionPool{},
			HeaderGasConsumption: &outport.HeaderGasConsumption{},
		},
	}
	dbBlock, err := bp.PrepareBlockForDB(outportBlockWithHeader)
	require.Equal(t, expectedErr, err)
	require.Nil(t, dbBlock)
}

func TestBlockProcessor_ComputeHeaderHash(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	header := &dataBlock.Header{}
	hashBytes, err := bp.ComputeHeaderHash(header)
	require.Nil(t, err)
	require.Equal(t, "c7c81a1b22b67680f35837b474387ddfe10f67e104034c80f94ab9e5a0a089fb", hex.EncodeToString(hashBytes))
}

func TestBlockProcessor_PrepareBlockForDBEpochStartMeta(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	outportBlockWithHeader := &outport.OutportBlockWithHeader{
		Header: &dataBlock.MetaBlock{
			TxCount: 1000,
			EpochStart: dataBlock.EpochStart{
				LastFinalizedHeaders: []dataBlock.EpochStartShardData{{
					ShardID:               1,
					Nonce:                 1234,
					Round:                 1500,
					Epoch:                 10,
					HeaderHash:            []byte("hh"),
					RootHash:              []byte("rh"),
					ScheduledRootHash:     []byte("sch"),
					FirstPendingMetaBlock: []byte("fpmb"),
					LastFinishedMetaBlock: []byte("lfmb"),
					PendingMiniBlockHeaders: []dataBlock.MiniBlockHeader{
						{
							Hash:            []byte("mbh"),
							SenderShardID:   0,
							ReceiverShardID: 1,
							Type:            dataBlock.TxBlock,
							Reserved:        []byte("rrr"),
						},
					},
				}},
				Economics: dataBlock.Economics{
					TotalSupply:                      big.NewInt(100),
					TotalToDistribute:                big.NewInt(55),
					TotalNewlyMinted:                 big.NewInt(20),
					RewardsPerBlock:                  big.NewInt(15),
					RewardsForProtocolSustainability: big.NewInt(2),
					NodePrice:                        big.NewInt(10),
					PrevEpochStartRound:              222,
					PrevEpochStartHash:               []byte("prevEpoch"),
				},
			},
			MiniBlockHeaders: []dataBlock.MiniBlockHeader{
				{
					TxCount: 50,
				},
				{
					TxCount: 120,
				},
			},
		},
		OutportBlock: outport.OutportBlock{
			BlockData: &outport.BlockData{
				HeaderHash: []byte("hash"),
				Body: &dataBlock.Body{
					MiniBlocks: []*dataBlock.MiniBlock{
						{},
						{},
					},
				},
			},
			TransactionPool:      &outport.TransactionPool{},
			HeaderGasConsumption: &outport.HeaderGasConsumption{},
		},
	}

	dbBlock, err := bp.PrepareBlockForDB(outportBlockWithHeader)
	require.Equal(t, nil, err)
	require.Equal(t, &data.Block{
		Nonce:                 0,
		Round:                 0,
		Epoch:                 0,
		Hash:                  "68617368",
		MiniBlocksHashes:      []string{"44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a", "44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"},
		NotarizedBlocksHashes: nil,
		Proposer:              0,
		Validators:            nil,
		PubKeyBitmap:          "",
		Size:                  643,
		SizeTxs:               0,
		Timestamp:             0,
		StateRootHash:         "",
		PrevHash:              "",
		ShardID:               core.MetachainShardId,
		EpochStartBlock:       true,
		SearchOrder:           0x3f2,
		EpochStartInfo: &data.EpochStartInfo{
			TotalSupply:                      "100",
			TotalToDistribute:                "55",
			TotalNewlyMinted:                 "20",
			RewardsPerBlock:                  "15",
			RewardsForProtocolSustainability: "2",
			NodePrice:                        "10",
			PrevEpochStartRound:              222,
			PrevEpochStartHash:               "7072657645706f6368",
		},
		MiniBlocksDetails: []*data.MiniBlocksDetails{
			{
				IndexFirstProcessedTx:    0,
				IndexLastProcessedTx:     49,
				MBIndex:                  0,
				Type:                     dataBlock.TxBlock.String(),
				ProcessingType:           dataBlock.Normal.String(),
				ExecutionOrderTxsIndices: []int{},
				TxsHashes:                []string{},
			},
			{
				IndexFirstProcessedTx:    0,
				IndexLastProcessedTx:     119,
				MBIndex:                  1,
				Type:                     dataBlock.TxBlock.String(),
				ProcessingType:           dataBlock.Normal.String(),
				ExecutionOrderTxsIndices: []int{},
				TxsHashes:                []string{},
			},
		},
		EpochStartShardsData: []*data.EpochStartShardData{
			{
				ShardID:               1,
				Epoch:                 10,
				Round:                 1500,
				Nonce:                 1234,
				HeaderHash:            "6868",
				RootHash:              "7268",
				ScheduledRootHash:     "736368",
				FirstPendingMetaBlock: "66706d62",
				LastFinishedMetaBlock: "6c666d62",
				PendingMiniBlockHeaders: []*data.Miniblock{
					{
						Hash:            "6d6268",
						SenderShardID:   0,
						ReceiverShardID: 1,
						Type:            "TxBlock",
						Reserved:        []byte("rrr"),
					},
				},
			},
		},
		NotarizedTxsCount: 830,
		TxCount:           170,
		AccumulatedFees:   "0",
		DeveloperFees:     "0",
	}, dbBlock)
}

func TestBlockProcessor_PrepareBlockForDBMiniBlocksDetails(t *testing.T) {
	t.Parallel()

	gogoMarshaller := &marshal.GogoProtoMarshalizer{}
	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	mbhr := &dataBlock.MiniBlockHeaderReserved{
		IndexOfFirstTxProcessed: 0,
		IndexOfLastTxProcessed:  1,
	}
	mbhrBytes, _ := gogoMarshaller.Marshal(mbhr)

	txHash, notExecutedTxHash, notFoundTxHash, invalidTxHash, rewardsTxHash, scrHash := "tx", "notExecuted", "notFound", "invalid", "reward", "scr"

	outportBlockWithHeader := &outport.OutportBlockWithHeader{
		Header: &dataBlock.Header{
			TxCount: 5,
			MiniBlockHeaders: []dataBlock.MiniBlockHeader{
				{
					TxCount:  1,
					Type:     dataBlock.TxBlock,
					Reserved: mbhrBytes,
				},
				{
					TxCount: 1,
					Type:    dataBlock.RewardsBlock,
				},
				{
					TxCount: 1,
					Type:    dataBlock.InvalidBlock,
				},
				{
					TxCount: 1,
					Type:    dataBlock.SmartContractResultBlock,
				},
			},
		},
		OutportBlock: outport.OutportBlock{
			BlockData: &outport.BlockData{
				HeaderHash: []byte("hash"),
				Body: &dataBlock.Body{
					MiniBlocks: []*dataBlock.MiniBlock{
						{
							Type:     dataBlock.TxBlock,
							TxHashes: [][]byte{[]byte(txHash), []byte(notFoundTxHash), []byte(notExecutedTxHash)},
						},
						{
							Type:     dataBlock.RewardsBlock,
							TxHashes: [][]byte{[]byte(rewardsTxHash)},
						},
						{
							Type:     dataBlock.InvalidBlock,
							TxHashes: [][]byte{[]byte(invalidTxHash)},
						},
						{
							Type:     dataBlock.SmartContractResultBlock,
							TxHashes: [][]byte{[]byte(scrHash)},
						},
					},
				},
			},
			SignersIndexes: []uint64{0, 1, 2},
			TransactionPool: &outport.TransactionPool{
				Transactions: map[string]*outport.TxInfo{
					txHash: {
						ExecutionOrder: 2,
					},
					notExecutedTxHash: {
						ExecutionOrder: 0,
					},
				},
				Rewards: map[string]*outport.RewardInfo{
					rewardsTxHash: {
						ExecutionOrder: 3,
					},
				},
				InvalidTxs: map[string]*outport.TxInfo{
					invalidTxHash: {
						ExecutionOrder: 1,
					}},
				SmartContractResults: map[string]*outport.SCRInfo{
					scrHash: {
						ExecutionOrder: 0,
					},
				},
			},
			HeaderGasConsumption: &outport.HeaderGasConsumption{},
		},
	}

	dbBlock, err := bp.PrepareBlockForDB(outportBlockWithHeader)
	require.Nil(t, err)

	require.Equal(t, &data.Block{
		Hash:            "68617368",
		Size:            int64(341),
		AccumulatedFees: "0",
		DeveloperFees:   "0",
		TxCount:         uint32(5),
		SearchOrder:     uint64(1020),
		MiniBlocksHashes: []string{
			"ee29d9b4a5017b7351974110d6a3f28ce6612476582f16b7849e3e87c647fc2d",
			"c067de5b3c0031a14578699b1c3cdb9a19039e4a7b3fae6a94932ad3f70cf375",
			"758f925b254ea0a6ad1bcbe3ddfcc73418ed4c8712506aafddc4da703295ad63",
			"28a96506c2999838923f5310b3bb1d6849b5a259b429790d9eeb21c2a1402f82",
		},
		MiniBlocksDetails: []*data.MiniBlocksDetails{
			{
				IndexFirstProcessedTx:    0,
				IndexLastProcessedTx:     1,
				MBIndex:                  0,
				Type:                     dataBlock.TxBlock.String(),
				ProcessingType:           dataBlock.Normal.String(),
				ExecutionOrderTxsIndices: []int{2, notFound, notExecutedInCurrentBlock},
				TxsHashes:                []string{"7478", "6e6f74466f756e64", "6e6f744578656375746564"},
			},
			{
				IndexFirstProcessedTx:    0,
				IndexLastProcessedTx:     0,
				MBIndex:                  1,
				Type:                     dataBlock.RewardsBlock.String(),
				ProcessingType:           dataBlock.Normal.String(),
				ExecutionOrderTxsIndices: []int{3},
				TxsHashes:                []string{"726577617264"},
			},
			{IndexFirstProcessedTx: 0,
				IndexLastProcessedTx:     0,
				MBIndex:                  2,
				Type:                     dataBlock.InvalidBlock.String(),
				ProcessingType:           dataBlock.Normal.String(),
				ExecutionOrderTxsIndices: []int{1},
				TxsHashes:                []string{"696e76616c6964"}},
			{IndexFirstProcessedTx: 0,
				IndexLastProcessedTx:     0,
				MBIndex:                  3,
				Type:                     dataBlock.SmartContractResultBlock.String(),
				ProcessingType:           dataBlock.Normal.String(),
				ExecutionOrderTxsIndices: []int{0},
				TxsHashes:                []string{"736372"}},
		},
	}, dbBlock)
}
