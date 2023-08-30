package block

import (
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
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
		Header: &dataBlock.Header{
			RandSeed:        []byte("randSeed"),
			PrevRandSeed:    []byte("prevRandSeed"),
			Signature:       []byte("signature"),
			LeaderSignature: []byte("leaderSignature"),
			ChainID:         []byte("1"),
			SoftwareVersion: []byte("1"),
			ReceiptsHash:    []byte("hash"),
			Reserved:        []byte("reserved"),
		},
		OutportBlock: &outport.OutportBlock{
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
		MiniBlocksHashes:      []string{"0796d34e8d443fd31bf4d9ec4051421b4d5d0e8c1db9ff942d6f4dc3a9ca2803", "4cc379ab1f0aef6602e85a0a7ffabb5bc9a2ba646dc0fd720028e06527bf873f"},
		NotarizedBlocksHashes: []string(nil),
		Size:                  114,
		AccumulatedFees:       "0",
		DeveloperFees:         "0",
		RandSeed:              "72616e6453656564",
		PrevRandSeed:          "7072657652616e6453656564",
		Signature:             "7369676e6174757265",
		LeaderSignature:       "6c65616465725369676e6174757265",
		ChainID:               "1",
		SoftwareVersion:       "31",
		ReceiptsHash:          "68617368",
		Reserved:              []byte("reserved"),
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
		OutportBlock: &outport.OutportBlock{
			BlockData: &outport.BlockData{},
		},
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
		OutportBlock: &outport.OutportBlock{
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

	expectedErr := errors.New("local error")
	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return nil, expectedErr
		},
	})

	outportBlockWithHeader := &outport.OutportBlockWithHeader{
		Header: &dataBlock.Header{},
		OutportBlock: &outport.OutportBlock{
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
	require.Equal(t, "96f7d09988eafbc99b45dfce0eaf9df1d02def2ae678d88bd154ebffa3247b2a", hex.EncodeToString(hashBytes))
}

func TestBlockProcessor_PrepareBlockForDBEpochStartMeta(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	header := &dataBlock.MetaBlock{
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
	}

	headerBytes, _ := bp.marshalizer.Marshal(header)
	outportBlockWithHeader := &outport.OutportBlockWithHeader{
		Header: header,
		OutportBlock: &outport.OutportBlock{
			BlockData: &outport.BlockData{
				HeaderBytes: headerBytes,
				HeaderHash:  []byte("hash"),
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
		MiniBlocksHashes:      []string{"8748c4677b01f7db984004fa8465afbf55feaab4b573174c8c0afa282941b9e4", "8748c4677b01f7db984004fa8465afbf55feaab4b573174c8c0afa282941b9e4"},
		NotarizedBlocksHashes: nil,
		Proposer:              0,
		Validators:            nil,
		PubKeyBitmap:          "",
		Size:                  898,
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

	header := &dataBlock.Header{
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
	}
	headerBytes, _ := bp.marshalizer.Marshal(header)
	outportBlockWithHeader := &outport.OutportBlockWithHeader{
		Header: header,
		OutportBlock: &outport.OutportBlock{
			BlockData: &outport.BlockData{
				HeaderBytes: headerBytes,
				HeaderHash:  []byte("hash"),
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
			TransactionPool: &outport.TransactionPool{
				Transactions: map[string]*outport.TxInfo{
					hex.EncodeToString([]byte(txHash)): {
						Transaction:    &transaction.Transaction{},
						ExecutionOrder: 2,
					},
					hex.EncodeToString([]byte(notExecutedTxHash)): {
						ExecutionOrder: 0,
					},
				},
				Rewards: map[string]*outport.RewardInfo{
					hex.EncodeToString([]byte(rewardsTxHash)): {
						Reward:         &rewardTx.RewardTx{},
						ExecutionOrder: 3,
					},
				},
				InvalidTxs: map[string]*outport.TxInfo{
					hex.EncodeToString([]byte(invalidTxHash)): {
						Transaction:    &transaction.Transaction{},
						ExecutionOrder: 1,
					}},
				SmartContractResults: map[string]*outport.SCRInfo{
					hex.EncodeToString([]byte(scrHash)): {
						SmartContractResult: &smartContractResult.SmartContractResult{},
						ExecutionOrder:      0,
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
		Size:            int64(723),
		SizeTxs:         15,
		AccumulatedFees: "0",
		DeveloperFees:   "0",
		TxCount:         uint32(5),
		SearchOrder:     uint64(1020),
		MiniBlocksHashes: []string{
			"8987edec270eb942d8ea9051fe301673aea29890919f5849882617aabcc7a248",
			"1183f422a5b76c3cb7b439334f1fe7235c8d09f577e0f1e15e62cd05b9a81950",
			"b24e307f3917e84603d3ebfb9c03c8fc651b62cb68ca884c3ff015b66a610a79",
			"c0a855563172b2f72be569963d26d4fae38d4371342e2bf3ded93466a72f36f3",
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
