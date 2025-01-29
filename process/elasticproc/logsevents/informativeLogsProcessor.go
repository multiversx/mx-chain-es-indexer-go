package logsevents

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

type informativeLogsProcessor struct {
	operations map[string]struct{}
}

func newInformativeLogsProcessor() *informativeLogsProcessor {
	return &informativeLogsProcessor{
		operations: map[string]struct{}{
			core.WriteLogIdentifier:         {},
			core.SignalErrorOperation:       {},
			core.CompletedTxEventIdentifier: {},
			core.InternalVMErrorsOperation:  {},
		},
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
		return processEventNoTx(args)
	}

	switch identifier {
	case core.CompletedTxEventIdentifier:
		{
			tx.CompletedEvent = true
		}
	case core.WriteLogIdentifier:
		{
			tx.Status = transaction.TxStatusSuccess.String()
		}
	case core.SignalErrorOperation, core.InternalVMErrorsOperation:
		{
			tx.Status = transaction.TxStatusFail.String()
			tx.ErrorEvent = true
		}
	}

	return argOutputProcessEvent{
		processed: true,
	}
}

func processEventNoTx(args *argsProcessEvent) argOutputProcessEvent {
	scr, ok := args.scrs[args.txHashHexEncoded]
	if !ok {
		return argOutputProcessEvent{
			processed: true,
		}
	}
	if scr.OriginalTxHash == "" {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	record := &outport.StatusInfo{}
	switch string(args.event.GetIdentifier()) {
	case core.CompletedTxEventIdentifier:
		{
			record.CompletedEvent = true
			args.txHashStatusInfoProc.addRecord(scr.OriginalTxHash, record)
		}
	case core.SignalErrorOperation, core.InternalVMErrorsOperation:
		{
			record.Status = transaction.TxStatusFail.String()
			record.ErrorEvent = true
			args.txHashStatusInfoProc.addRecord(scr.OriginalTxHash, record)
		}
	}

	return argOutputProcessEvent{
		processed: true,
	}
}
