package mock

import (
	coreData "github.com/multiversx/mx-chain-core-go/data"
)

// TxHashExtractorMock -
type TxHashExtractorMock struct {
	ExtractExecutedTxHashesCalled func(mbIndex int, mbTxHashes [][]byte, header coreData.HeaderHandler) [][]byte
}

// ExtractExecutedTxHashes -
func (the *TxHashExtractorMock) ExtractExecutedTxHashes(mbIndex int, mbTxHashes [][]byte, header coreData.HeaderHandler) [][]byte {
	if the.ExtractExecutedTxHashesCalled != nil {
		return the.ExtractExecutedTxHashesCalled(mbIndex, mbTxHashes, header)
	}

	return make([][]byte, 0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (the *TxHashExtractorMock) IsInterfaceNil() bool {
	return the == nil
}
