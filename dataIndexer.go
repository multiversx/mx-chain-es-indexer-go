package indexer

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/workItems"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

type dataIndexer struct {
	isNilIndexer     bool
	dispatcher       DispatcherHandler
	elasticProcessor ElasticProcessor
	marshalizer      marshal.Marshalizer
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
		marshalizer:      arguments.Marshalizer,
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
		return core.ErrNilMarshalizer
	}
	if check.IfNil(arguments.ShardCoordinator) {
		return ErrNilShardCoordinator
	}

	return nil
}

// SaveBlock saves the block info in the queue to be sent to elastic
func (di *dataIndexer) SaveBlock(args *indexer.ArgsSaveBlockData) {
	wi := workItems.NewItemBlock(
		di.elasticProcessor,
		di.marshalizer,
		args,
	)
	di.dispatcher.Add(wi)
}

// Close will stop goroutine that index data in database
func (di *dataIndexer) Close() error {
	return di.dispatcher.Close()
}

// RevertIndexedBlock will remove from database block and miniblocks
func (di *dataIndexer) RevertIndexedBlock(header coreData.HeaderHandler, body coreData.BodyHandler) {
	wi := workItems.NewItemRemoveBlock(
		di.elasticProcessor,
		body,
		header,
	)
	di.dispatcher.Add(wi)
}

// SaveRoundsInfo will save data about a slice of rounds in elasticsearch
func (di *dataIndexer) SaveRoundsInfo(rf []*indexer.RoundInfo) {
	roundsInfo := make([]*data.RoundInfo, 0)
	for _, info := range rf {
		roundsInfo = append(roundsInfo, &data.RoundInfo{
			Index:            info.Index,
			SignersIndexes:   info.SignersIndexes,
			BlockWasProposed: info.BlockWasProposed,
			ShardId:          info.ShardId,
			Timestamp:        info.Timestamp,
		})
	}

	wi := workItems.NewItemRounds(di.elasticProcessor, roundsInfo)
	di.dispatcher.Add(wi)
}

// SaveValidatorsRating will save all validators rating info to elasticsearch
func (di *dataIndexer) SaveValidatorsRating(indexID string, validatorsRatingInfo []*indexer.ValidatorRatingInfo) {
	valRatingInfo := make([]*data.ValidatorRatingInfo, 0)
	for _, info := range validatorsRatingInfo {
		valRatingInfo = append(valRatingInfo, &data.ValidatorRatingInfo{
			PublicKey: info.PublicKey,
			Rating:    info.Rating,
		})
	}

	wi := workItems.NewItemRating(
		di.elasticProcessor,
		indexID,
		valRatingInfo,
	)
	di.dispatcher.Add(wi)
}

// SaveValidatorsPubKeys will save all validators public keys to elasticsearch
func (di *dataIndexer) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) {
	wi := workItems.NewItemValidators(
		di.elasticProcessor,
		epoch,
		validatorsPubKeys,
	)
	di.dispatcher.Add(wi)
}

// SaveAccounts will save the provided accounts
func (di *dataIndexer) SaveAccounts(timestamp uint64, accounts []coreData.UserAccountHandler) {
	wi := workItems.NewItemAccounts(di.elasticProcessor, timestamp, accounts)
	di.dispatcher.Add(wi)
}

// IsNilIndexer will return a bool value that signals if the indexer's implementation is a NilIndexer
func (di *dataIndexer) IsNilIndexer() bool {
	return di.isNilIndexer
}

// IsInterfaceNil returns true if there is no value under the interface
func (di *dataIndexer) IsInterfaceNil() bool {
	return di == nil
}
