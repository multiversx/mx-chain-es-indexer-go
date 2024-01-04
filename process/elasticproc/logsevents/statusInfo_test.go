package logsevents

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestStatusInfoAddRecord(t *testing.T) {
	t.Parallel()

	statusInfoProc := newTxHashStatusInfoProcessor()

	txHash := "txHash1"
	statusInfoProc.addRecord(txHash, &outport.StatusInfo{
		CompletedEvent: true,
		ErrorEvent:     false,
		Status:         transaction.TxStatusSuccess.String(),
	})
	require.Equal(t, &outport.StatusInfo{
		CompletedEvent: true,
		Status:         "success",
	}, statusInfoProc.getAllRecords()[txHash])

	statusInfoProc.addRecord(txHash, &outport.StatusInfo{
		ErrorEvent: true,
		Status:     transaction.TxStatusFail.String(),
	})
	require.Equal(t, &outport.StatusInfo{
		CompletedEvent: true,
		ErrorEvent:     true,
		Status:         "fail",
	}, statusInfoProc.getAllRecords()[txHash])

	statusInfoProc.addRecord(txHash, &outport.StatusInfo{
		ErrorEvent:     false,
		CompletedEvent: false,
		Status:         transaction.TxStatusSuccess.String(),
	})
	require.Equal(t, &outport.StatusInfo{
		CompletedEvent: true,
		ErrorEvent:     true,
		Status:         "fail",
	}, statusInfoProc.getAllRecords()[txHash])
}
