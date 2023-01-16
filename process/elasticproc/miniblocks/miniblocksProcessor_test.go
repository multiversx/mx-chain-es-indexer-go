package miniblocks

import (
	"testing"

	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
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

	miniblocks := mp.PrepareDBMiniblocks(header, body)
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
	}
	body := &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{
				SenderShardID:   0,
				ReceiverShardID: 1,
			},
			{
				SenderShardID:   0,
				ReceiverShardID: 1,
			},
		},
	}

	miniblocks := mp.PrepareDBMiniblocks(header, body)
	require.Len(t, miniblocks, 2)
	require.Equal(t, dataBlock.Scheduled.String(), miniblocks[1].ProcessingTypeOnSource)
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
		"c57392e53257b4861f5e406349a8deb89c6dbc2127564ee891a41a188edbf01a",
		"28fda294dc987e5099d75e53cd6f87a9a42b96d55242a634385b5d41175c0c21",
		"44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
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
		"3acf97c324f5e8cd1e2d87de862b3105a9f08262c7914be2e186ced2a1cf1124",
		"40a551b2ebc5e4b5a55e73d49ec056c72af6314606850c4d54dadfad3a7e23e5",
		"4a270e1ddac6b429c14c7ebccdcdd53e4f68aeebfc41552c775a7f5a5c35d06d",
	}
	miniblocksHashes := mp.GetMiniblocksHashesHexEncoded(header, body)
	require.Equal(t, expectedHashes, miniblocksHashes)
}
