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
	require.True(t, esdtProc.isESDTTx([]byte("ESDTNFTAddQuantity@01@01")))
	require.False(t, esdtProc.isESDTTx([]byte("ESDTNFTAddQuantitya@01@01")))
	require.False(t, esdtProc.isESDTTx([]byte("EESDTTransfer@01@01")))
	require.False(t, esdtProc.isESDTTx([]byte("")))
}

func TestIsNftTransaction(t *testing.T) {
	t.Parallel()

	esdtProc := newEsdtTransactionHandler(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{})

	require.True(t, esdtProc.isNFTTx([]byte("ESDTNFTAddQuantity@01@01")))
	require.False(t, esdtProc.isNFTTx([]byte("ESDTTransfer@01@01")))
	require.False(t, esdtProc.isNFTTx([]byte("")))
}

func TestGetNFTInfo(t *testing.T) {
	t.Parallel()

	esdtProc := newEsdtTransactionHandler(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{})

	tokenIdentifier, nonce := esdtProc.getNFTTxInfo([]byte("ESDTNFTTransfer@544f4b454e2d666437653066@01@01@b7a5acba50ff6a2821876693a4e62d60ec8645af696591e04ead2e2cb6e4cb4f"))
	require.Equal(t, "TOKEN-fd7e0f", tokenIdentifier)
	require.Equal(t, uint64(1), nonce)

	tokenIdentifier, nonce = esdtProc.getNFTTxInfo([]byte("@01@01@b7a5acba50ff6a2821876693a4e62d60ec8645af696591e04ead2e2cb6e4cb4f"))
	require.Equal(t, "", tokenIdentifier)
	require.Equal(t, uint64(0), nonce)

	tokenIdentifier, nonce = esdtProc.getNFTTxInfo([]byte("myMethod"))
	require.Equal(t, "", tokenIdentifier)
	require.Equal(t, uint64(0), nonce)

	tokenIdentifier, nonce = esdtProc.getNFTTxInfo([]byte("myMethod@01"))
	require.Equal(t, "", tokenIdentifier)
	require.Equal(t, uint64(0), nonce)

	tokenIdentifier, nonce = esdtProc.getNFTTxInfo([]byte("ESDTNFTTransfer@544f4b454e2d666437653066"))
	require.Equal(t, "TOKEN-fd7e0f", tokenIdentifier)
	require.Equal(t, uint64(0), nonce)
}
