package dataindexer

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
)

// NilIndexer will be used when an Indexer is required, but another one isn't necessary or available
type NilIndexer struct {
}

// SaveBlock returns nil
func (ni *NilIndexer) SaveBlock(_ *outport.OutportBlock) error {
	return nil
}

// RevertIndexedBlock returns nil
func (ni *NilIndexer) RevertIndexedBlock(_ *outport.BlockData) error {
	return nil
}

// SaveRoundsInfo returns nil
func (ni *NilIndexer) SaveRoundsInfo(_ *outport.RoundsInfo) error {
	return nil
}

// SaveValidatorsRating returns nil
func (ni *NilIndexer) SaveValidatorsRating(_ *outport.ValidatorsRating) error {
	return nil
}

// SaveValidatorsPubKeys returns nil
func (ni *NilIndexer) SaveValidatorsPubKeys(_ *outport.ValidatorsPubKeys) error {
	return nil
}

// SaveAccounts returns nil
func (ni *NilIndexer) SaveAccounts(_ *outport.Accounts) error {
	return nil
}

// Close will do nothing
func (ni *NilIndexer) Close() error {
	return nil
}

// FinalizedBlock returns nil
func (ni *NilIndexer) FinalizedBlock(_ []byte) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ni *NilIndexer) IsInterfaceNil() bool {
	return ni == nil
}

// IsNilIndexer will return a bool value that signals if the indexer's implementation is a NilIndexer
func (ni *NilIndexer) IsNilIndexer() bool {
	return true
}
