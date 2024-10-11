package innerTxs

import (
	"testing"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func TestInnerTxsProcessor_ExtractInnerTxsNoTransactions(t *testing.T) {
	t.Parallel()

	innerTxsProc, _ := NewInnerTxsProcessor(&mock.HasherMock{})

	res := innerTxsProc.ExtractInnerTxs(nil)
	require.Equal(t, 0, len(res))

	res = innerTxsProc.ExtractInnerTxs([]*data.Transaction{{}, {}})
	require.Equal(t, 0, len(res))
}

func TestInnerTxsProcessor_ExtractInnerTxs(t *testing.T) {
	t.Parallel()

	innerTxsProc, _ := NewInnerTxsProcessor(&mock.HasherMock{})

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
			ID:            "b0a05b65beb800084f80b9b5c7c5894378213cd58e09663aa99322597f86c23e",
			Type:          InnerTxType,
			Hash:          "inner1",
			RelayedTxHash: "txHash",
		},
		{
			ID:            "a9188f77578a4b629c91a237e66f2e833d3f9b850f1b9341bbeab5c7e38a14d9",
			Type:          InnerTxType,
			Hash:          "inner2",
			RelayedTxHash: "txHash",
		},
	}, res)
}
