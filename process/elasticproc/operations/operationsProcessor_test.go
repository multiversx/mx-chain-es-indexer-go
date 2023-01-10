package operations

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestNewOperationsProcessor(t *testing.T) {
	t.Parallel()

	op, err := NewOperationsProcessor()
	require.NotNil(t, op)
	require.Nil(t, err)
}

func TestOperationsProcessor_ProcessTransactionsAndSCRSTransactions(t *testing.T) {
	t.Parallel()

	op, _ := NewOperationsProcessor()

	txs := []*data.Transaction{
		{},
		{
			ReceiverShard: 1,
		},
	}

	processedTxs, _ := op.ProcessTransactionsAndSCRs(txs, nil, true, 0)
	require.Equal(t, []*data.Transaction{
		{Type: string(transaction.TxTypeNormal)},
	}, processedTxs)
}

func TestOperationsProcessor_ProcessTransactionsAndSCRSSmartContractResults(t *testing.T) {
	t.Parallel()

	op, _ := NewOperationsProcessor()

	scrs := []*data.ScResult{
		{},
		{
			ReceiverShard: 1,
		},
	}

	_, processedSCRs := op.ProcessTransactionsAndSCRs(nil, scrs, true, 0)
	require.Equal(t, []*data.ScResult{
		{Type: string(transaction.TxTypeUnsigned), Status: transaction.TxStatusSuccess.String()},
	}, processedSCRs)
}

func TestOperationsProcessor_ShouldIgnoreSCRs(t *testing.T) {
	t.Parallel()

	op, _ := NewOperationsProcessor()

	scrs := []*data.ScResult{
		{
			ReturnMessage: data.GasRefundForRelayerMessage,
			Data:          nil,
		},
		{
			Data: []byte("@6f6b"),
		},
		{
			Operation:          "ESDTNFTTransfer",
			SenderAddressBytes: []byte("sender"),
		},
	}

	_, processedSCRs := op.ProcessTransactionsAndSCRs(nil, scrs, false, 0)
	for _, scr := range processedSCRs {
		require.True(t, scr.CanBeIgnored)
	}
}
