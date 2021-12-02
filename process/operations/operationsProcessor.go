package operations

import (
	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
)

type operationsProcessor struct {
	importDBMode     bool
	shardCoordinator indexer.ShardCoordinator
}

// NewOperationsProcessor will create a new instance of operationsProcessor
func NewOperationsProcessor(importDBMode bool, shardCoordinator indexer.ShardCoordinator) (*operationsProcessor, error) {
	return &operationsProcessor{
		shardCoordinator: shardCoordinator,
		importDBMode:     importDBMode,
	}, nil
}

func (op *operationsProcessor) ProcessTransactionsAndSCRS(txs []*data.Transaction, scrs []*data.ScResult) {
	for _, tx := range txs {
		tx.Logs = nil
		tx.SmartContractResults = nil
		tx.Type = string(transaction.TxTypeNormal)
	}

	// TODO check if need to add token identifier and value in case of  ESDT scr
	for idx := 0; idx < len(scrs); idx++ {
		if !op.shouldIndex(scrs[idx].ReceiverShard) {
			// remove scr from slice
			scrs = append(scrs[:idx], scrs[idx+1:]...)
		}

		scr := scrs[idx]
		scr.Logs = nil
		scr.Type = string(transaction.TxTypeUnsigned)

		selfShard := op.shardCoordinator.SelfId()
		if selfShard == scr.ReceiverShard {
			scr.Status = transaction.TxStatusSuccess.String()
		} else {
			scr.Status = transaction.TxStatusPending.String()
		}
	}
}

func (op *operationsProcessor) shouldIndex(destinationShardID uint32) bool {
	if !op.importDBMode {
		return true
	}

	return op.shardCoordinator.SelfId() == destinationShardID
}
