package wsindexer

import (
	"errors"

	"github.com/multiversx/mx-chain-core-go/core/check"
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
	argsSaveBlockS, err := i.getArgsSaveBlock(marshalledData)
	if err != nil {
		return err
	}

	return i.di.SaveBlock(argsSaveBlockS)
}

func (i *indexer) revertIndexedBlock(marshalledData []byte) error {
	header, body, err := i.getHeaderAndBody(marshalledData)
	if err != nil {
		return err
	}

	return i.di.RevertIndexedBlock(header, body)
}

func (i *indexer) saveRounds(marshalledData []byte) error {
	argsRounds := &data.ArgsSaveRoundsInfo{}
	err := i.marshaller.Unmarshal(argsRounds, marshalledData)
	if err != nil {
		return err
	}

	return i.di.SaveRoundsInfo(argsRounds.RoundsInfos)
}

func (i *indexer) saveValidatorsRating(marshalledData []byte) error {
	argsValidatorsRating := &data.ArgsSaveValidatorsRating{}
	err := i.marshaller.Unmarshal(argsValidatorsRating, marshalledData)
	if err != nil {
		return err
	}

	return i.di.SaveValidatorsRating(argsValidatorsRating.IndexID, argsValidatorsRating.InfoRating)
}

func (i *indexer) saveValidatorsPubKeys(marshalledData []byte) error {
	argsValidators := &data.ArgsSaveValidatorsPubKeys{}
	err := i.marshaller.Unmarshal(argsValidators, marshalledData)
	if err != nil {
		return err
	}

	return i.di.SaveValidatorsPubKeys(argsValidators.ValidatorsPubKeys, argsValidators.Epoch)
}

func (i *indexer) saveAccounts(marshalledData []byte) error {
	argsSaveAccounts := &data.ArgsSaveAccounts{}
	err := i.marshaller.Unmarshal(argsSaveAccounts, marshalledData)
	if err != nil {
		return err
	}

	return i.di.SaveAccounts(argsSaveAccounts.BlockTimestamp, argsSaveAccounts.Acc, argsSaveAccounts.ShardID)
}

func (i *indexer) finalizedBlock(_ []byte) error {
	return nil
}

// Close will close the indexer
func (i *indexer) Close() error {
	return i.di.Close()
}
