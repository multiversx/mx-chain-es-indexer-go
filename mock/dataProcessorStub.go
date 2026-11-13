package mock

import (
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/outport"
)

// DataProcessorStub -
type DataProcessorStub struct {
	SaveHeaderCalled                 func(outportBlockWithHeader *outport.OutportBlockWithHeader) error
	SaveMiniblocksCalled             func(outportBlockWithHeader *outport.OutportBlockWithHeader) error
	SaveTransactionsCalled           func(outportBlockWithHeader *outport.OutportBlockWithHeader) error
	RevertBlockCalled                func(header coreData.HeaderHandler, blockData *outport.BlockData) error
	SaveValidatorsRatingCalled       func(validatorsRating *outport.ValidatorsRating) error
	SaveRoundsInfoCalled             func(infos *outport.RoundsInfo) error
	SaveShardValidatorsPubKeysCalled func(validators *outport.ValidatorsPubKeys) error
	SaveAccountsCalled               func(accountsData *outport.Accounts) error
	SetOutportConfigCalled           func(cfg outport.OutportConfig) error
	CloseCalled                      func() error
}

// SaveHeader -
func (dps *DataProcessorStub) SaveHeader(obh *outport.OutportBlockWithHeader) error {
	if dps.SaveHeaderCalled != nil {
		return dps.SaveHeaderCalled(obh)
	}
	return nil
}

// SaveMiniblocks -
func (dps *DataProcessorStub) SaveMiniblocks(obh *outport.OutportBlockWithHeader) error {
	if dps.SaveMiniblocksCalled != nil {
		return dps.SaveMiniblocksCalled(obh)
	}
	return nil
}

// SaveTransactions -
func (dps *DataProcessorStub) SaveTransactions(obh *outport.OutportBlockWithHeader) error {
	if dps.SaveTransactionsCalled != nil {
		return dps.SaveTransactionsCalled(obh)
	}
	return nil
}

// RevertBlock -
func (dps *DataProcessorStub) RevertBlock(header coreData.HeaderHandler, blockData *outport.BlockData) error {
	if dps.RevertBlockCalled != nil {
		return dps.RevertBlockCalled(header, blockData)
	}
	return nil
}

// SaveValidatorsRating -
func (dps *DataProcessorStub) SaveValidatorsRating(validatorsRating *outport.ValidatorsRating) error {
	if dps.SaveValidatorsRatingCalled != nil {
		return dps.SaveValidatorsRatingCalled(validatorsRating)
	}
	return nil
}

// SaveRoundsInfo -
func (dps *DataProcessorStub) SaveRoundsInfo(infos *outport.RoundsInfo) error {
	if dps.SaveRoundsInfoCalled != nil {
		return dps.SaveRoundsInfoCalled(infos)
	}
	return nil
}

// SaveShardValidatorsPubKeys -
func (dps *DataProcessorStub) SaveShardValidatorsPubKeys(validators *outport.ValidatorsPubKeys) error {
	if dps.SaveShardValidatorsPubKeysCalled != nil {
		return dps.SaveShardValidatorsPubKeysCalled(validators)
	}
	return nil
}

// SaveAccounts -
func (dps *DataProcessorStub) SaveAccounts(accountsData *outport.Accounts) error {
	if dps.SaveAccountsCalled != nil {
		return dps.SaveAccountsCalled(accountsData)
	}
	return nil
}

// SetOutportConfig -
func (dps *DataProcessorStub) SetOutportConfig(cfg outport.OutportConfig) error {
	if dps.SetOutportConfigCalled != nil {
		return dps.SetOutportConfigCalled(cfg)
	}
	return nil
}

// Close -
func (dps *DataProcessorStub) Close() error {
	if dps.CloseCalled != nil {
		return dps.CloseCalled()
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dps *DataProcessorStub) IsInterfaceNil() bool {
	return dps == nil
}
