package indexer

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/workItems"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
	"github.com/ElrondNetwork/elrond-go/process"
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
	SaveShardStatistics(tpsBenchmark statistics.TPSBenchmark) error
	SaveHeader(header nodeData.HeaderHandler, signersIndexes []uint64, body *block.Body, notarizedHeadersHashes []string, txsSize int) error
	RemoveHeader(header nodeData.HeaderHandler) error
	RemoveMiniblocks(header nodeData.HeaderHandler, body *block.Body) error
	RemoveTransactions(header nodeData.HeaderHandler, body *block.Body) error
	SaveMiniblocks(header nodeData.HeaderHandler, body *block.Body) (map[string]bool, error)
	SaveTransactions(body *block.Body, header nodeData.HeaderHandler, pool *indexer.Pool, mbsInDb map[string]bool) error
	SaveValidatorsRating(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) error
	SaveRoundsInfo(infos []*data.RoundInfo) error
	SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error
	SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase)
	SaveAccounts(blockTimestamp uint64, accountsEGLD []*data.Account) error
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
	ComputeGasUsedAndFeeBasedOnRefundValue(tx process.TransactionWithFeeHandler, refundValueStr string) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsed(tx process.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeMoveBalanceGasUsed(tx process.TransactionWithFeeHandler) uint64
}
