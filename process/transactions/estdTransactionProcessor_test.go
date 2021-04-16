package transactions

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestNewEsdtTransactionHandler(t *testing.T) {
	t.Parallel()

	esdtTxProc := newEsdtTransactionHandler()

	tx1 := &transaction.Transaction{
		Data: []byte(`ESDTTransfer@544b4e2d626231323061@010f0cf064dd59200000`),
	}

	tokenIdentifier, value := esdtTxProc.getTokenIdentifierAndValue(tx1)
	require.Equal(t, "TKN-bb120a", tokenIdentifier)
	require.Equal(t, "5000000000000000000000", value)
}
