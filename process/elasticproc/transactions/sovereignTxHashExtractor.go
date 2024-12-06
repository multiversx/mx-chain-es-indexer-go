package transactions

import (
	coreData "github.com/multiversx/mx-chain-core-go/data"
)

type sovereignTxHashExtractor struct{}

// NewSovereignTxHashExtractor creates a new sovereign tx hash extractor
func NewSovereignTxHashExtractor() *sovereignTxHashExtractor {
	return &sovereignTxHashExtractor{}
}

// ExtractExecutedTxHashes returns directly the provided mini block tx hashes
func (the *sovereignTxHashExtractor) ExtractExecutedTxHashes(_ int, mbTxHashes [][]byte, _ coreData.HeaderHandler) [][]byte {
	return mbTxHashes
}

// IsInterfaceNil returns true if there is no value under the interface
func (the *sovereignTxHashExtractor) IsInterfaceNil() bool {
	return the == nil
}
