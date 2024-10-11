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
			ID:            "054efde0c8bb1ee713c9fe5981340d7efbf23b7aa72abeab9a63b64c21000188",
			Type:          InnerTxType,
			Hash:          "inner1",
			RelayedTxHash: "txHash",
		},
		{
			ID:            "e50a3fd30c5c7bba673dd18a0b329760f1bff34342978821c5d341067da70fa1",
			Type:          InnerTxType,
			Hash:          "inner2",
			RelayedTxHash: "txHash",
		},
	}, res)
}
