package wsindexer

import (
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go-core/websocketOutportDriver/data"
)

type indexer struct {
	marshaller marshal.Marshalizer
	di         DataIndexer
}

func NewIndexer(marshaller marshal.Marshalizer, dataIndexer DataIndexer) (*indexer, error) {
	return &indexer{
		marshaller: marshaller,
		di:         dataIndexer,
	}, nil
}

func (i *indexer) GetFunctionsMap() map[data.OperationType]func(d []byte) error {
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
	argsSaveBlock := &outport.ArgsSaveBlockData{}
	err := i.marshaller.Unmarshal(argsSaveBlock, marshalledData)
	if err != nil {
		return err
	}

	return i.di.SaveBlock(argsSaveBlock)
}

func (i *indexer) revertIndexedBlock(marshalledData []byte) error {
	argsRevert := &data.ArgsRevertIndexedBlock{}
	err := i.marshaller.Unmarshal(argsRevert, marshalledData)
	if err != nil {
		return err
	}

	return i.di.RevertIndexedBlock(argsRevert.Header, argsRevert.Body)
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
