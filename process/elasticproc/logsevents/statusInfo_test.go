package logsevents

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/stretchr/testify/require"
)

func TestStatusInfoAddRecord(t *testing.T) {
	t.Parallel()

	statusInfOProc := newTxHashStatusInfo()

	txHash := "txHash1"
	statusInfOProc.addRecord(txHash, &data.StatusInfo{
		CompletedEvent: true,
		ErrorEvent:     false,
		Status:         transaction.TxStatusSuccess.String(),
	})
	require.Equal(t, statusInfOProc.getAllRecords()[txHash], &data.StatusInfo{
		CompletedEvent: true,
		Status:         "success",
	})

	statusInfOProc.addRecord(txHash, &data.StatusInfo{
		ErrorEvent: true,
		Status:     transaction.TxStatusFail.String(),
	})
	require.Equal(t, statusInfOProc.getAllRecords()[txHash], &data.StatusInfo{
		CompletedEvent: true,
		ErrorEvent:     true,
		Status:         "fail",
	})

	statusInfOProc.addRecord(txHash, &data.StatusInfo{
		ErrorEvent:     false,
		CompletedEvent: false,
		Status:         transaction.TxStatusSuccess.String(),
	})
	require.Equal(t, statusInfOProc.getAllRecords()[txHash], &data.StatusInfo{
		CompletedEvent: true,
		ErrorEvent:     true,
		Status:         "fail",
	})
}
