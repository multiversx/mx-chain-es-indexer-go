package transactions

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
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
	}
	scResult2 := &data.ScResult{
		Hash:          "hash2",
		Nonce:         2,
		GasPrice:      10,
		GasLimit:      50,
		SenderShard:   2,
		ReceiverShard: 3,
	}
	scrs := []*data.ScResult{scResult1, scResult2}

	res, err := (&txsDatabaseProcessor{}).SerializeScResults(scrs)
	require.Nil(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "hash1" } }
{"nonce":1,"gasLimit":50,"gasPrice":10,"value":"","sender":"","receiver":"","senderShard":0,"receiverShard":1,"prevTxHash":"","originalTxHash":"","callType":"","timestamp":0}
{ "index" : { "_id" : "hash2" } }
{"nonce":2,"gasLimit":50,"gasPrice":10,"value":"","sender":"","receiver":"","senderShard":2,"receiverShard":3,"prevTxHash":"","originalTxHash":"","callType":"","timestamp":0}
`
	require.Equal(t, expectedRes, res[0].String())
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

	res, err := (&txsDatabaseProcessor{}).SerializeReceipts(recs)
	require.Nil(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "recHash1" } }
{"value":"","sender":"sender1","txHash":"txHash1"}
{ "index" : { "_id" : "recHash2" } }
{"value":"","sender":"sender2","txHash":"txHash2"}
`
	require.Equal(t, expectedRes, res[0].String())
}

func TestSerializeTransactionsIntraShardTx(t *testing.T) {
	t.Parallel()

	buffers, err := (&txsDatabaseProcessor{}).SerializeTransactions([]*data.Transaction{{
		Hash:                 "txHash",
		SmartContractResults: map[string]*data.ScResult{},
	}}, map[string]string{}, 0, []*data.ScResult{})
	require.Nil(t, err)

	expectedBuff := `{ "index" : { "_id" : "txHash", "_type" : "_doc" } }
{"miniBlockHash":"","nonce":0,"round":0,"value":"","receiver":"","sender":"","receiverShard":0,"senderShard":0,"gasPrice":0,"gasLimit":0,"gasUsed":0,"fee":"","data":null,"signature":"","timestamp":0,"status":"","searchOrder":0}
`
	require.Equal(t, expectedBuff, buffers[0].String())
}

func TestSerializeTransactionCrossShardTxSource(t *testing.T) {
	t.Parallel()

	buffers, err := (&txsDatabaseProcessor{}).SerializeTransactions([]*data.Transaction{{
		Hash:                 "txHash",
		SenderShard:          0,
		ReceiverShard:        1,
		SmartContractResults: map[string]*data.ScResult{},
	}}, map[string]string{}, 0, []*data.ScResult{})
	require.Nil(t, err)

	expectedBuff := `{"update":{"_id":"txHash", "_type": "_doc"}}
{"script":{"source":"return"},"upsert":{"miniBlockHash":"","nonce":0,"round":0,"value":"","receiver":"","sender":"","receiverShard":1,"senderShard":0,"gasPrice":0,"gasLimit":0,"gasUsed":0,"fee":"","data":null,"signature":"","timestamp":0,"status":"","searchOrder":0}}
`
	require.Equal(t, expectedBuff, buffers[0].String())
}

func TestSerializeTransactionsCrossShardTxDestination(t *testing.T) {
	t.Parallel()

	buffers, err := (&txsDatabaseProcessor{}).SerializeTransactions([]*data.Transaction{{
		Hash:          "txHash",
		SenderShard:   1,
		ReceiverShard: 0,
	}}, map[string]string{}, 0, []*data.ScResult{})
	require.Nil(t, err)

	expectedBuff := `{ "index" : { "_id" : "txHash", "_type" : "_doc" } }
{"miniBlockHash":"","nonce":0,"round":0,"value":"","receiver":"","sender":"","receiverShard":0,"senderShard":1,"gasPrice":0,"gasLimit":0,"gasUsed":0,"fee":"","data":null,"signature":"","timestamp":0,"status":"","searchOrder":0}
`
	require.Equal(t, expectedBuff, buffers[0].String())
}

func TestTxsDatabaseProcessor_SerializeTransactionWithRefund(t *testing.T) {
	t.Parallel()

	txs := map[string]*data.Transaction{
		"txHash": {
			Sender:   "sender",
			Receiver: "receiver",
			GasLimit: 150000000,
			GasPrice: 1000000000,
		},
	}
	txHashRefund := map[string]*data.RefundData{
		"txHash": {
			Value:    "101676480000000",
			Receiver: "sender",
		},
	}
	buffers, err := (&txsDatabaseProcessor{
		txFeeCalculator: &mock.EconomicsHandlerMock{},
	}).SerializeTransactionWithRefund(txs, txHashRefund)
	require.Nil(t, err)

	expectedBuff := `{ "index" : { "_id" : "txHash" } }
{"miniBlockHash":"","nonce":0,"round":0,"value":"","receiver":"receiver","sender":"sender","receiverShard":0,"senderShard":0,"gasPrice":1000000000,"gasLimit":150000000,"gasUsed":139832352,"fee":"1447823520000000","data":null,"signature":"","timestamp":0,"status":"","searchOrder":0}
`
	require.Equal(t, expectedBuff, buffers[0].String())
}
