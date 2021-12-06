package operations

import (
	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
)

type operationsProcessor struct {
	importDBMode     bool
	shardCoordinator indexer.ShardCoordinator
}

// NewOperationsProcessor will create a new instance of operationsProcessor
func NewOperationsProcessor(importDBMode bool, shardCoordinator indexer.ShardCoordinator) (*operationsProcessor, error) {
	if check.IfNil(shardCoordinator) {
		return nil, indexer.ErrNilShardCoordinator
	}

	return &operationsProcessor{
		shardCoordinator: shardCoordinator,
		importDBMode:     importDBMode,
	}, nil
}

func (op *operationsProcessor) ProcessTransactionsAndSCRS(
	txs []*data.Transaction,
	scrs []*data.ScResult,
) ([]*data.Transaction, []*data.ScResult) {
	for idx, tx := range txs {
		if !op.shouldIndex(txs[idx].ReceiverShard) {
			// remove tx from slice
			txs = append(txs[:idx], txs[idx+1:]...)
			continue
		}

		tx.Logs = nil
		tx.SmartContractResults = nil
		tx.Receipt = nil
		tx.Type = string(transaction.TxTypeNormal)
	}

	// TODO check if need to add token identifier and value in case of  ESDT scr
	for idx := 0; idx < len(scrs); idx++ {
		if !op.shouldIndex(scrs[idx].ReceiverShard) {
			// remove scr from slice
			scrs = append(scrs[:idx], scrs[idx+1:]...)
			continue
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

	return txs, scrs
}

func (op *operationsProcessor) shouldIndex(destinationShardID uint32) bool {
	if !op.importDBMode {
		return true
	}

	return op.shardCoordinator.SelfId() == destinationShardID
}
