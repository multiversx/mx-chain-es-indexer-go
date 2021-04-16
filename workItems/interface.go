package workItems

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
)

// WorkItemHandler defines the interface for item that needs to be saved in elasticsearch database
type WorkItemHandler interface {
	Save() error
	IsInterfaceNil() bool
}

type saveBlockIndexer interface {
	SaveHeader(header nodeData.HeaderHandler, signersIndexes []uint64, body *block.Body, notarizedHeadersHashes []string, txsSize int) error
	SaveMiniblocks(header nodeData.HeaderHandler, body *block.Body) (map[string]bool, error)
	SaveTransactions(body *block.Body, header nodeData.HeaderHandler, pool *indexer.Pool, mbsInDb map[string]bool) error
}

type saveRatingIndexer interface {
	SaveValidatorsRating(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) error
}

type removeIndexer interface {
	RemoveHeader(header nodeData.HeaderHandler) error
	RemoveMiniblocks(header nodeData.HeaderHandler, body *block.Body) error
	RemoveTransactions(header nodeData.HeaderHandler, body *block.Body) error
}

type saveRounds interface {
	SaveRoundsInfo(infos []*data.RoundInfo) error
}

type saveTpsBenchmark interface {
	SaveShardStatistics(tpsBenchmark statistics.TPSBenchmark) error
}

type saveValidatorsIndexer interface {
	SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error
}

type saveAccountsIndexer interface {
	SaveAccounts(blockTimestamp uint64, accounts []*data.Account) error
}
