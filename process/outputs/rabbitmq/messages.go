package rabbitmq

import (
	"encoding/hex"
	"sort"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

// BlockInfo holds the data that identifies a block
type BlockInfo struct {
	ShardID uint32 `json:"shardID"`
	Hash    string `json:"hash"`
	Round   uint64 `json:"round"`
}

// BlockMessage is the message published for a block: it holds the operations (transactions and smart contract results)
// and the events of the block, in the same format as the documents from the operations and events indices
type BlockMessage struct {
	BlockInfo
	Operations []any          `json:"operations"`
	Events     []*eventWithID `json:"events"`
}

// RevertMessage is the message published when a block is reverted
type RevertMessage struct {
	BlockInfo
	// Operations holds the hashes of the transactions and smart contract results of the reverted block
	Operations []string `json:"operations"`
}

// the hash of a transaction or of a smart contract result and the id of an event are the ids of the documents
// from the indices, they are not part of the documents, so they are added in the message

type transactionWithHash struct {
	Hash string `json:"hash"`
	*data.Transaction
}

type scResultWithHash struct {
	Hash string `json:"hash"`
	*data.ScResult
}

type eventWithID struct {
	ID string `json:"id"`
	*data.LogEvent
}

// newBlockMessage creates the message for the provided block unit.
// If the block has no operations and no events, the lists from the message are empty
func newBlockMessage(unit *data.PreparedBlockUnit) *BlockMessage {
	numOperations := len(unit.OperationsTxs) + len(unit.OperationsSCRs)
	numEvents := 0
	if unit.Logs != nil {
		numEvents = len(unit.Logs.DBEvents)
	}

	message := &BlockMessage{
		BlockInfo: BlockInfo{
			ShardID: unit.HeaderData.ShardID,
			Hash:    hex.EncodeToString(unit.BlockHash),
			Round:   unit.HeaderData.Round,
		},
		Operations: make([]any, 0, numOperations),
		Events:     make([]*eventWithID, 0, numEvents),
	}

	var txHashStatusInfo map[string]*outport.StatusInfo
	if unit.Logs != nil {
		txHashStatusInfo = unit.Logs.TxHashStatusInfo
	}

	for _, tx := range sortedTransactions(unit.OperationsTxs) {
		message.Operations = append(message.Operations, &transactionWithHash{
			Hash:        tx.Hash,
			Transaction: applyStatusInfo(tx, txHashStatusInfo[tx.Hash]),
		})
	}
	for _, scr := range sortedSCRs(unit.OperationsSCRs) {
		message.Operations = append(message.Operations, &scResultWithHash{
			Hash:     scr.Hash,
			ScResult: scr,
		})
	}

	if unit.Logs != nil {
		for _, event := range unit.Logs.DBEvents {
			message.Events = append(message.Events, &eventWithID{
				ID:       event.ID,
				LogEvent: event,
			})
		}
	}

	return message
}

func (bm *BlockMessage) isEmpty() bool {
	return len(bm.Operations) == 0 && len(bm.Events) == 0
}

// applyStatusInfo returns a copy of the transaction that has the status changed by the logs of the same block.
// In the indices this status is set with an update of the document, the prepared transaction does not have it
func applyStatusInfo(tx *data.Transaction, statusInfo *outport.StatusInfo) *data.Transaction {
	if statusInfo == nil {
		return tx
	}

	copiedTx := *tx
	if statusInfo.Status != "" {
		copiedTx.Status = statusInfo.Status
	}
	if statusInfo.CompletedEvent {
		copiedTx.CompletedEvent = true
	}
	if statusInfo.ErrorEvent {
		copiedTx.ErrorEvent = true
	}

	return &copiedTx
}

// the prepared transactions and smart contract results come from maps, they are sorted in order to have
// the same message every time the same block is published
func sortedTransactions(txs []*data.Transaction) []*data.Transaction {
	sorted := make([]*data.Transaction, len(txs))
	copy(sorted, txs)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].ExecutionOrder != sorted[j].ExecutionOrder {
			return sorted[i].ExecutionOrder < sorted[j].ExecutionOrder
		}

		return sorted[i].Hash < sorted[j].Hash
	})

	return sorted
}

func sortedSCRs(scrs []*data.ScResult) []*data.ScResult {
	sorted := make([]*data.ScResult, len(scrs))
	copy(sorted, scrs)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].ExecutionOrder != sorted[j].ExecutionOrder {
			return sorted[i].ExecutionOrder < sorted[j].ExecutionOrder
		}

		return sorted[i].Hash < sorted[j].Hash
	})

	return sorted
}

func newRevertMessage(revertedBlock *data.RevertedBlock) *RevertMessage {
	operations := make([]string, 0, len(revertedBlock.TxHashes)+len(revertedBlock.ScrHashes))
	operations = append(operations, revertedBlock.TxHashes...)
	operations = append(operations, revertedBlock.ScrHashes...)

	return &RevertMessage{
		BlockInfo: BlockInfo{
			ShardID: revertedBlock.Header.GetShardID(),
			Hash:    hex.EncodeToString(revertedBlock.BlockData.HeaderHash),
			Round:   revertedBlock.Header.GetRound(),
		},
		Operations: operations,
	}
}
