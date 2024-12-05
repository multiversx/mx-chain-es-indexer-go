package transactions

import (
	coreData "github.com/multiversx/mx-chain-core-go/data"
)

type txHashExtractor struct{}

// NewTxHashExtractor creates a new tx hash extractor
func NewTxHashExtractor() *txHashExtractor {
	return &txHashExtractor{}
}

// ExtractExecutedTxHashes returns executed tx hashes
func (the *txHashExtractor) ExtractExecutedTxHashes(mbIndex int, mbTxHashes [][]byte, header coreData.HeaderHandler) [][]byte {
	miniblockHeaders := header.GetMiniBlockHeaderHandlers()
	if len(miniblockHeaders) <= mbIndex {
		return mbTxHashes
	}

	firstProcessed := miniblockHeaders[mbIndex].GetIndexOfFirstTxProcessed()
	lastProcessed := miniblockHeaders[mbIndex].GetIndexOfLastTxProcessed()

	executedTxHashes := make([][]byte, 0)
	for txIndex, txHash := range mbTxHashes {
		if int32(txIndex) < firstProcessed || int32(txIndex) > lastProcessed {
			continue
		}

		executedTxHashes = append(executedTxHashes, txHash)
	}

	return executedTxHashes
}

// IsInterfaceNil returns true if there is no value under the interface
func (the *txHashExtractor) IsInterfaceNil() bool {
	return the == nil
}
