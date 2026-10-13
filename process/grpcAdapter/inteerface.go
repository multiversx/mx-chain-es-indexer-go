package grpcAdapter

import "github.com/multiversx/mx-chain-core-go/data/outport"

// DataIndexer dines what a data indexer should do
type DataIndexer interface {
	SaveBlock(outportBlock *outport.OutportBlock) error
	RevertIndexedBlock(blockData *outport.BlockData) error
	SaveRoundsInfo(roundsInfos *outport.RoundsInfo) error
	SaveValidatorsPubKeys(validatorsPubKeys *outport.ValidatorsPubKeys) error
	SaveValidatorsRating(ratingData *outport.ValidatorsRating) error
	SaveAccounts(accountsData *outport.Accounts) error
	FinalizedBlock(finalizedBlock *outport.FinalizedBlock) error
	SetCurrentSettings(settings outport.OutportConfig) error
	Close() error
	IsInterfaceNil() bool
}
