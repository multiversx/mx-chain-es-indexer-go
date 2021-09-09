package transactions

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func TestAttachSCRsToTransactionsAndReturnSCRsWithoutTx(t *testing.T) {
	t.Parallel()

	scrsDataToTxs := newScrsDataToTransactions(&mock.EconomicsHandlerMock{})

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	tx1 := &data.Transaction{
		Hash:     hex.EncodeToString(txHash1),
		Nonce:    1,
		Sender:   "sender",
		Receiver: "receiver",
		GasLimit: 10000000,
		GasPrice: 1000000000,
		Data:     []byte("callSomething"),
	}
	tx2 := &data.Transaction{}
	txs := map[string]*data.Transaction{
		string(txHash1): tx1,
		string(txHash2): tx2,
	}
	scrs := []*data.ScResult{
		{
			Nonce:          2,
			Sender:         "receiver",
			Receiver:       "sender",
			OriginalTxHash: hex.EncodeToString(txHash1),
			PrevTxHash:     hex.EncodeToString(txHash1),
			Value:          "40365000000000",
			Data:           []byte("@6f6b"),
		},
		{
			OriginalTxHash: "0102030405",
		},
	}

	scrsWithoutTx := scrsDataToTxs.attachSCRsToTransactionsAndReturnSCRsWithoutTx(txs, scrs)
	require.Len(t, scrsWithoutTx, 1)
	require.Len(t, tx1.SmartContractResults, 1)
	require.Equal(t, uint64(5963500), tx1.GasUsed)
	require.Equal(t, "128440000000000", tx1.Fee)

	require.Equal(t, scrsWithoutTx[0].OriginalTxHash, "0102030405")
}

func TestProcessTransactionsAfterSCRsWereAttached(t *testing.T) {
	t.Parallel()

	scrsDataToTxs := newScrsDataToTransactions(&mock.EconomicsHandlerMock{})

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	txHash3 := []byte("txHash3")
	tx1 := &data.Transaction{
		Hash:     hex.EncodeToString(txHash1),
		Nonce:    1,
		Sender:   "sender",
		Receiver: "receiver",
		GasLimit: 10000000,
		GasPrice: 1000000000,
		Data:     []byte("callSomething"),
		SmartContractResults: []*data.ScResult{
			{
				ReturnMessage: "user error",
			},
		},
	}
	tx2 := &data.Transaction{}
	tx3 := &data.Transaction{
		GasLimit: 5000000,
		GasPrice: 1000000000,
		Data:     []byte("relayedTxV2@01020304"),
		SmartContractResults: []*data.ScResult{
			{},
		},
	}
	txs := map[string]*data.Transaction{
		string(txHash1): tx1,
		string(txHash2): tx2,
		string(txHash3): tx3,
	}

	scrsDataToTxs.processTransactionsAfterSCRsWereAttached(txs)
	require.Equal(t, "fail", tx1.Status)
	require.Equal(t, tx1.GasLimit, tx1.GasUsed)
	require.Equal(t, "168805000000000", tx1.Fee)

	require.Equal(t, tx3.GasLimit, tx3.GasUsed)
	require.Equal(t, "129200000000000", tx3.Fee)
}

func TestIsESDTNFTTransferWithUserError(t *testing.T) {
	t.Parallel()

	require.False(t, isESDTNFTTransferWithUserError("ESDTNFTTransfer@45474c444d4558462d333766616239@06f5@045d2bd2629df0d2ea@0801120a00045d2bd2629df0d2ea226408f50d1a2000000000000000000500e809539d1d8febc54df4e6fe826fdc8ab6c88cf07ceb32003a3b00000007401c82df9c05a80000000000000407000000000000040f010000000009045d2bd2629df0d2ea0000000000000009045d2bd2629df0d2ea@636c61696d52657761726473"))
	require.False(t, isESDTNFTTransferWithUserError("ESDTTransfer@4d45582d623662623764@74b7e37e3c2efe5f11@"))
	require.False(t, isESDTNFTTransferWithUserError("ESDTNFTTransfer@45474c444d4558462d333766616239@070f@045d2bd2629df0d2ea@0801120a00045d2bd2629df0d2ea2264088f0e1a2000000000000000000500e809539d1d8febc54df4e6fe826fdc8ab6c88cf07ceb32003a3b000000074034d62af2b6930000000000000407000000000000040f010000000009045d2bd2629df0d2ea0000000000000009045d2bd2629df0d2ea@"))
}
