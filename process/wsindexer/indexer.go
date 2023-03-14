package wsindexer

import (
	"errors"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-core-go/websocketOutportDriver/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var (
	log               = logger.GetOrCreate("process/wsindexer")
	errNilDataIndexer = errors.New("nil data indexer")
)

type indexer struct {
	marshaller marshal.Marshalizer
	di         DataIndexer
}

// NewIndexer will create a new instance of *indexer
func NewIndexer(marshaller marshal.Marshalizer, dataIndexer DataIndexer) (*indexer, error) {
	if check.IfNil(marshaller) {
		return nil, dataindexer.ErrNilMarshalizer
	}
	if check.IfNil(dataIndexer) {
		return nil, errNilDataIndexer
	}

	return &indexer{
		marshaller: marshaller,
		di:         dataIndexer,
	}, nil
}

// GetOperationsMap returns the map with all the operations that will index data
func (i *indexer) GetOperationsMap() map[data.OperationType]func(d []byte) error {
	return map[data.OperationType]func(d []byte) error{
		data.OperationSaveBlock:             i.saveBlock,
		data.OperationRevertIndexedBlock:    i.revertIndexedBlock,
		data.OperationSaveRoundsInfo:        i.saveRounds,
		data.OperationSaveValidatorsRating:  i.saveValidatorsRating,
		data.OperationSaveValidatorsPubKeys: i.saveValidatorsPubKeys,
		data.OperationSaveAccounts:          i.saveAccounts,
		data.OperationFinalizedBlock:        i.finalizedBlock,
	}
}

func (i *indexer) saveBlock(marshalledData []byte) error {
	outportBlock := &outport.OutportBlock{}
	err := i.marshaller.Unmarshal(outportBlock, marshalledData)
	if err != nil {
		return err
	}

	return i.di.SaveBlock(outportBlock)
}

func (i *indexer) revertIndexedBlock(marshalledData []byte) error {
	blockData := &outport.BlockData{}
	err := i.marshaller.Unmarshal(blockData, marshalledData)
	if err != nil {
		return err
	}

	return i.di.RevertIndexedBlock(blockData)
}

func (i *indexer) saveRounds(marshalledData []byte) error {
	roundsInfo := &outport.RoundsInfo{}
	err := i.marshaller.Unmarshal(roundsInfo, marshalledData)
	if err != nil {
		return err
	}

	return i.di.SaveRoundsInfo(roundsInfo)
}

func (i *indexer) saveValidatorsRating(marshalledData []byte) error {
	ratingData := &outport.ValidatorsRating{}
	err := i.marshaller.Unmarshal(ratingData, marshalledData)
	if err != nil {
		return err
	}

	return i.di.SaveValidatorsRating(ratingData)
}

func (i *indexer) saveValidatorsPubKeys(marshalledData []byte) error {
	validatorsPubKeys := &outport.ValidatorsPubKeys{}
	err := i.marshaller.Unmarshal(validatorsPubKeys, marshalledData)
	if err != nil {
		return err
	}

	return i.di.SaveValidatorsPubKeys(validatorsPubKeys)
}

func (i *indexer) saveAccounts(marshalledData []byte) error {
	accounts := &outport.Accounts{}
	err := i.marshaller.Unmarshal(accounts, marshalledData)
	if err != nil {
		return err
	}

	return i.di.SaveAccounts(accounts)
}

func (i *indexer) finalizedBlock(_ []byte) error {
	return nil
}

// Close will close the indexer
func (i *indexer) Close() error {
	return i.di.Close()
}
