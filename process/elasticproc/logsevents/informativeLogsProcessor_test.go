package logsevents

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/stretchr/testify/require"
)

func TestInformativeShouldIgnoreLog(t *testing.T) {
	informativeLogsProc := newInformativeLogsProcessor()

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte("doSomething"),
	}
	args := &argsProcessEvent{
		timestamp:  1234,
		event:      event,
		logAddress: []byte("contract"),
	}

	res := informativeLogsProc.processEvent(args)
	require.False(t, res.processed)
}

func TestInformativeLogsProcessorWriteLog(t *testing.T) {
	t.Parallel()

	tx := &data.Transaction{
		GasLimit: 500000,
		GasPrice: 100000,
		Data:     []byte("doSomething"),
	}

	hexEncodedTxHash := "01020304"
	txs := map[string]*data.Transaction{}
	txs[hexEncodedTxHash] = tx

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.WriteLogIdentifier),
	}
	args := &argsProcessEvent{
		timestamp:        1234,
		event:            event,
		logAddress:       []byte("contract"),
		txs:              txs,
		txHashHexEncoded: hexEncodedTxHash,
	}

	informativeLogsProc := newInformativeLogsProcessor()

	res := informativeLogsProc.processEvent(args)

	require.Equal(t, transaction.TxStatusSuccess.String(), tx.Status)
	require.True(t, res.processed)
}

func TestInformativeLogsProcessorSignalError(t *testing.T) {
	t.Parallel()

	tx := &data.Transaction{
		GasLimit: 200000,
		GasPrice: 100000,
		Data:     []byte("callMe"),
	}

	hexEncodedTxHash := "01020304"
	txs := map[string]*data.Transaction{}
	txs[hexEncodedTxHash] = tx

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.SignalErrorOperation),
	}
	args := &argsProcessEvent{
		timestamp:        1234,
		event:            event,
		logAddress:       []byte("contract"),
		txs:              txs,
		txHashHexEncoded: hexEncodedTxHash,
	}

	informativeLogsProc := newInformativeLogsProcessor()

	res := informativeLogsProc.processEvent(args)

	require.Equal(t, transaction.TxStatusFail.String(), tx.Status)
	require.True(t, tx.ErrorEvent)
	require.Equal(t, true, res.processed)
}

func TestInformativeLogsProcessorCompletedEvent(t *testing.T) {
	t.Parallel()

	tx := &data.Transaction{
		GasLimit: 200000,
		GasPrice: 100000,
		Data:     []byte("callMe"),
	}

	hexEncodedTxHash := "01020304"
	txs := map[string]*data.Transaction{}
	txs[hexEncodedTxHash] = tx

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.CompletedTxEventIdentifier),
	}
	args := &argsProcessEvent{
		timestamp:        1234,
		event:            event,
		logAddress:       []byte("contract"),
		txs:              txs,
		txHashHexEncoded: hexEncodedTxHash,
	}

	informativeLogsProc := newInformativeLogsProcessor()

	res := informativeLogsProc.processEvent(args)

	require.True(t, tx.CompletedEvent)
	require.Equal(t, true, res.processed)
}

func TestInformativeLogsProcessorLogsGeneratedByScrsSignalError(t *testing.T) {
	t.Parallel()

	txHash := "txHash"
	scrHash := "scrHash"
	scr := &data.ScResult{
		OriginalTxHash: txHash,
	}
	scrs := make(map[string]*data.ScResult)
	scrs[scrHash] = scr

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.SignalErrorOperation),
	}

	txStatusProc := newTxHashStatusInfoProcessor()
	args := &argsProcessEvent{
		timestamp:            1234,
		event:                event,
		logAddress:           []byte("contract"),
		scrs:                 scrs,
		txHashHexEncoded:     scrHash,
		txHashStatusInfoProc: txStatusProc,
	}

	informativeLogsProc := newInformativeLogsProcessor()
	res := informativeLogsProc.processEvent(args)
	require.True(t, res.processed)

	require.Equal(t, &outport.StatusInfo{
		Status:     transaction.TxStatusFail.String(),
		ErrorEvent: true,
	}, txStatusProc.getAllRecords()[txHash])
}

func TestInformativeLogsProcessorLogsGeneratedByScrsCompletedEvent(t *testing.T) {
	t.Parallel()

	txHash := "txHash"
	scrHash := "scrHash"
	scr := &data.ScResult{
		OriginalTxHash: txHash,
	}
	scrs := make(map[string]*data.ScResult)
	scrs[scrHash] = scr

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.CompletedTxEventIdentifier),
	}

	txStatusProc := newTxHashStatusInfoProcessor()
	args := &argsProcessEvent{
		timestamp:            1234,
		event:                event,
		logAddress:           []byte("contract"),
		scrs:                 scrs,
		txHashHexEncoded:     scrHash,
		txHashStatusInfoProc: txStatusProc,
	}

	informativeLogsProc := newInformativeLogsProcessor()
	res := informativeLogsProc.processEvent(args)
	require.True(t, res.processed)

	require.Equal(t, &outport.StatusInfo{
		CompletedEvent: true,
	}, txStatusProc.getAllRecords()[txHash])
}

func TestInformativeLogsProcessorLogsGeneratedByScrNotFoundInMap(t *testing.T) {
	t.Parallel()

	scrHash := "scrHash"

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.CompletedTxEventIdentifier),
	}

	txStatusProc := newTxHashStatusInfoProcessor()
	args := &argsProcessEvent{
		timestamp:            1234,
		event:                event,
		logAddress:           []byte("contract"),
		txHashHexEncoded:     scrHash,
		txHashStatusInfoProc: txStatusProc,
	}

	informativeLogsProc := newInformativeLogsProcessor()
	res := informativeLogsProc.processEvent(args)
	require.True(t, res.processed)
}
