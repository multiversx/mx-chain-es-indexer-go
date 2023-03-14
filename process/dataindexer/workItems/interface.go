package workItems

import (
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
)

// WorkItemHandler defines the interface for item that needs to be saved in elasticsearch database
type WorkItemHandler interface {
	Save() error
	IsInterfaceNil() bool
}

type saveBlockIndexer interface {
	SaveHeader(outportBlockWithHeader *outport.OutportBlockWithHeader) error
	SaveMiniblocks(header coreData.HeaderHandler, body *block.Body) error
	SaveTransactions(outportBlockWithHeader *outport.OutportBlockWithHeader) error
}

type saveRatingIndexer interface {
	SaveValidatorsRating(ratingData *outport.ValidatorsRating) error
}

type removeIndexer interface {
	RemoveHeader(header coreData.HeaderHandler) error
	RemoveMiniblocks(header coreData.HeaderHandler, body *block.Body) error
	RemoveTransactions(header coreData.HeaderHandler, body *block.Body) error
	RemoveAccountsESDT(headerTimestamp uint64, shardID uint32) error
}

type saveRounds interface {
	SaveRoundsInfo(rounds *outport.RoundsInfo) error
}

type saveValidatorsIndexer interface {
	SaveShardValidatorsPubKeys(validatorsPubKeys *outport.ValidatorsPubKeys) error
}

type saveAccountsIndexer interface {
	SaveAccounts(accounts *outport.Accounts) error
}
