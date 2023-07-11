package transactions

import (
	"encoding/hex"
	"testing"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
	"github.com/stretchr/testify/require"
)

func TestAttachSCRsToTransactionsAndReturnSCRsWithoutTx(t *testing.T) {
	t.Parallel()

	bc, _ := converters.NewBalanceConverter(18)
	scrsDataToTxs := newScrsDataToTransactions(bc)

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
		GasUsed:  5963500,
		Fee:      "128440000000000",
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

	bc, _ := converters.NewBalanceConverter(18)
	scrsDataToTxs := newScrsDataToTransactions(bc)

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
		SmartContractResults: []*data.ScResult{
			{
				ReturnMessage: "user error",
			},
		},
		GasUsed: 10000000,
		Fee:     "168805000000000",
	}
	tx2 := &data.Transaction{}
	txs := map[string]*data.Transaction{
		string(txHash1): tx1,
		string(txHash2): tx2,
	}

	scrsDataToTxs.processTransactionsAfterSCRsWereAttached(txs)
	require.Equal(t, "", tx1.Status)
	require.Equal(t, tx1.GasLimit, tx1.GasUsed)
	require.Equal(t, "168805000000000", tx1.Fee)
}
