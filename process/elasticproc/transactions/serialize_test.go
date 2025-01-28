package transactions

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/stretchr/testify/require"
)

func TestSerializeScResults(t *testing.T) {
	t.Parallel()

	scResult1 := &data.ScResult{
		Hash:          "hash1",
		Nonce:         1,
		GasPrice:      10,
		GasLimit:      50,
		SenderShard:   0,
		ReceiverShard: 1,
		Value:         "100",
		ValueNum:      1e-16,
	}
	scResult2 := &data.ScResult{
		Hash:          "hash2",
		Nonce:         2,
		GasPrice:      10,
		GasLimit:      50,
		SenderShard:   2,
		ReceiverShard: 3,
		Value:         "20",
		ValueNum:      2e-17,
	}
	scrs := []*data.ScResult{scResult1, scResult2}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	err := (&txsDatabaseProcessor{}).SerializeScResults(scrs, buffSlice, "transactions")
	require.Nil(t, err)
	require.Equal(t, 1, len(buffSlice.Buffers()))

	expectedRes := `{ "index" : { "_index": "transactions", "_id" : "hash1" } }
{"uuid":"","nonce":1,"gasLimit":50,"gasPrice":10,"value":"100","valueNum":1e-16,"sender":"","receiver":"","senderShard":0,"receiverShard":1,"prevTxHash":"","originalTxHash":"","callType":"","timestamp":0}
{ "index" : { "_index": "transactions", "_id" : "hash2" } }
{"uuid":"","nonce":2,"gasLimit":50,"gasPrice":10,"value":"20","valueNum":2e-17,"sender":"","receiver":"","senderShard":2,"receiverShard":3,"prevTxHash":"","originalTxHash":"","callType":"","timestamp":0}
`
	require.Equal(t, expectedRes, buffSlice.Buffers()[0].String())
}

func TestSerializeReceipts(t *testing.T) {
	t.Parallel()

	rec1 := &data.Receipt{
		Hash:   "recHash1",
		Sender: "sender1",
		TxHash: "txHash1",
	}
	rec2 := &data.Receipt{
		Hash:   "recHash2",
		Sender: "sender2",
		TxHash: "txHash2",
	}

	recs := []*data.Receipt{rec1, rec2}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	err := (&txsDatabaseProcessor{}).SerializeReceipts(recs, buffSlice, "receipts")
	require.Nil(t, err)
	require.Equal(t, 1, len(buffSlice.Buffers()))

	expectedRes := `{ "index" : { "_index": "receipts", "_id" : "recHash1" } }
{"value":"","sender":"sender1","txHash":"txHash1","timestamp":0}
{ "index" : { "_index": "receipts", "_id" : "recHash2" } }
{"value":"","sender":"sender2","txHash":"txHash2","timestamp":0}
`
	require.Equal(t, expectedRes, buffSlice.Buffers()[0].String())
}

func TestSerializeTransactionsIntraShardTx(t *testing.T) {
	t.Parallel()

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	err := (&txsDatabaseProcessor{}).SerializeTransactions([]*data.Transaction{{
		Hash:                 "txHash",
		SmartContractResults: []*data.ScResult{{}},
	}}, map[string]*outport.StatusInfo{}, 0, buffSlice, "transactions")
	require.Nil(t, err)

	expectedBuff := `{ "index" : { "_index":"transactions", "_id" : "txHash" } }
{"uuid":"","miniBlockHash":"","nonce":0,"round":0,"value":"","valueNum":0,"receiver":"","sender":"","receiverShard":0,"senderShard":0,"gasPrice":0,"gasLimit":0,"gasUsed":0,"fee":"","feeNum":0,"data":null,"signature":"","timestamp":0,"status":"","searchOrder":0}
`
	require.Equal(t, expectedBuff, buffSlice.Buffers()[0].String())
}

func TestSerializeTransactionCrossShardTxSource(t *testing.T) {
	t.Parallel()

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	err := (&txsDatabaseProcessor{}).SerializeTransactions([]*data.Transaction{{
		Hash:                 "txHash",
		SenderShard:          0,
		ReceiverShard:        1,
		SmartContractResults: []*data.ScResult{{}},
		Version:              1,
	}}, map[string]*outport.StatusInfo{}, 0, buffSlice, "transactions")
	require.Nil(t, err)

	expectedBuff := `{"update":{ "_index":"transactions", "_id":"txHash"}}
{"script":{"source":"return"},"upsert":{"uuid":"","miniBlockHash":"","nonce":0,"round":0,"value":"","valueNum":0,"receiver":"","sender":"","receiverShard":1,"senderShard":0,"gasPrice":0,"gasLimit":0,"gasUsed":0,"fee":"","feeNum":0,"data":null,"signature":"","timestamp":0,"status":"","searchOrder":0,"version":1}}
`
	require.Equal(t, expectedBuff, buffSlice.Buffers()[0].String())
}

func TestSerializeTransactionsCrossShardTxDestination(t *testing.T) {
	t.Parallel()

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	err := (&txsDatabaseProcessor{}).SerializeTransactions([]*data.Transaction{{
		Hash:                 "txHash",
		SenderShard:          1,
		ReceiverShard:        0,
		SmartContractResults: []*data.ScResult{{}},
		Version:              1,
	}}, map[string]*outport.StatusInfo{}, 0, buffSlice, "transactions")
	require.Nil(t, err)

	expectedBuff := `{ "index" : { "_index":"transactions", "_id" : "txHash" } }
{"uuid":"","miniBlockHash":"","nonce":0,"round":0,"value":"","valueNum":0,"receiver":"","sender":"","receiverShard":0,"senderShard":1,"gasPrice":0,"gasLimit":0,"gasUsed":0,"fee":"","feeNum":0,"data":null,"signature":"","timestamp":0,"status":"","searchOrder":0,"version":1}
`
	require.Equal(t, expectedBuff, buffSlice.Buffers()[0].String())
}

func TestTxsDatabaseProcessor_SerializeTransactionWithRefund(t *testing.T) {
	t.Parallel()

	txHashRefund := map[string]*data.FeeData{
		"txHash": {
			FeeNum:   5e-15,
			Fee:      "100000",
			GasUsed:  5000,
			Receiver: "sender",
		},
	}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	err := (&txsDatabaseProcessor{}).SerializeTransactionsFeeData(txHashRefund, buffSlice, "transactions")
	require.Nil(t, err)

	expectedBuff := `{"update":{ "_index":"transactions","_id":"txHash"}}
{"scripted_upsert": true, "script": {"source": "if ('create' == ctx.op) {ctx.op = 'noop'} else {ctx._source.fee = params.fee;ctx._source.feeNum = params.feeNum;ctx._source.gasUsed = params.gasUsed;}","lang": "painless","params": {"fee": "100000", "gasUsed": 5000, "feeNum": 5e-15, "gasRefunded": 0}},"upsert": {}}
`
	require.Equal(t, expectedBuff, buffSlice.Buffers()[0].String())
}
