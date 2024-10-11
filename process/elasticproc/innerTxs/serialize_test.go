package innerTxs

import (
	"testing"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestInnerTxsProcessor_SerializeInnerTxs(t *testing.T) {
	t.Parallel()

	innerTxs := []*data.InnerTransaction{
		{
			ID:   "id1",
			Hash: "h1",
			Type: InnerTxType,
		},
		{
			ID:   "id2",
			Hash: "h2",
			Type: InnerTxType,
		},
	}

	innerTxsProc, _ := NewInnerTxsProcessor(&mock.HasherMock{})
	buffSlice := data.NewBufferSlice(0)
	err := innerTxsProc.SerializeInnerTxs(innerTxs, buffSlice, dataindexer.OperationsIndex)
	require.Nil(t, err)
	require.Equal(t, `{ "index" : { "_index":"operations","_id" : "id1" } }
{"hash":"h1","type":"innerTx","nonce":0,"value":"","receiver":"","sender":"","gasPrice":0,"gasLimit":0,"chainID":"","version":0}
{ "index" : { "_index":"operations","_id" : "id2" } }
{"hash":"h2","type":"innerTx","nonce":0,"value":"","receiver":"","sender":"","gasPrice":0,"gasLimit":0,"chainID":"","version":0}
`, buffSlice.Buffers()[0].String())
}
