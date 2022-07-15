package logsevents

import (
	"math/big"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
)

const (
	writeLogOperation    = "writeLog"
	signalErrorOperation = "signalError"
)

type informativeLogsProcessor struct {
	operations      map[string]struct{}
	txFeeCalculator indexer.FeesProcessorHandler
}

func newInformativeLogsProcessor(txFeeCalculator indexer.FeesProcessorHandler) *informativeLogsProcessor {
	return &informativeLogsProcessor{
		operations: map[string]struct{}{
			writeLogOperation:    {},
			signalErrorOperation: {},
		},
		txFeeCalculator: txFeeCalculator,
	}
}

func (ilp *informativeLogsProcessor) processEvent(args *argsProcessEvent) argOutputProcessEvent {
	identifier := string(args.event.GetIdentifier())
	_, ok := ilp.operations[identifier]
	if !ok {
		return argOutputProcessEvent{}
	}

	tx, ok := args.txs[args.txHashHexEncoded]
	if !ok {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	switch identifier {
	case writeLogOperation:
		{
			if !tx.HadRefund {
				gasLimit, fee := ilp.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(tx, big.NewInt(0))
				tx.GasUsed = gasLimit
				tx.Fee = fee.String()
			}

			tx.Status = transaction.TxStatusSuccess.String()
		}
	case signalErrorOperation:
		{
			tx.GasUsed = tx.GasLimit
			fee := ilp.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, tx.GasLimit)
			tx.Fee = fee.String()
			tx.Status = transaction.TxStatusFail.String()
		}
	}

	return argOutputProcessEvent{
		processed: true,
	}
}
