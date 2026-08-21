package mock

import "github.com/multiversx/mx-chain-core-go/data/outport"

// DataIndexerStub -
type DataIndexerStub struct {
	SaveBlockCalled             func(outportBlock *outport.OutportBlock) error
	RevertIndexedBlockCalled    func(blockData *outport.BlockData) error
	SaveRoundsInfoCalled        func(roundsInfos *outport.RoundsInfo) error
	SaveValidatorsPubKeysCalled func(validatorsPubKeys *outport.ValidatorsPubKeys) error
	SaveValidatorsRatingCalled  func(ratingData *outport.ValidatorsRating) error
	SaveAccountsCalled          func(accountsData *outport.Accounts) error
	FinalizedBlockCalled        func(finalizedBlock *outport.FinalizedBlock) error
	SetCurrentSettingsCalled    func(settings outport.OutportConfig) error
	CloseCalled                 func() error
}

// SaveBlock -
func (s *DataIndexerStub) SaveBlock(outportBlock *outport.OutportBlock) error {
	if s.SaveBlockCalled != nil {
		return s.SaveBlockCalled(outportBlock)
	}
	return nil
}

// RevertIndexedBlock -
func (s *DataIndexerStub) RevertIndexedBlock(blockData *outport.BlockData) error {
	if s.RevertIndexedBlockCalled != nil {
		return s.RevertIndexedBlockCalled(blockData)
	}
	return nil
}

// SaveRoundsInfo -
func (s *DataIndexerStub) SaveRoundsInfo(roundsInfos *outport.RoundsInfo) error {
	if s.SaveRoundsInfoCalled != nil {
		return s.SaveRoundsInfoCalled(roundsInfos)
	}
	return nil
}

// SaveValidatorsPubKeys -
func (s *DataIndexerStub) SaveValidatorsPubKeys(validatorsPubKeys *outport.ValidatorsPubKeys) error {
	if s.SaveValidatorsPubKeysCalled != nil {
		return s.SaveValidatorsPubKeysCalled(validatorsPubKeys)
	}
	return nil
}

// SaveValidatorsRating -
func (s *DataIndexerStub) SaveValidatorsRating(ratingData *outport.ValidatorsRating) error {
	if s.SaveValidatorsRatingCalled != nil {
		return s.SaveValidatorsRatingCalled(ratingData)
	}
	return nil
}

// SaveAccounts -
func (s *DataIndexerStub) SaveAccounts(accountsData *outport.Accounts) error {
	if s.SaveAccountsCalled != nil {
		return s.SaveAccountsCalled(accountsData)
	}
	return nil
}

// FinalizedBlock -
func (s *DataIndexerStub) FinalizedBlock(finalizedBlock *outport.FinalizedBlock) error {
	if s.FinalizedBlockCalled != nil {
		return s.FinalizedBlockCalled(finalizedBlock)
	}
	return nil
}

// SetCurrentSettings -
func (s *DataIndexerStub) SetCurrentSettings(settings outport.OutportConfig) error {
	if s.SetCurrentSettingsCalled != nil {
		return s.SetCurrentSettingsCalled(settings)
	}
	return nil
}

// Close -
func (s *DataIndexerStub) Close() error {
	if s.CloseCalled != nil {
		return s.CloseCalled()
	}
	return nil
}

// IsInterfaceNil -
func (s *DataIndexerStub) IsInterfaceNil() bool {
	return s == nil
}
