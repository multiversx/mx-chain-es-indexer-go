package indexer

import (
	"math/big"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/workItems"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
	"github.com/ElrondNetwork/elrond-go/process"
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
	SaveAccounts(blockTimestamp uint64, accounts []*data.Account) error
	IsInterfaceNil() bool
}

// FeesProcessorHandler defines the interface for the transaction fees processor
type FeesProcessorHandler interface {
	ComputeGasUsedAndFeeBasedOnRefundValue(tx process.TransactionWithFeeHandler, refundValueStr string) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsed(tx process.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeMoveBalanceGasUsed(tx process.TransactionWithFeeHandler) uint64
}
