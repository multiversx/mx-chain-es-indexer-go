package transactions

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
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

func TestSearchTxsWithNFTCreateAndPutNonceInAlteredAddress(t *testing.T) {
	t.Parallel()

	esdtProc := newEsdtTransactionHandler(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{})

	txHash := []byte("txHash")
	alteredAddresses := data.NewAlteredAccounts()
	alteredAddresses.Add("sender", &data.AlteredAccount{})

	alteredAddressesEmpty := data.NewAlteredAccounts()

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

	altered, ok := alteredAddresses.Get("sender")
	require.True(t, ok)

	require.Equal(t, &data.AlteredAccount{
		NFTNonce:       1,
		IsNFTOperation: true,
	}, altered[0])

	esdtProc.searchTxsWithNFTCreateAndPutNonceInAlteredAddress(alteredAddressesEmpty, txs, scrs)
	require.Equal(t, 1, alteredAddressesEmpty.Len())
}

func TestSearchSCRSWithCreateNFTAndPutNonceInAlteredAddress(t *testing.T) {
	t.Parallel()

	esdtProc := newEsdtTransactionHandler(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{})

	alteredAddresses := data.NewAlteredAccounts()
	alteredAddresses.Add("sender", &data.AlteredAccount{})

	txHash := []byte("txHash")
	scrs := []*data.ScResult{
		{
			OriginalTxHash: hex.EncodeToString([]byte("anotherHash")),
		},
		{
			OriginalTxHash:      hex.EncodeToString(txHash),
			Sender:              "sender",
			Receiver:            "sender",
			EsdtTokenIdentifier: "my-token",
			Data:                []byte("ESDTNFTCreate@4d494841492d666437653066@01@6d796e6674@0b@@@"),
		},
		{
			OriginalTxHash: hex.EncodeToString(txHash),
			Sender:         "sender",
			Receiver:       "sender",
			Data:           []byte("@6f6b@01"),
		},
	}

	esdtProc.searchForESDTInScrs(alteredAddresses, scrs)

	altered, ok := alteredAddresses.Get("sender")
	require.True(t, ok)

	require.Equal(t, &data.AlteredAccount{
		NFTNonce:        1,
		IsNFTOperation:  true,
		TokenIdentifier: "my-token",
	}, altered[0])
}

func TestSearchTxWithNFTTransfer(t *testing.T) {
	t.Parallel()

	esdtProc := newEsdtTransactionHandler(mock.NewPubkeyConverterMock(9), &mock.ShardCoordinatorMock{})
	alteredAddresses := data.NewAlteredAccounts()

	txs := map[string]*data.Transaction{
		"hash": {
			Sender:   "sender",
			Receiver: "receiver",
			Data:     []byte("ESDTNFTTransfer@746f6b656e@01@01@726563656976657231"),
		},
	}

	esdtProc.searchForReceiverNFTTransferAndPutInAlteredAddress(txs, alteredAddresses)
	res, ok := alteredAddresses.Get("726563656976657231")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsSender:        false,
		IsESDTOperation: false,
		IsNFTOperation:  true,
		TokenIdentifier: "token",
		NFTNonce:        1,
	}, res[0])
}

func TestSearchTxWithNFTTransferWrongAddress(t *testing.T) {
	t.Parallel()

	addressEncoder, _ := pubkeyConverter.NewBech32PubkeyConverter(32)
	esdtProc := newEsdtTransactionHandler(addressEncoder, &mock.ShardCoordinatorMock{})
	alteredAddresses := data.NewAlteredAccounts()

	txs := map[string]*data.Transaction{
		"hash": {
			Sender:   "sender",
			Receiver: "receiver",
			Data:     []byte("ESDTNFTTransfer@746f6b656e@01@01@726563656976657231"),
		},
	}

	esdtProc.searchForReceiverNFTTransferAndPutInAlteredAddress(txs, alteredAddresses)
	_, ok := alteredAddresses.Get("726563656976657231")
	require.False(t, ok)
}
