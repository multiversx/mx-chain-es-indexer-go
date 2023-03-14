package workItems

import "github.com/multiversx/mx-chain-core-go/data/outport"

type itemValidators struct {
	indexer           saveValidatorsIndexer
	validatorsPubKeys *outport.ValidatorsPubKeys
}

// NewItemValidators will create a new instance of itemValidators
func NewItemValidators(
	indexer saveValidatorsIndexer,
	validatorsPubKeys *outport.ValidatorsPubKeys,
) WorkItemHandler {
	return &itemValidators{
		indexer:           indexer,
		validatorsPubKeys: validatorsPubKeys,
	}
}

// Save will save information about validators
func (wiv *itemValidators) Save() error {
	return wiv.indexer.SaveShardValidatorsPubKeys(wiv.validatorsPubKeys)
}

// IsInterfaceNil returns true if there is no value under the interface
func (wiv *itemValidators) IsInterfaceNil() bool {
	return wiv == nil
}
