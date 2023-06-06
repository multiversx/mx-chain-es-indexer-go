package dataindexer

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("dataindexer")

// ArgDataIndexer is a structure that is used to store all the components that are needed to create an indexer
type ArgDataIndexer struct {
	HeaderMarshaller marshal.Marshalizer
	ElasticProcessor ElasticProcessor
	BlockContainer   BlockContainerHandler
}

type dataIndexer struct {
	elasticProcessor ElasticProcessor
	headerMarshaller marshal.Marshalizer
	blockContainer   BlockContainerHandler
}

// NewDataIndexer will create a new data indexer
func NewDataIndexer(arguments ArgDataIndexer) (*dataIndexer, error) {
	err := checkIndexerArgs(arguments)
	if err != nil {
		return nil, err
	}

	dataIndexerObj := &dataIndexer{
		elasticProcessor: arguments.ElasticProcessor,
		headerMarshaller: arguments.HeaderMarshaller,
		blockContainer:   arguments.BlockContainer,
	}

	return dataIndexerObj, nil
}

func checkIndexerArgs(arguments ArgDataIndexer) error {
	if check.IfNil(arguments.ElasticProcessor) {
		return ErrNilElasticProcessor
	}
	if check.IfNil(arguments.HeaderMarshaller) {
		return ErrNilMarshalizer
	}
	if check.IfNilReflect(arguments.BlockContainer) {
		return ErrNilBlockContainerHandler
	}

	return nil
}

func (di *dataIndexer) getHeaderFromBytes(headerType core.HeaderType, headerBytes []byte) (header data.HeaderHandler, err error) {
	creator, err := di.blockContainer.Get(headerType)
	if err != nil {
		return nil, err
	}

	return block.GetHeaderFromBytes(di.headerMarshaller, creator, headerBytes)
}

// SaveBlock saves the block info in the queue to be sent to elastic
func (di *dataIndexer) SaveBlock(outportBlock *outport.OutportBlock) error {
	header, err := di.getHeaderFromBytes(core.HeaderType(outportBlock.BlockData.HeaderType), outportBlock.BlockData.HeaderBytes)
	if err != nil {
		return err
	}

	headerHash := outportBlock.BlockData.HeaderHash
	shardID := header.GetShardID()
	headerNonce := header.GetNonce()
	defer func(startTime time.Time, headerHash []byte, headerNonce uint64, shardID uint32) {
		log.Debug("di.SaveBlockData",
			"duration", time.Since(startTime),
			"shardID", shardID,
			"nonce", headerNonce,
			"hash", headerHash,
		)
	}(time.Now(), headerHash, headerNonce, shardID)
	log.Debug("indexer: starting indexing block", "hash", headerHash, "nonce", headerNonce)

	if outportBlock.TransactionPool == nil {
		outportBlock.TransactionPool = &outport.TransactionPool{}
	}

	return di.saveBlockData(outportBlock, header)
}

func (di *dataIndexer) saveBlockData(outportBlock *outport.OutportBlock, header data.HeaderHandler) error {
	outportBlockWithHeader := &outport.OutportBlockWithHeader{
		OutportBlock: outportBlock,
		Header:       header,
	}

	headerHash := outportBlock.BlockData.HeaderHash
	headerNonce := header.GetNonce()
	err := di.elasticProcessor.SaveHeader(outportBlockWithHeader)
	if err != nil {
		return fmt.Errorf("%w when saving header block, hash %s, nonce %d",
			err, hex.EncodeToString(headerHash), headerNonce)
	}

	if len(outportBlock.BlockData.Body.MiniBlocks) == 0 {
		return nil
	}

	err = di.elasticProcessor.SaveMiniblocks(header, outportBlock.BlockData.Body)
	if err != nil {
		return fmt.Errorf("%w when saving miniblocks, block hash %s, nonce %d",
			err, hex.EncodeToString(headerHash), headerNonce)
	}

	err = di.elasticProcessor.SaveTransactions(outportBlockWithHeader)
	if err != nil {
		return fmt.Errorf("%w when saving transactions, block hash %s, nonce %d",
			err, hex.EncodeToString(headerHash), headerNonce)
	}

	return nil
}

// Close will stop goroutine that index data in database
func (di *dataIndexer) Close() error {
	return nil
}

// RevertIndexedBlock will remove from database block and miniblocks
func (di *dataIndexer) RevertIndexedBlock(blockData *outport.BlockData) error {
	header, err := di.getHeaderFromBytes(core.HeaderType(blockData.HeaderType), blockData.HeaderBytes)
	if err != nil {
		return err
	}

	err = di.elasticProcessor.RemoveHeader(header)
	if err != nil {
		return err
	}

	err = di.elasticProcessor.RemoveMiniblocks(header, blockData.Body)
	if err != nil {
		return err
	}

	err = di.elasticProcessor.RemoveTransactions(header, blockData.Body)
	if err != nil {
		return err
	}

	return di.elasticProcessor.RemoveAccountsESDT(header.GetTimeStamp(), header.GetShardID())
}

// SaveRoundsInfo will save data about a slice of rounds in elasticsearch
func (di *dataIndexer) SaveRoundsInfo(rounds *outport.RoundsInfo) error {
	return di.elasticProcessor.SaveRoundsInfo(rounds)
}

// SaveValidatorsRating will save all validators rating info to elasticsearch
func (di *dataIndexer) SaveValidatorsRating(ratingData *outport.ValidatorsRating) error {
	return di.elasticProcessor.SaveValidatorsRating(ratingData)
}

// SaveValidatorsPubKeys will save all validators public keys to elasticsearch
func (di *dataIndexer) SaveValidatorsPubKeys(validatorsPubKeys *outport.ValidatorsPubKeys) error {
	return di.elasticProcessor.SaveShardValidatorsPubKeys(validatorsPubKeys)
}

// SaveAccounts will save the provided accounts
func (di *dataIndexer) SaveAccounts(accounts *outport.Accounts) error {
	return di.elasticProcessor.SaveAccounts(accounts)
}

// FinalizedBlock returns nil
func (di *dataIndexer) FinalizedBlock(_ *outport.FinalizedBlock) error {
	return nil
}

// GetMarshaller return the marshaller
func (di *dataIndexer) GetMarshaller() marshal.Marshalizer {
	return di.headerMarshaller
}

// IsInterfaceNil returns true if there is no value under the interface
func (di *dataIndexer) IsInterfaceNil() bool {
	return di == nil
}
