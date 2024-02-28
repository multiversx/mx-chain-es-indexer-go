package miniblocks

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func TestMiniblocksProcessor_SerializeBulkMiniBlocks(t *testing.T) {
	t.Parallel()

	mp, _ := NewMiniblocksProcessor(mock.HasherMock{}, &mock.MarshalizerMock{})

	miniblocks := []*data.Miniblock{
		{Hash: "h1", SenderShardID: 0, ReceiverShardID: 1},
		{Hash: "h2", SenderShardID: 0, ReceiverShardID: 2},
	}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	mp.SerializeBulkMiniBlocks(miniblocks, buffSlice, "miniblocks", 0)

	expectedBuff := `{ "update" : {"_index":"miniblocks", "_id" : "h1" } }
{"scripted_upsert": true, "script": {"source": "if ('create' == ctx.op) {ctx._source = params.mb} else {if (params.npos) {ctx._source.senderBlockHash = params.mb.senderBlockHash;ctx._source.senderBlockNonce = params.mb.senderBlockNonce;ctx._source.procTypeS = params.mb.procTypeS;} else {ctx._source.receiverBlockHash = params.mb.receiverBlockHash;ctx._source.receiverBlockNonce = params.mb.receiverBlockNonce;ctx._source.procTypeD = params.mb.procTypeD;}}","lang": "painless","params": { "mb": {"senderShard":0,"receiverShard":1,"type":"","timestamp":0}, "npos": true }},"upsert": {}}
{ "update" : {"_index":"miniblocks", "_id" : "h2" } }
{"scripted_upsert": true, "script": {"source": "if ('create' == ctx.op) {ctx._source = params.mb} else {if (params.npos) {ctx._source.senderBlockHash = params.mb.senderBlockHash;ctx._source.senderBlockNonce = params.mb.senderBlockNonce;ctx._source.procTypeS = params.mb.procTypeS;} else {ctx._source.receiverBlockHash = params.mb.receiverBlockHash;ctx._source.receiverBlockNonce = params.mb.receiverBlockNonce;ctx._source.procTypeD = params.mb.procTypeD;}}","lang": "painless","params": { "mb": {"senderShard":0,"receiverShard":2,"type":"","timestamp":0}, "npos": true }},"upsert": {}}
`
	require.Equal(t, expectedBuff, buffSlice.Buffers()[0].String())
}

func TestMiniblocksProcessor_SerializeBulkMiniBlocksInDB(t *testing.T) {
	t.Parallel()

	mp, _ := NewMiniblocksProcessor(mock.HasherMock{}, &mock.MarshalizerMock{})

	miniblocks := []*data.Miniblock{
		{Hash: "h1", SenderShardID: 0, ReceiverShardID: 1},
		{Hash: "h2", SenderShardID: 0, ReceiverShardID: 2},
	}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	mp.SerializeBulkMiniBlocks(miniblocks, buffSlice, "miniblocks", 0)

	expectedBuff := `{ "update" : {"_index":"miniblocks", "_id" : "h1" } }
{"scripted_upsert": true, "script": {"source": "if ('create' == ctx.op) {ctx._source = params.mb} else {if (params.npos) {ctx._source.senderBlockHash = params.mb.senderBlockHash;ctx._source.senderBlockNonce = params.mb.senderBlockNonce;ctx._source.procTypeS = params.mb.procTypeS;} else {ctx._source.receiverBlockHash = params.mb.receiverBlockHash;ctx._source.receiverBlockNonce = params.mb.receiverBlockNonce;ctx._source.procTypeD = params.mb.procTypeD;}}","lang": "painless","params": { "mb": {"senderShard":0,"receiverShard":1,"type":"","timestamp":0}, "npos": true }},"upsert": {}}
{ "update" : {"_index":"miniblocks", "_id" : "h2" } }
{"scripted_upsert": true, "script": {"source": "if ('create' == ctx.op) {ctx._source = params.mb} else {if (params.npos) {ctx._source.senderBlockHash = params.mb.senderBlockHash;ctx._source.senderBlockNonce = params.mb.senderBlockNonce;ctx._source.procTypeS = params.mb.procTypeS;} else {ctx._source.receiverBlockHash = params.mb.receiverBlockHash;ctx._source.receiverBlockNonce = params.mb.receiverBlockNonce;ctx._source.procTypeD = params.mb.procTypeD;}}","lang": "painless","params": { "mb": {"senderShard":0,"receiverShard":2,"type":"","timestamp":0}, "npos": true }},"upsert": {}}
`
	require.Equal(t, expectedBuff, buffSlice.Buffers()[0].String())
}

func TestSerializeMiniblock_CrossShardNormal(t *testing.T) {
	mp, _ := NewMiniblocksProcessor(mock.HasherMock{}, &mock.MarshalizerMock{})

	miniblocks := []*data.Miniblock{
		{Hash: "h1", SenderShardID: 0, ReceiverShardID: 1, ReceiverBlockHash: "receiverBlock"},
	}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	mp.SerializeBulkMiniBlocks(miniblocks, buffSlice, "miniblocks", 1)

	expectedBuff := `{ "update" : {"_index":"miniblocks", "_id" : "h1" } }
{"scripted_upsert": true, "script": {"source": "if ('create' == ctx.op) {ctx._source = params.mb} else {if (params.npos) {ctx._source.senderBlockHash = params.mb.senderBlockHash;ctx._source.senderBlockNonce = params.mb.senderBlockNonce;ctx._source.procTypeS = params.mb.procTypeS;} else {ctx._source.receiverBlockHash = params.mb.receiverBlockHash;ctx._source.receiverBlockNonce = params.mb.receiverBlockNonce;ctx._source.procTypeD = params.mb.procTypeD;}}","lang": "painless","params": { "mb": {"senderShard":0,"receiverShard":1,"receiverBlockHash":"receiverBlock","type":"","timestamp":0}, "npos": false }},"upsert": {}}
`
	require.Equal(t, expectedBuff, buffSlice.Buffers()[0].String())
}

func TestSerializeMiniblock_IntraShardScheduled(t *testing.T) {
	mp, _ := NewMiniblocksProcessor(mock.HasherMock{}, &mock.MarshalizerMock{})

	miniblocks := []*data.Miniblock{
		{Hash: "h1", SenderShardID: 1, ReceiverShardID: 1, SenderBlockHash: "senderBlock",
			ProcessingTypeOnSource: block.Scheduled.String()},
	}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	mp.SerializeBulkMiniBlocks(miniblocks, buffSlice, "miniblocks", 1)

	expectedBuff := `{ "update" : {"_index":"miniblocks", "_id" : "h1" } }
{"scripted_upsert": true, "script": {"source": "if ('create' == ctx.op) {ctx._source = params.mb} else {if (params.npos) {ctx._source.senderBlockHash = params.mb.senderBlockHash;ctx._source.senderBlockNonce = params.mb.senderBlockNonce;ctx._source.procTypeS = params.mb.procTypeS;} else {ctx._source.receiverBlockHash = params.mb.receiverBlockHash;ctx._source.receiverBlockNonce = params.mb.receiverBlockNonce;ctx._source.procTypeD = params.mb.procTypeD;}}","lang": "painless","params": { "mb": {"senderShard":1,"receiverShard":1,"senderBlockHash":"senderBlock","type":"","procTypeS":"Scheduled","timestamp":0}, "npos": true }},"upsert": {}}
`
	require.Equal(t, expectedBuff, buffSlice.Buffers()[0].String())

	miniblocks = []*data.Miniblock{
		{Hash: "h1", SenderShardID: 1, ReceiverShardID: 1, ReceiverBlockHash: "receiverBlock",
			ProcessingTypeOnDestination: block.Processed.String()},
	}

	buffSlice = data.NewBufferSlice(data.DefaultMaxBulkSize)
	mp.SerializeBulkMiniBlocks(miniblocks, buffSlice, "miniblocks", 1)

	expectedBuff = `{ "update" : {"_index":"miniblocks", "_id" : "h1" } }
{"scripted_upsert": true, "script": {"source": "if ('create' == ctx.op) {ctx._source = params.mb} else {if (params.npos) {ctx._source.senderBlockHash = params.mb.senderBlockHash;ctx._source.senderBlockNonce = params.mb.senderBlockNonce;ctx._source.procTypeS = params.mb.procTypeS;} else {ctx._source.receiverBlockHash = params.mb.receiverBlockHash;ctx._source.receiverBlockNonce = params.mb.receiverBlockNonce;ctx._source.procTypeD = params.mb.procTypeD;}}","lang": "painless","params": { "mb": {"senderShard":1,"receiverShard":1,"receiverBlockHash":"receiverBlock","type":"","procTypeD":"Processed","timestamp":0}, "npos": false }},"upsert": {}}
`
	require.Equal(t, expectedBuff, buffSlice.Buffers()[0].String())
}
