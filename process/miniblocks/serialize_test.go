package miniblocks

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func TestMiniblocksProcessor_SerializeBulkMiniBlocks(t *testing.T) {
	t.Parallel()

	mp, _ := NewMiniblocksProcessor(0, mock.HasherMock{}, &mock.MarshalizerMock{}, false)

	miniblocks := []*data.Miniblock{
		{Hash: "h1", SenderShardID: 0, ReceiverShardID: 1},
		{Hash: "h2", SenderShardID: 0, ReceiverShardID: 2},
	}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	mp.SerializeBulkMiniBlocks(miniblocks, nil, buffSlice, "miniblocks")

	expectedBuff := `{ "index" : { "_index":"miniblocks", "_id" : "h1"} }
{"senderShard":0,"receiverShard":1,"senderBlockHash":"","receiverBlockHash":"","type":"","timestamp":0}
{ "index" : { "_index":"miniblocks", "_id" : "h2"} }
{"senderShard":0,"receiverShard":2,"senderBlockHash":"","receiverBlockHash":"","type":"","timestamp":0}
`
	require.Equal(t, expectedBuff, buffSlice.Buffers()[0].String())
}

func TestMiniblocksProcessor_SerializeBulkMiniBlocksInDB(t *testing.T) {
	t.Parallel()

	mp, _ := NewMiniblocksProcessor(0, mock.HasherMock{}, &mock.MarshalizerMock{}, false)

	miniblocks := []*data.Miniblock{
		{Hash: "h1", SenderShardID: 0, ReceiverShardID: 1},
		{Hash: "h2", SenderShardID: 0, ReceiverShardID: 2},
	}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	mp.SerializeBulkMiniBlocks(miniblocks, map[string]bool{
		"h1": true,
	}, buffSlice, "miniblocks")

	expectedBuff := `{ "update" : {"_index":"miniblocks", "_id" : "h1" } }
{ "doc" : { "senderBlockHash" : "", "procTypeS": "" } }
{ "index" : { "_index":"miniblocks", "_id" : "h2"} }
{"senderShard":0,"receiverShard":2,"senderBlockHash":"","receiverBlockHash":"","type":"","timestamp":0}
`
	require.Equal(t, expectedBuff, buffSlice.Buffers()[0].String())
}
