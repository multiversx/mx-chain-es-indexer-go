package transactions

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestNewEsdtTransactionHandler(t *testing.T) {
	t.Parallel()

	esdtTxProc := newEsdtTransactionHandler(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{})

	tx1 := &transaction.Transaction{
		Data: []byte(`ESDTTransfer@544b4e2d626231323061@010f0cf064dd59200000`),
	}

	tokenIdentifier := esdtTxProc.getTokenIdentifier(tx1.Data)
	require.Equal(t, "TKN-bb120a", tokenIdentifier)
}

func TestIsEsdtTransaction(t *testing.T) {
	t.Parallel()

	esdtProc := newEsdtTransactionHandler(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{})

	require.True(t, esdtProc.isESDTTx([]byte("ESDTTransfer@01@01")))
	require.False(t, esdtProc.isESDTTx([]byte("EESDTTransfer@01@01")))
	require.False(t, esdtProc.isESDTTx([]byte("")))
}
