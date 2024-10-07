package innerTxs

import (
	"testing"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestInnerTxsProcessor_SerializeInnerTxs(t *testing.T) {
	t.Parallel()

	innerTxs := []*data.InnerTransaction{
		{
			Hash: "h1",
			Type: InnerTxType,
		},
		{
			Hash: "h2",
			Type: InnerTxType,
		},
	}

	innerTxsProc := NewInnerTxsProcessor()
	buffSlice := data.NewBufferSlice(0)
	err := innerTxsProc.SerializeInnerTxs(innerTxs, buffSlice, dataindexer.OperationsIndex)
	require.Nil(t, err)
	require.Equal(t, `{ "index" : { "_index":"operations","_id" : "h1" } }
{"type":"innerTx","nonce":0,"value":"","receiver":"","sender":"","gasPrice":0,"gasLimit":0,"chainID":"","version":0}
{ "index" : { "_index":"operations","_id" : "h2" } }
{"type":"innerTx","nonce":0,"value":"","receiver":"","sender":"","gasPrice":0,"gasLimit":0,"chainID":"","version":0}
`, buffSlice.Buffers()[0].String())
}
