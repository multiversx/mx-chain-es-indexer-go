package workItems

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// ErrBodyTypeAssertion signals that body type assertion failed
var ErrBodyTypeAssertion = errors.New("elasticsearch - body type assertion failed")

var log = logger.GetOrCreate("indexer/workItems")

type itemBlock struct {
	indexer       saveBlockIndexer
	marshalizer   marshal.Marshalizer
	argsSaveBlock *outport.ArgsSaveBlockData
}

// NewItemBlock will create a new instance of ItemBlock
func NewItemBlock(
	indexer saveBlockIndexer,
	marshalizer marshal.Marshalizer,
	args *outport.ArgsSaveBlockData,
) WorkItemHandler {
	return &itemBlock{
		indexer:       indexer,
		marshalizer:   marshalizer,
		argsSaveBlock: args,
	}
}

// Save will prepare and save a block item in elasticsearch database
func (wib *itemBlock) Save() error {
	if check.IfNil(wib.argsSaveBlock.Header) {
		log.Warn("nil header provided when trying to index block, will skip")
		return nil
	}

	defer func(startTime time.Time) {
		log.Debug("wib.SaveBlockData duration", "time", time.Since(startTime))
	}(time.Now())

	log.Debug("indexer: starting indexing block",
		"hash", wib.argsSaveBlock.HeaderHash,
		"nonce", wib.argsSaveBlock.Header.GetNonce())

	body, ok := wib.argsSaveBlock.Body.(*block.Body)
	if !ok {
		return fmt.Errorf("%w when trying body assertion, block hash %s, nonce %d",
			ErrBodyTypeAssertion, wib.argsSaveBlock.HeaderHash, wib.argsSaveBlock.Header.GetNonce())
	}

	if wib.argsSaveBlock.TransactionsPool == nil {
		wib.argsSaveBlock.TransactionsPool = &outport.Pool{}
	}

	txsSizeInBytes := ComputeSizeOfTxs(wib.marshalizer, wib.argsSaveBlock.TransactionsPool)
	err := wib.indexer.SaveHeader(
		wib.argsSaveBlock.HeaderHash,
		wib.argsSaveBlock.Header,
		wib.argsSaveBlock.SignersIndexes,
		body,
		wib.argsSaveBlock.NotarizedHeadersHashes,
		wib.argsSaveBlock.HeaderGasConsumption,
		txsSizeInBytes,
		wib.argsSaveBlock.TransactionsPool,
	)
	if err != nil {
		return fmt.Errorf("%w when saving header block, hash %s, nonce %d",
			err, hex.EncodeToString(wib.argsSaveBlock.HeaderHash), wib.argsSaveBlock.Header.GetNonce())
	}

	if len(body.MiniBlocks) == 0 {
		return nil
	}

	err = wib.indexer.SaveMiniblocks(wib.argsSaveBlock.Header, body)
	if err != nil {
		return fmt.Errorf("%w when saving miniblocks, block hash %s, nonce %d",
			err, hex.EncodeToString(wib.argsSaveBlock.HeaderHash), wib.argsSaveBlock.Header.GetNonce())
	}

	err = wib.indexer.SaveTransactions(body, wib.argsSaveBlock.Header, wib.argsSaveBlock.TransactionsPool, wib.argsSaveBlock.AlteredAccounts, wib.argsSaveBlock.IsImportDB, wib.argsSaveBlock.NumberOfShards)
	if err != nil {
		return fmt.Errorf("%w when saving transactions, block hash %s, nonce %d",
			err, hex.EncodeToString(wib.argsSaveBlock.HeaderHash), wib.argsSaveBlock.Header.GetNonce())
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (wib *itemBlock) IsInterfaceNil() bool {
	return wib == nil
}

// ComputeSizeOfTxs will compute size of transactions in bytes
func ComputeSizeOfTxs(marshalizer marshal.Marshalizer, pool *outport.Pool) int {
	sizeTxs := 0
	sizeTxs += computeSizeOfMapTxs(marshalizer, pool.Txs)
	sizeTxs += computeSizeOfMapTxs(marshalizer, pool.Scrs)
	sizeTxs += computeSizeOfMapTxs(marshalizer, pool.Invalid)
	sizeTxs += computeSizeOfMapTxs(marshalizer, pool.Rewards)
	sizeTxs += computeSizeOfMapTxs(marshalizer, pool.Receipts)

	return sizeTxs
}

func computeSizeOfMapTxs(marshalizer marshal.Marshalizer, mapTxs map[string]data.TransactionHandlerWithGasUsedAndFee) int {
	txsSize := 0
	for _, tx := range mapTxs {
		txsSize += computeTxSize(marshalizer, tx.GetTxHandler())
	}

	return txsSize
}

func computeTxSize(marshalizer marshal.Marshalizer, tx data.TransactionHandler) int {
	txBytes, err := marshalizer.Marshal(tx)
	if err != nil {
		log.Debug("itemBlock.computeTxSize", "error", err)
		return 0
	}

	return len(txBytes)
}
