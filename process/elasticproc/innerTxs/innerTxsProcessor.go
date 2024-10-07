package innerTxs

import "github.com/multiversx/mx-chain-es-indexer-go/data"

// InnerTxType is the type for an inner transaction
const InnerTxType = "innerTx"

type innerTxsProcessor struct {
}

// NewInnerTxsProcessor will create a new instance of inner transactions processor
func NewInnerTxsProcessor() *innerTxsProcessor {
	return &innerTxsProcessor{}
}

// ExtractInnerTxs will extract the inner transactions from the transaction array
func (ip *innerTxsProcessor) ExtractInnerTxs(
	txs []*data.Transaction,
) []*data.InnerTransaction {
	innerTxs := make([]*data.InnerTransaction, 0)
	for _, tx := range txs {
		for _, innerTx := range tx.InnerTransactions {
			innerTxCopy := *innerTx

			innerTxCopy.Type = InnerTxType
			innerTxCopy.RelayedTxHash = tx.Hash
			innerTxs = append(innerTxs, &innerTxCopy)
		}
	}

	return innerTxs
}
