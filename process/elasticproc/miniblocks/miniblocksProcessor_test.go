package miniblocks

import (
	"testing"

	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestNewMiniblocksProcessor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		args  func() (hashing.Hasher, marshal.Marshalizer)
		exErr error
	}{
		{
			name: "NilHash",
			args: func() (hashing.Hasher, marshal.Marshalizer) {
				return nil, &mock.MarshalizerMock{}
			},
			exErr: dataindexer.ErrNilHasher,
		},
		{
			name: "NilMarshalizer",
			args: func() (hashing.Hasher, marshal.Marshalizer) {
				return &mock.HasherMock{}, nil
			},
			exErr: dataindexer.ErrNilMarshalizer,
		},
	}

	for _, test := range tests {
		_, err := NewMiniblocksProcessor(test.args())
		require.Equal(t, test.exErr, err)
	}
}

func TestMiniblocksProcessor_PrepareDBMiniblocks(t *testing.T) {
	t.Parallel()

	mp, _ := NewMiniblocksProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	header := &dataBlock.Header{}
	body := &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{
				SenderShardID:   0,
				ReceiverShardID: 1,
			},
			{
				SenderShardID:   1,
				ReceiverShardID: 0,
			},
			{
				SenderShardID:   0,
				ReceiverShardID: 0,
			},
		},
	}

	miniblocks := mp.PrepareDBMiniblocks(header, body.MiniBlocks, 1234000)
	require.Len(t, miniblocks, 3)
}

func TestMiniblocksProcessor_PrepareScheduledMB(t *testing.T) {
	t.Parallel()

	marshalizer := &marshal.GogoProtoMarshalizer{}
	mp, _ := NewMiniblocksProcessor(&mock.HasherMock{}, marshalizer)

	mbhr := &dataBlock.MiniBlockHeaderReserved{
		ExecutionType: dataBlock.ProcessingType(1),
	}

	mbhrBytes, _ := marshalizer.Marshal(mbhr)

	header := &dataBlock.Header{
		MiniBlockHeaders: []dataBlock.MiniBlockHeader{
			{
				Reserved: []byte{0},
			},
			{
				Reserved: mbhrBytes,
			},
		},
		TimeStamp: 1234,
	}
	miniBlocks := []*dataBlock.MiniBlock{
		{
			SenderShardID:   0,
			ReceiverShardID: 1,
		},
		{
			SenderShardID:   0,
			ReceiverShardID: 1,
		},
		{
			SenderShardID:   0,
			ReceiverShardID: 0,
		},
	}

	miniblocks := mp.PrepareDBMiniblocks(header, miniBlocks, 1234000)
	require.Len(t, miniblocks, 3)
	require.Equal(t, dataBlock.Scheduled.String(), miniblocks[1].ProcessingTypeOnSource)

	require.Equal(t, &data.Miniblock{
		Hash:                        "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		SenderShardID:               0,
		ReceiverShardID:             0,
		SenderBlockHash:             "64ad61aaddb68f8d0b38ceda3b2b1f76a6749a0e848ed9e95bdaff46b4e73423",
		ReceiverBlockHash:           "64ad61aaddb68f8d0b38ceda3b2b1f76a6749a0e848ed9e95bdaff46b4e73423",
		ProcessingTypeOnSource:      dataBlock.Normal.String(),
		ProcessingTypeOnDestination: dataBlock.Normal.String(),
		Type:                        dataBlock.TxBlock.String(),
		Timestamp:                   1234,
		TimestampMs:                 1234000,
	}, miniblocks[2])
}

func TestMiniblocksProcessor_GetMiniblocksHashesHexEncoded(t *testing.T) {
	t.Parallel()

	mp, _ := NewMiniblocksProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	header := &dataBlock.Header{
		MiniBlockHeaders: []dataBlock.MiniBlockHeader{
			{}, {}, {},
		},
	}
	body := &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{
				SenderShardID:   0,
				ReceiverShardID: 1,
			},
			{
				SenderShardID:   0,
				ReceiverShardID: 2,
			},
			{
				SenderShardID:   0,
				ReceiverShardID: 0,
			},
		},
	}

	expectedHashes := []string{
		"0796d34e8d443fd31bf4d9ec4051421b4d5d0e8c1db9ff942d6f4dc3a9ca2803",
		"4cc379ab1f0aef6602e85a0a7ffabb5bc9a2ba646dc0fd720028e06527bf873f",
		"8748c4677b01f7db984004fa8465afbf55feaab4b573174c8c0afa282941b9e4",
	}
	miniblocksHashes := mp.GetMiniblocksHashesHexEncoded(header, body)
	require.Equal(t, expectedHashes, miniblocksHashes)
}

func TestMiniblocksProcessor_GetMiniblocksHashesHexEncodedImportDBMode(t *testing.T) {
	t.Parallel()

	mp, _ := NewMiniblocksProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	header := &dataBlock.Header{
		MiniBlockHeaders: []dataBlock.MiniBlockHeader{
			{}, {}, {},
		},
		ShardID: 1,
	}
	body := &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{
				SenderShardID:   1,
				ReceiverShardID: 2,
			},
			{
				SenderShardID:   0,
				ReceiverShardID: 1,
			},
			{
				SenderShardID:   1,
				ReceiverShardID: 0,
			},
			{
				SenderShardID:   1,
				ReceiverShardID: 1,
			},
		},
	}

	expectedHashes := []string{
		"11a1bb4065e16a2e93b2b5ac5957b7b69f1cfba7579b170b24f30dab2d3162e0",
		"68e9a4360489ab7a6e99f92e05d1a3f06a982b7b157ac23fdf39f2392bf88e15",
		"d1fd2a5c95c8899ebbaad035b6b0f77c5103b3aacfe630b1a7c51468d682bb1b",
	}
	miniblocksHashes := mp.GetMiniblocksHashesHexEncoded(header, body)
	require.Equal(t, expectedHashes, miniblocksHashes)
}
