package dataindexer

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/unmarshal"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer/workItems"
)

// ArgDataIndexer is a structure that is used to store all the components that are needed to create an indexer
type ArgDataIndexer struct {
	Marshalizer      marshal.Marshalizer
	DataDispatcher   DispatcherHandler
	ElasticProcessor ElasticProcessor
}

type dataIndexer struct {
	isNilIndexer     bool
	dispatcher       DispatcherHandler
	elasticProcessor ElasticProcessor
	marshaller       marshal.Marshalizer
	headerMarshaller marshal.Marshalizer
}

// NewDataIndexer will create a new data indexer
func NewDataIndexer(arguments ArgDataIndexer) (*dataIndexer, error) {
	err := checkIndexerArgs(arguments)
	if err != nil {
		return nil, err
	}

	dataIndexerObj := &dataIndexer{
		isNilIndexer:     false,
		dispatcher:       arguments.DataDispatcher,
		elasticProcessor: arguments.ElasticProcessor,
		marshaller:       arguments.Marshalizer,
	}

	return dataIndexerObj, nil
}

func checkIndexerArgs(arguments ArgDataIndexer) error {
	if check.IfNil(arguments.DataDispatcher) {
		return ErrNilDataDispatcher
	}
	if check.IfNil(arguments.ElasticProcessor) {
		return ErrNilElasticProcessor
	}
	if check.IfNil(arguments.Marshalizer) {
		return ErrNilMarshalizer
	}

	return nil
}

// SaveBlock saves the block info in the queue to be sent to elastic
func (di *dataIndexer) SaveBlock(outportBlock *outport.OutportBlock) error {
	header, err := unmarshal.GetHeaderFromBytes(di.headerMarshaller, core.HeaderType(outportBlock.BlockData.HeaderType), outportBlock.BlockData.HeaderBytes)
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
	wi := workItems.NewItemRemoveBlock(
		di.elasticProcessor,
		// TODO possible to be json or proto marshaller
		di.headerMarshaller,
		blockData,
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
func (di *dataIndexer) FinalizedBlock(_ []byte) error {
	return nil
}

// IsNilIndexer will return a bool value that signals if the indexer's implementation is a NilIndexer
func (di *dataIndexer) IsNilIndexer() bool {
	return di.isNilIndexer
}

// IsInterfaceNil returns true if there is no value under the interface
func (di *dataIndexer) IsInterfaceNil() bool {
	return di == nil
}
