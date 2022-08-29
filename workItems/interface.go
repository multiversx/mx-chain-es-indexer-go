package workItems

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
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
		gasConsumptionData indexer.HeaderGasConsumption,
		txsSize int,
	) error
	SaveMiniblocks(header coreData.HeaderHandler, body *block.Body) error
	SaveTransactions(body *block.Body, header coreData.HeaderHandler, pool *indexer.Pool, coreAlteredAccounts map[string]*indexer.AlteredAccount) error
}

type saveRatingIndexer interface {
	SaveValidatorsRating(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) error
}

type removeIndexer interface {
	RemoveHeader(header coreData.HeaderHandler) error
	RemoveMiniblocks(header coreData.HeaderHandler, body *block.Body) error
	RemoveTransactions(header coreData.HeaderHandler, body *block.Body) error
	RemoveAccountsESDT(headerTimestamp uint64) error
}

type saveRounds interface {
	SaveRoundsInfo(infos []*data.RoundInfo) error
}

type saveValidatorsIndexer interface {
	SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error
}

type saveAccountsIndexer interface {
	SaveAccounts(blockTimestamp uint64, accounts []*data.Account) error
}
