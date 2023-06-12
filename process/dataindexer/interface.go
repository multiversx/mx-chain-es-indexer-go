package dataindexer

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

// ElasticProcessor defines the interface for the elastic search indexer
type ElasticProcessor interface {
	SaveHeader(outportBlockWithHeader *outport.OutportBlockWithHeader) error
	RemoveHeader(header coreData.HeaderHandler) error
	RemoveMiniblocks(header coreData.HeaderHandler, body *block.Body) error
	RemoveTransactions(header coreData.HeaderHandler, body *block.Body) error
	RemoveAccountsESDT(headerTimestamp uint64, shardID uint32) error
	SaveMiniblocks(header coreData.HeaderHandler, body *block.Body) error
	SaveTransactions(outportBlockWithHeader *outport.OutportBlockWithHeader) error
	SaveValidatorsRating(ratingData *outport.ValidatorsRating) error
	SaveRoundsInfo(rounds *outport.RoundsInfo) error
	SaveShardValidatorsPubKeys(validatorsPubKeys *outport.ValidatorsPubKeys) error
	SaveAccounts(accounts *outport.Accounts) error
	SetOutportConfig(cfg outport.OutportConfig) error
	IsInterfaceNil() bool
}

// FeesProcessorHandler defines the interface for the transaction fees processor
type FeesProcessorHandler interface {
	ComputeGasUsedAndFeeBasedOnRefundValue(tx coreData.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsed(tx coreData.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeGasLimit(tx coreData.TransactionWithFeeHandler) uint64
	IsInterfaceNil() bool
}

// ShardCoordinator defines what a shard state coordinator should hold
type ShardCoordinator interface {
	NumberOfShards() uint32
	ComputeId(address []byte) uint32
	SelfId() uint32
	SameShard(firstAddress, secondAddress []byte) bool
	CommunicationIdentifier(destShardID uint32) string
	IsInterfaceNil() bool
}

// Indexer is an interface for saving node specific data to other storage.
// This could be an elastic search index, a MySql database or any other external services.
type Indexer interface {
	SaveBlock(outportBlock *outport.OutportBlock) error
	RevertIndexedBlock(blockData *outport.BlockData) error
	SaveRoundsInfo(roundsInfos *outport.RoundsInfo) error
	SaveValidatorsPubKeys(validatorsPubKeys *outport.ValidatorsPubKeys) error
	SaveValidatorsRating(ratingData *outport.ValidatorsRating) error
	SaveAccounts(accountsData *outport.Accounts) error
	FinalizedBlock(finalizedBlock *outport.FinalizedBlock) error
	GetMarshaller() marshal.Marshalizer
	RegisterHandler(handler func() error, topic string) error
	SetCurrentSettings(cfg outport.OutportConfig) error
	Close() error
	IsInterfaceNil() bool
}

// BalanceConverter defines what a balance converter should be able to do
type BalanceConverter interface {
	ComputeBalanceAsFloat(balance *big.Int) (float64, error)
	ComputeESDTBalanceAsFloat(balance *big.Int) (float64, error)
	ComputeSliceOfStringsAsFloat(values []string) ([]float64, error)
	IsInterfaceNil() bool
}

// BlockContainerHandler defines what a block container should be able to do
type BlockContainerHandler interface {
	Get(headerType core.HeaderType) (block.EmptyBlockCreator, error)
}
