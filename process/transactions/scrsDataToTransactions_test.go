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
