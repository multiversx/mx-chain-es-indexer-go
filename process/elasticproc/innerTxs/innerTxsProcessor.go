package innerTxs

import (
	"encoding/hex"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

// InnerTxType is the type for an inner transaction
const InnerTxType = "innerTx"

type innerTxsProcessor struct {
	hasher hashing.Hasher
}

// NewInnerTxsProcessor will create a new instance of inner transactions processor
func NewInnerTxsProcessor(hasher hashing.Hasher) (*innerTxsProcessor, error) {
	if check.IfNil(hasher) {
		return nil, core.ErrNilHasher
	}

	return &innerTxsProcessor{
		hasher: hasher,
	}, nil
}

// ExtractInnerTxs will extract the inner transactions from the transaction array
func (ip *innerTxsProcessor) ExtractInnerTxs(
	txs []*data.Transaction,
) []*data.InnerTransaction {
	innerTxs := make([]*data.InnerTransaction, 0)
	for _, tx := range txs {
		for _, innerTx := range tx.InnerTransactions {
			innerTxCopy := *innerTx

			id := ip.hasher.Compute(innerTxCopy.Hash + innerTxCopy.RelayedTxHash)
			innerTxCopy.ID = hex.EncodeToString(id)
			innerTxCopy.Type = InnerTxType
			innerTxCopy.RelayedTxHash = tx.Hash
			innerTxs = append(innerTxs, &innerTxCopy)
		}
	}

	return innerTxs
}
