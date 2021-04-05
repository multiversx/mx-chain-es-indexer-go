package miniblocks

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func TestMiniblocksProcessor_SerializeBulkMiniBlocks(t *testing.T) {
	t.Parallel()

	mp, _ := NewMiniblocksProcessor(0, mock.HasherMock{}, &mock.MarshalizerMock{})

	miniblocks := []*data.Miniblock{
		{Hash: "h1", SenderShardID: 0, ReceiverShardID: 1},
		{Hash: "h2", SenderShardID: 0, ReceiverShardID: 2},
	}

	buff := mp.SerializeBulkMiniBlocks(miniblocks, nil)

	expectedBuff := `{ "index" : { "_id" : "h1", "_type" : "_doc" } }
{"senderShard":0,"receiverShard":1,"senderBlockHash":"","receiverBlockHash":"","type":"","timestamp":0}
{ "index" : { "_id" : "h2", "_type" : "_doc" } }
{"senderShard":0,"receiverShard":2,"senderBlockHash":"","receiverBlockHash":"","type":"","timestamp":0}
`
	require.Equal(t, expectedBuff, buff.String())
}

func TestMiniblocksProcessor_SerializeBulkMiniBlocksInDB(t *testing.T) {
	t.Parallel()

	mp, _ := NewMiniblocksProcessor(0, mock.HasherMock{}, &mock.MarshalizerMock{})

	miniblocks := []*data.Miniblock{
		{Hash: "h1", SenderShardID: 0, ReceiverShardID: 1},
		{Hash: "h2", SenderShardID: 0, ReceiverShardID: 2},
	}

	buff := mp.SerializeBulkMiniBlocks(miniblocks, map[string]bool{
		"h1": true,
	})

	expectedBuff := `{ "update" : { "_id" : "h1" } }
{ "doc" : { "senderBlockHash" : "" } }
{ "index" : { "_id" : "h2", "_type" : "_doc" } }
{"senderShard":0,"receiverShard":2,"senderBlockHash":"","receiverBlockHash":"","type":"","timestamp":0}
`
	require.Equal(t, expectedBuff, buff.String())
}
