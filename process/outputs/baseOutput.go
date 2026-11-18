package outputs

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

// BaseOutput implements all the methods of the Output interface and does nothing.
// It should be embedded by the outputs that need only a part of the data
type BaseOutput struct {
}

// SaveHeader does nothing
func (bo *BaseOutput) SaveHeader(_ *outport.OutportBlockWithHeader) error {
	return nil
}

// SaveMiniblocks does nothing
func (bo *BaseOutput) SaveMiniblocks(_ *outport.OutportBlockWithHeader) error {
	return nil
}

// SaveTransactions does nothing
func (bo *BaseOutput) SaveTransactions(_ *data.PreparedBlock) error {
	return nil
}

// RevertBlock does nothing
func (bo *BaseOutput) RevertBlock(_ *data.RevertedBlock) error {
	return nil
}

// SaveValidatorsRating does nothing
func (bo *BaseOutput) SaveValidatorsRating(_ *outport.ValidatorsRating) error {
	return nil
}

// SaveRoundsInfo does nothing
func (bo *BaseOutput) SaveRoundsInfo(_ *outport.RoundsInfo) error {
	return nil
}

// SaveShardValidatorsPubKeys does nothing
func (bo *BaseOutput) SaveShardValidatorsPubKeys(_ *outport.ValidatorsPubKeys) error {
	return nil
}

// SaveAccounts does nothing
func (bo *BaseOutput) SaveAccounts(_ *outport.Accounts) error {
	return nil
}

// Close does nothing
func (bo *BaseOutput) Close() error {
	return nil
}
