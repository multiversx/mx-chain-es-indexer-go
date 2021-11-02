package operations

import (
	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
)

type operationsProcessor struct {
	shardCoordinator indexer.ShardCoordinator
}

// NewOperationsProcessor will create a new instance of operationsProcessor
func NewOperationsProcessor(shardCoordinator indexer.ShardCoordinator) (*operationsProcessor, error) {
	return &operationsProcessor{
		shardCoordinator: shardCoordinator,
	}, nil
}

func (op *operationsProcessor) ProcessTransactionsAndSCRS(txs []*data.Transaction, scrs []*data.ScResult) {
	for _, tx := range txs {
		tx.Logs = nil
		tx.SmartContractResults = nil
		tx.Type = string(transaction.TxTypeNormal)
	}

	// TODO check if need to add token identifier and value in case of  ESDT scr
	for _, scr := range scrs {
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
