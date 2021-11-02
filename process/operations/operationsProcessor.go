package operations

import "github.com/ElrondNetwork/elastic-indexer-go/data"

type operationsProcessor struct {
}

// NewOperationsProcessor will create a new instance of operationsProcessor
func NewOperationsProcessor() (*operationsProcessor, error) {
	return &operationsProcessor{}, nil
}

func (op *operationsProcessor) ProcessTransactionsAndSCRS(txs []*data.Transaction, scrs []*data.ScResult) {
	for _, tx := range txs {
		tx.Logs = nil
		tx.SmartContractResults = nil
	}

	// TODO check if need to add token identifier and value in case of  ESDT scr
	for _, scr := range scrs {
		scr.Logs = nil
	}
}
