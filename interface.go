package indexer

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/workItems"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// DispatcherHandler defines the interface for the dispatcher that will manage when items are saved in elasticsearch database
type DispatcherHandler interface {
	StartIndexData()
	Close() error
	Add(item workItems.WorkItemHandler)
	IsInterfaceNil() bool
}

// ElasticProcessor defines the interface for the elastic search indexer
type ElasticProcessor interface {
	SaveHeader(header coreData.HeaderHandler, signersIndexes []uint64, body *block.Body, notarizedHeadersHashes []string, gasConsumptionData indexer.HeaderGasConsumption, txsSize int) error
	RemoveHeader(header coreData.HeaderHandler) error
	RemoveMiniblocks(header coreData.HeaderHandler, body *block.Body) error
	RemoveTransactions(header coreData.HeaderHandler, body *block.Body) error
	SaveMiniblocks(header coreData.HeaderHandler, body *block.Body) (map[string]bool, error)
	SaveTransactions(body *block.Body, header coreData.HeaderHandler, pool *indexer.Pool, mbsInDb map[string]bool) error
	SaveValidatorsRating(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) error
	SaveRoundsInfo(infos []*data.RoundInfo) error
	SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error
	SaveAccounts(blockTimestamp uint64, accounts []*data.Account) error
	IsInterfaceNil() bool
}

// DatabaseClientHandler is an interface that do requests to elasticsearch server
type DatabaseClientHandler interface {
	DoRequest(req *esapi.IndexRequest) error
	DoBulkRequest(buff *bytes.Buffer, index string) error
	DoBulkRemove(index string, hashes []string) error
	DoMultiGet(query objectsMap, index string) (objectsMap, error)

	CheckAndCreateIndex(index string) error
	CheckAndCreateAlias(alias string, index string) error
	CheckAndCreateTemplate(templateName string, template *bytes.Buffer) error
	CheckAndCreatePolicy(policyName string, policy *bytes.Buffer) error

	IsInterfaceNil() bool
}

// FeesProcessorHandler defines the interface for the transaction fees processor
type FeesProcessorHandler interface {
	ComputeGasUsedAndFeeBasedOnRefundValue(tx coreData.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsed(tx coreData.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeGasLimit(tx coreData.TransactionWithFeeHandler) uint64
	IsInterfaceNil() bool
}

// Coordinator defines what a shard state coordinator should hold
type Coordinator interface {
	ComputeId(address []byte) uint32
	SelfId() uint32
	IsInterfaceNil() bool
}

// Indexer is an interface for saving node specific data to other storage.
// This could be an elastic search index, a MySql database or any other external services.
type Indexer interface {
	SaveBlock(args *indexer.ArgsSaveBlockData) error
	RevertIndexedBlock(header coreData.HeaderHandler, body coreData.BodyHandler) error
	SaveRoundsInfo(roundsInfos []*indexer.RoundInfo)
	SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32)
	SaveValidatorsRating(indexID string, infoRating []*indexer.ValidatorRatingInfo)
	SaveAccounts(blockTimestamp uint64, acc []coreData.UserAccountHandler)
	FinalizedBlock(headerHash []byte) error
	Close() error
	IsInterfaceNil() bool
	IsNilIndexer() bool
}

type AccountsAdapter interface {
	LoadAccount(address []byte) (vmcommon.AccountHandler, error)
	IsInterfaceNil() bool
}
