package operations

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func TestOperationsProcessor_SerializeSCRS(t *testing.T) {
	t.Parallel()

	op, _ := NewOperationsProcessor(false, &mock.ShardCoordinatorMock{})

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

	res, err := op.SerializeSCRs(scrs)
	require.Nil(t, err)
	require.Equal(t, `{"update":{"_id":"", "_type": "_doc"}}
{"script":{"source":"return"},"upsert":{"nonce":0,"gasLimit":0,"gasPrice":0,"value":"","sender":"","receiver":"","senderShard":0,"receiverShard":1,"prevTxHash":"","originalTxHash":"","callType":"","timestamp":0}}
{ "index" : { "_id" : "", "_type" : "_doc" } }
{"nonce":0,"gasLimit":0,"gasPrice":0,"value":"","sender":"","receiver":"","senderShard":2,"receiverShard":0,"prevTxHash":"","originalTxHash":"","callType":"","timestamp":0}
`, res[0].String())
}
