package operations

import (
	"testing"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/stretchr/testify/require"
)

func TestOperationsProcessor_SerializeSCRS(t *testing.T) {
	t.Parallel()

	op, _ := NewOperationsProcessor()

	scrs := []*data.ScResult{
		{
			SenderShard:   0,
			ReceiverShard: 1,
		},
		{
			SenderShard:   2,
			ReceiverShard: 0,
		},
	}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	err := op.SerializeSCRs(scrs, buffSlice, "operations", 0)
	require.Nil(t, err)
	require.Equal(t, `{"update":{"_index":"operations","_id":""}}
{"script":{"source":"return"},"upsert":{"uuid":"","nonce":0,"gasLimit":0,"gasPrice":0,"value":"","valueNum":0,"sender":"","receiver":"","senderShard":0,"receiverShard":1,"prevTxHash":"","originalTxHash":"","callType":"","timestamp":0,"epoch":0}}
{ "index" : { "_index":"operations","_id" : "" } }
{"uuid":"","nonce":0,"gasLimit":0,"gasPrice":0,"value":"","valueNum":0,"sender":"","receiver":"","senderShard":2,"receiverShard":0,"prevTxHash":"","originalTxHash":"","callType":"","timestamp":0,"epoch":0}
`, buffSlice.Buffers()[0].String())
}
