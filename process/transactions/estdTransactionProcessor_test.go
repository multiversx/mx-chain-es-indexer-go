package transactions

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestNewEsdtTransactionHandler(t *testing.T) {
	t.Parallel()

	esdtTxProc := newEsdtTransactionHandler()

	tx1 := &transaction.Transaction{
		Data: []byte(`ESDTTransfer@544b4e2d626231323061@010f0cf064dd59200000`),
	}

	tokenIdentifier := esdtTxProc.getTokenIdentifier(tx1.Data)
	require.Equal(t, "TKN-bb120a", tokenIdentifier)
}

func TestIsEsdtTransaction(t *testing.T) {
	t.Parallel()

	esdtProc := newEsdtTransactionHandler()

	require.True(t, esdtProc.isESDTTx([]byte("ESDTTransfer@01@01")))
	require.True(t, esdtProc.isESDTTx([]byte("ESDTNFTAddQuantity@01@01")))
	require.False(t, esdtProc.isESDTTx([]byte("ESDTNFTAddQuantitya@01@01")))
	require.False(t, esdtProc.isESDTTx([]byte("EESDTTransfer@01@01")))
	require.False(t, esdtProc.isESDTTx([]byte("")))
}

func TestIsNftTransaction(t *testing.T) {
	t.Parallel()

	esdtProc := newEsdtTransactionHandler()

	require.True(t, esdtProc.isNFTTx([]byte("ESDTNFTAddQuantity@01@01")))
	require.False(t, esdtProc.isNFTTx([]byte("ESDTTransfer@01@01")))
	require.False(t, esdtProc.isNFTTx([]byte("")))
}

func TestGetNFTInfo(t *testing.T) {
	t.Parallel()

	esdtProc := newEsdtTransactionHandler()

	tokenIdentifier, nonceStr := esdtProc.getNFTTxInfo([]byte("ESDTNFTTransfer@4d494841492d666437653066@01@01@b7a5acba50ff6a2821876693a4e62d60ec8645af696591e04ead2e2cb6e4cb4f"))
	require.Equal(t, "MIHAI-fd7e0f", tokenIdentifier)
	require.Equal(t, "1", nonceStr)

	tokenIdentifier, nonceStr = esdtProc.getNFTTxInfo([]byte("@01@01@b7a5acba50ff6a2821876693a4e62d60ec8645af696591e04ead2e2cb6e4cb4f"))
	require.Equal(t, "", tokenIdentifier)
	require.Equal(t, "", nonceStr)

	tokenIdentifier, nonceStr = esdtProc.getNFTTxInfo([]byte("myMethod"))
	require.Equal(t, "", tokenIdentifier)
	require.Equal(t, "", nonceStr)

	tokenIdentifier, nonceStr = esdtProc.getNFTTxInfo([]byte("myMethod@01"))
	require.Equal(t, "", tokenIdentifier)
	require.Equal(t, "", nonceStr)

	tokenIdentifier, nonceStr = esdtProc.getNFTTxInfo([]byte("ESDTNFTTransfer@4d494841492d666437653066"))
	require.Equal(t, "MIHAI-fd7e0f", tokenIdentifier)
	require.Equal(t, "", nonceStr)
}

func TestSearchTxsWithNFTCreateAndPutNonceInAlteredAddress(t *testing.T) {
	t.Parallel()

	esdtProc := newEsdtTransactionHandler()

	txHash := []byte("txHash")
	alteredAddresses := map[string]*data.AlteredAccount{
		"sender": {},
	}
	alteredAddressesEmpty := map[string]*data.AlteredAccount{}

	txs := map[string]*data.Transaction{
		string(txHash): {
			Sender:   "sender",
			Receiver: "sender",
			Data:     []byte("ESDTNFTCreate@4d494841492d666437653066@01@6d796e6674@0b@@@"),
		},
	}
	scrs := []*data.ScResult{
		{
			OriginalTxHash: hex.EncodeToString([]byte("anotherHash")),
		},
		{
			OriginalTxHash: hex.EncodeToString(txHash),
			Sender:         "sender",
			Receiver:       "receiver",
			Data:           []byte("@6f6b@01"),
		},
		{
			OriginalTxHash: hex.EncodeToString(txHash),
			Sender:         "sender",
			Receiver:       "sender",
			Data:           []byte("error"),
		},
		{
			OriginalTxHash: hex.EncodeToString(txHash),
			Sender:         "sender",
			Receiver:       "sender",
			Data:           []byte("error@error@error"),
		},
		{
			OriginalTxHash: hex.EncodeToString(txHash),
			Sender:         "sender",
			Receiver:       "sender",
			Data:           append([]byte("@6f6b@"), []byte{10}...),
		},
		{
			OriginalTxHash: hex.EncodeToString(txHash),
			Sender:         "sender",
			Receiver:       "sender",
			Data:           []byte("@6f6b@01"),
		},
	}

	esdtProc.searchTxsWithNFTCreateAndPutNonceInAlteredAddress(alteredAddresses, txs, scrs)
	require.Equal(t, &data.AlteredAccount{
		NFTNonceString: "1",
	}, alteredAddresses["sender"])

	esdtProc.searchTxsWithNFTCreateAndPutNonceInAlteredAddress(alteredAddressesEmpty, txs, scrs)
	require.Equal(t, 0, len(alteredAddressesEmpty))
}

func TestSearchSCRSWithCreateNFTAndPutNonceInAlteredAddress(t *testing.T) {
	t.Parallel()

	esdtProc := newEsdtTransactionHandler()
	alteredAddresses := map[string]*data.AlteredAccount{
		"sender": {},
	}

	txHash := []byte("txHash")
	scrs := []*data.ScResult{
		{
			OriginalTxHash: hex.EncodeToString([]byte("anotherHash")),
		},
		{
			OriginalTxHash: hex.EncodeToString(txHash),
			Sender:         "sender",
			Receiver:       "sender",
			Data:           []byte("ESDTNFTCreate@4d494841492d666437653066@01@6d796e6674@0b@@@"),
		},
		{
			OriginalTxHash: hex.EncodeToString(txHash),
			Sender:         "sender",
			Receiver:       "sender",
			Data:           []byte("@6f6b@01"),
		},
	}

	esdtProc.searchSCRSWithCreateNFTAndPutNonceInAlteredAddress(alteredAddresses, scrs)
	require.Equal(t, &data.AlteredAccount{
		NFTNonceString: "1",
	}, alteredAddresses["sender"])
}
