package workItems

import (
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

// WorkItemHandler defines the interface for item that needs to be saved in elasticsearch database
type WorkItemHandler interface {
	Save() error
	IsInterfaceNil() bool
}

type saveBlockIndexer interface {
	SaveHeader(
		headerHash []byte,
		header coreData.HeaderHandler,
		signersIndexes []uint64,
		body *block.Body,
		notarizedHeadersHashes []string,
		gasConsumptionData outport.HeaderGasConsumption,
		txsSize int,
		pool *outport.Pool,
	) error
	SaveMiniblocks(header coreData.HeaderHandler, body *block.Body) error
	SaveTransactions(body *block.Body, header coreData.HeaderHandler, pool *outport.Pool, coreAlteredAccounts map[string]*outport.AlteredAccount, isImportDB bool, numOfShards uint32) error
}

type saveRatingIndexer interface {
	SaveValidatorsRating(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) error
}

type removeIndexer interface {
	RemoveHeader(header coreData.HeaderHandler) error
	RemoveMiniblocks(header coreData.HeaderHandler, body *block.Body) error
	RemoveTransactions(header coreData.HeaderHandler, body *block.Body) error
	RemoveAccountsESDT(headerTimestamp uint64, shardID uint32) error
}

type saveRounds interface {
	SaveRoundsInfo(infos []*data.RoundInfo) error
}

type saveValidatorsIndexer interface {
	SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error
}

type saveAccountsIndexer interface {
	SaveAccounts(blockTimestamp uint64, accounts []*data.Account, shardID uint32) error
}
