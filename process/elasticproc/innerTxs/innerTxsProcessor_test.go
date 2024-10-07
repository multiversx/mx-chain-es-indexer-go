package innerTxs

import (
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInnerTxsProcessor_ExtractInnerTxsNoTransactions(t *testing.T) {
	t.Parallel()

	innerTxsProc := NewInnerTxsProcessor()

	res := innerTxsProc.ExtractInnerTxs(nil)
	require.Equal(t, 0, len(res))

	res = innerTxsProc.ExtractInnerTxs([]*data.Transaction{{}, {}})
	require.Equal(t, 0, len(res))
}

func TestInnerTxsProcessor_ExtractInnerTxs(t *testing.T) {
	t.Parallel()

	innerTxsProc := NewInnerTxsProcessor()

	res := innerTxsProc.ExtractInnerTxs([]*data.Transaction{{
		Hash: "txHash",
		InnerTransactions: []*data.InnerTransaction{
			{
				Hash: "inner1",
			},
			{
				Hash: "inner2",
			},
		},
	}})

	require.Equal(t, 2, len(res))
	require.Equal(t, []*data.InnerTransaction{
		{
			Type:          InnerTxType,
			Hash:          "inner1",
			RelayedTxHash: "txHash",
		},
		{
			Type:          InnerTxType,
			Hash:          "inner2",
			RelayedTxHash: "txHash",
		},
	}, res)
}
