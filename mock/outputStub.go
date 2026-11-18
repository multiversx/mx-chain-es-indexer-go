package mock

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

// OutputStub -
type OutputStub struct {
	SaveHeaderCalled                 func(outportBlockWithHeader *outport.OutportBlockWithHeader) error
	SaveMiniblocksCalled             func(outportBlockWithHeader *outport.OutportBlockWithHeader) error
	SaveTransactionsCalled           func(block *data.PreparedBlock) error
	RevertBlockCalled                func(revertedBlock *data.RevertedBlock) error
	SaveValidatorsRatingCalled       func(validatorsRating *outport.ValidatorsRating) error
	SaveRoundsInfoCalled             func(infos *outport.RoundsInfo) error
	SaveShardValidatorsPubKeysCalled func(validators *outport.ValidatorsPubKeys) error
	SaveAccountsCalled               func(accountsData *outport.Accounts) error
	CloseCalled                      func() error
}

// SaveHeader -
func (os *OutputStub) SaveHeader(obh *outport.OutportBlockWithHeader) error {
	if os.SaveHeaderCalled != nil {
		return os.SaveHeaderCalled(obh)
	}
	return nil
}

// SaveMiniblocks -
func (os *OutputStub) SaveMiniblocks(obh *outport.OutportBlockWithHeader) error {
	if os.SaveMiniblocksCalled != nil {
		return os.SaveMiniblocksCalled(obh)
	}
	return nil
}

// SaveTransactions -
func (os *OutputStub) SaveTransactions(block *data.PreparedBlock) error {
	if os.SaveTransactionsCalled != nil {
		return os.SaveTransactionsCalled(block)
	}
	return nil
}

// RevertBlock -
func (os *OutputStub) RevertBlock(revertedBlock *data.RevertedBlock) error {
	if os.RevertBlockCalled != nil {
		return os.RevertBlockCalled(revertedBlock)
	}
	return nil
}

// SaveValidatorsRating -
func (os *OutputStub) SaveValidatorsRating(validatorsRating *outport.ValidatorsRating) error {
	if os.SaveValidatorsRatingCalled != nil {
		return os.SaveValidatorsRatingCalled(validatorsRating)
	}
	return nil
}

// SaveRoundsInfo -
func (os *OutputStub) SaveRoundsInfo(infos *outport.RoundsInfo) error {
	if os.SaveRoundsInfoCalled != nil {
		return os.SaveRoundsInfoCalled(infos)
	}
	return nil
}

// SaveShardValidatorsPubKeys -
func (os *OutputStub) SaveShardValidatorsPubKeys(validators *outport.ValidatorsPubKeys) error {
	if os.SaveShardValidatorsPubKeysCalled != nil {
		return os.SaveShardValidatorsPubKeysCalled(validators)
	}
	return nil
}

// SaveAccounts -
func (os *OutputStub) SaveAccounts(accountsData *outport.Accounts) error {
	if os.SaveAccountsCalled != nil {
		return os.SaveAccountsCalled(accountsData)
	}
	return nil
}

// Close -
func (os *OutputStub) Close() error {
	if os.CloseCalled != nil {
		return os.CloseCalled()
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (os *OutputStub) IsInterfaceNil() bool {
	return os == nil
}
