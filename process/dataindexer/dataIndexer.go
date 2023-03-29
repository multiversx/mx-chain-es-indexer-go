package dataindexer

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer/workItems"
)

// ArgDataIndexer is a structure that is used to store all the components that are needed to create an indexer
type ArgDataIndexer struct {
	HeaderMarshaller marshal.Marshalizer
	DataDispatcher   DispatcherHandler
	ElasticProcessor ElasticProcessor
}

type dataIndexer struct {
	isNilIndexer     bool
	dispatcher       DispatcherHandler
	elasticProcessor ElasticProcessor
	headerMarshaller marshal.Marshalizer
	blockContainer   blockContainerHandler
}

// NewDataIndexer will create a new data indexer
func NewDataIndexer(arguments ArgDataIndexer) (*dataIndexer, error) {
	err := checkIndexerArgs(arguments)
	if err != nil {
		return nil, err
	}

	blockContainer, err := createBlockCreatorsContainer()
	if err != nil {
		return nil, err
	}

	dataIndexerObj := &dataIndexer{
		isNilIndexer:     false,
		dispatcher:       arguments.DataDispatcher,
		elasticProcessor: arguments.ElasticProcessor,
		headerMarshaller: arguments.HeaderMarshaller,
		blockContainer:   blockContainer,
	}

	return dataIndexerObj, nil
}

func createBlockCreatorsContainer() (blockContainerHandler, error) {
	container := block.NewEmptyBlockCreatorsContainer()
	err := container.Add(core.ShardHeaderV1, block.NewEmptyHeaderCreator())
	if err != nil {
		return nil, err
	}
	err = container.Add(core.ShardHeaderV2, block.NewEmptyHeaderV2Creator())
	if err != nil {
		return nil, err
	}
	err = container.Add(core.MetaHeader, block.NewEmptyMetaBlockCreator())
	if err != nil {
		return nil, err
	}

	return container, nil
}

func checkIndexerArgs(arguments ArgDataIndexer) error {
	if check.IfNil(arguments.DataDispatcher) {
		return ErrNilDataDispatcher
	}
	if check.IfNil(arguments.ElasticProcessor) {
		return ErrNilElasticProcessor
	}
	if check.IfNil(arguments.HeaderMarshaller) {
		return ErrNilMarshalizer
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

	wi := workItems.NewItemBlock(
		di.elasticProcessor,
		&outport.OutportBlockWithHeader{
			OutportBlock: outportBlock,
			Header:       header,
		},
	)
	di.dispatcher.Add(wi)

	return nil
}

// Close will stop goroutine that index data in database
func (di *dataIndexer) Close() error {
	return di.dispatcher.Close()
}

// RevertIndexedBlock will remove from database block and miniblocks
func (di *dataIndexer) RevertIndexedBlock(blockData *outport.BlockData) error {
	header, err := di.getHeaderFromBytes(core.HeaderType(blockData.HeaderType), blockData.HeaderBytes)
	if err != nil {
		return err
	}

	wi := workItems.NewItemRemoveBlock(
		di.elasticProcessor,
		header,
		blockData.Body,
	)
	di.dispatcher.Add(wi)

	return nil
}

// SaveRoundsInfo will save data about a slice of rounds in elasticsearch
func (di *dataIndexer) SaveRoundsInfo(rounds *outport.RoundsInfo) error {
	wi := workItems.NewItemRounds(di.elasticProcessor, rounds)
	di.dispatcher.Add(wi)

	return nil
}

// SaveValidatorsRating will save all validators rating info to elasticsearch
func (di *dataIndexer) SaveValidatorsRating(ratingData *outport.ValidatorsRating) error {
	wi := workItems.NewItemRating(
		di.elasticProcessor,
		ratingData,
	)
	di.dispatcher.Add(wi)

	return nil
}

// SaveValidatorsPubKeys will save all validators public keys to elasticsearch
func (di *dataIndexer) SaveValidatorsPubKeys(validatorsPubKeys *outport.ValidatorsPubKeys) error {
	wi := workItems.NewItemValidators(
		di.elasticProcessor,
		validatorsPubKeys,
	)
	di.dispatcher.Add(wi)

	return nil
}

// SaveAccounts will save the provided accounts
func (di *dataIndexer) SaveAccounts(accounts *outport.Accounts) error {
	wi := workItems.NewItemAccounts(di.elasticProcessor, accounts)
	di.dispatcher.Add(wi)

	return nil
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
