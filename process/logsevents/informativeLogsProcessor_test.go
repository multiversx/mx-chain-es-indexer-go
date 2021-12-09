package logsevents

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestInformativeShouldIgnoreLog(t *testing.T) {
	feeHandler := &mock.EconomicsHandlerMock{}
	informativeLogsProc := newInformativeLogsProcessor(feeHandler)

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
		Identifier: []byte(writeLogOperation),
	}
	args := &argsProcessEvent{
		timestamp:        1234,
		event:            event,
		logAddress:       []byte("contract"),
		txs:              txs,
		txHashHexEncoded: hexEncodedTxHash,
	}

	feeHandler := &mock.EconomicsHandlerMock{}
	informativeLogsProc := newInformativeLogsProcessor(feeHandler)

	res := informativeLogsProc.processEvent(args)

	require.Equal(t, transaction.TxStatusSuccess.String(), tx.Status)
	require.Equal(t, tx.GasLimit, tx.GasUsed)
	require.Equal(t, "7083500000", tx.Fee)
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
		Identifier: []byte(signalErrorOperation),
	}
	args := &argsProcessEvent{
		timestamp:        1234,
		event:            event,
		logAddress:       []byte("contract"),
		txs:              txs,
		txHashHexEncoded: hexEncodedTxHash,
	}

	feeHandler := &mock.EconomicsHandlerMock{}
	informativeLogsProc := newInformativeLogsProcessor(feeHandler)

	res := informativeLogsProc.processEvent(args)

	require.Equal(t, transaction.TxStatusFail.String(), tx.Status)
	require.Equal(t, tx.GasLimit, tx.GasUsed)
	require.Equal(t, "6041000000", tx.Fee)
	require.Equal(t, true, res.processed)
}
