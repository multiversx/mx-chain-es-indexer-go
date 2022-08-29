package block

import (
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/core"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
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

	dbBlock, err := bp.PrepareBlockForDB(
		[]byte("hash"),
		&dataBlock.Header{},
		[]uint64{0, 1, 2},
		&dataBlock.Body{
			MiniBlocks: dataBlock.MiniBlockSlice{
				{
					ReceiverShardID: 1,
				},
				{
					ReceiverShardID: 2,
				},
			},
		}, nil, outport.HeaderGasConsumption{}, 0)
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

	dbBlock, err := bp.PrepareBlockForDB([]byte("hash"),nil, nil, &dataBlock.Body{}, nil, outport.HeaderGasConsumption{}, 0)
	require.Equal(t, indexer.ErrNilHeaderHandler, err)
	require.Nil(t, dbBlock)
}

func TestBlockProcessor_PrepareBlockForDBNilBody(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	dbBlock, err := bp.PrepareBlockForDB([]byte("hash"),&dataBlock.MetaBlock{}, nil, nil, nil, outport.HeaderGasConsumption{}, 0)
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

	dbBlock, err := bp.PrepareBlockForDB([]byte("hash"), &dataBlock.MetaBlock{}, nil, &dataBlock.Body{}, nil, outport.HeaderGasConsumption{}, 0)
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

	dbBlock, err := bp.PrepareBlockForDB([]byte("hash"), &dataBlock.MetaBlock{}, nil, &dataBlock.Body{}, nil, outport.HeaderGasConsumption{}, 0)
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

	dbBlock, err := bp.PrepareBlockForDB([]byte("hash"), &dataBlock.MetaBlock{
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
	}, nil, &dataBlock.Body{}, nil, outport.HeaderGasConsumption{}, 0)
	require.Equal(t, nil, err)
	require.Equal(t, &data.Block{
		Nonce:                 0,
		Round:                 0,
		Epoch:                 0,
		Hash:                  "68617368",
		MiniBlocksHashes:      []string{},
		NotarizedBlocksHashes: nil,
		Proposer:              0,
		Validators:            nil,
		PubKeyBitmap:          "",
		Size:                  623,
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
