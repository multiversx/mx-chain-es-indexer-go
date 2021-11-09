package indexer

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
)

// NilIndexer will be used when an Indexer is required, but another one isn't necessary or available
type NilIndexer struct {
}

// NewNilIndexer will return a Nil indexer
func NewNilIndexer() *NilIndexer {
	return new(NilIndexer)
}

// SaveBlock returns nil
func (ni *NilIndexer) SaveBlock(_ *indexer.ArgsSaveBlockData) error {
	return nil
}

// RevertIndexedBlock returns nil
func (ni *NilIndexer) RevertIndexedBlock(_ data.HeaderHandler, _ data.BodyHandler) error {
	return nil
}

// SaveRoundsInfo will do nothing
func (ni *NilIndexer) SaveRoundsInfo(_ []*indexer.RoundInfo) {
}

// SaveValidatorsRating -
func (ni *NilIndexer) SaveValidatorsRating(_ string, _ []*indexer.ValidatorRatingInfo) {
}

// SaveValidatorsPubKeys will do nothing
func (ni *NilIndexer) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) {
}

// SaveAccounts won't do anything as this is a nil implementation
func (ni *NilIndexer) SaveAccounts(_ uint64, _ []data.UserAccountHandler) {
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
