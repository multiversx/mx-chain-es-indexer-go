package mock

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
)

// ElasticProcessorStub -
type ElasticProcessorStub struct {
	SaveHeaderCalled func(
		headerHash []byte,
		header coreData.HeaderHandler,
		signersIndexes []uint64,
		body *block.Body,
		notarizedHeadersHashes []string,
		gasConsumptionData indexer.HeaderGasConsumption,
		txsSize int,
	) error
	RemoveHeaderCalled               func(header coreData.HeaderHandler) error
	RemoveMiniblocksCalled           func(header coreData.HeaderHandler, body *block.Body) error
	RemoveTransactionsCalled         func(header coreData.HeaderHandler, body *block.Body) error
	SaveMiniblocksCalled             func(header coreData.HeaderHandler, body *block.Body) error
	SaveTransactionsCalled           func(body *block.Body, header coreData.HeaderHandler, pool *indexer.Pool) error
	SaveValidatorsRatingCalled       func(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) error
	SaveRoundsInfoCalled             func(infos []*data.RoundInfo) error
	SaveShardValidatorsPubKeysCalled func(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error
	SaveAccountsCalled               func(timestamp uint64, acc []*data.Account) error
	RemoveAccountsESDTCalled         func(headerTimestamp uint64) error
}

// RemoveAccountsESDT -
func (eim *ElasticProcessorStub) RemoveAccountsESDT(headerTimestamp uint64) error {
	if eim.RemoveAccountsESDTCalled != nil {
		return eim.RemoveAccountsESDTCalled(headerTimestamp)
	}

	return nil
}

// SaveHeader -
func (eim *ElasticProcessorStub) SaveHeader(
	headerHash []byte,
	header coreData.HeaderHandler,
	signersIndexes []uint64,
	body *block.Body,
	notarizedHeadersHashes []string,
	gasConsumptionData indexer.HeaderGasConsumption,
	txsSize int) error {
	if eim.SaveHeaderCalled != nil {
		return eim.SaveHeaderCalled(headerHash, header, signersIndexes, body, notarizedHeadersHashes, gasConsumptionData, txsSize)
	}
	return nil
}

// RemoveHeader -
func (eim *ElasticProcessorStub) RemoveHeader(header coreData.HeaderHandler) error {
	if eim.RemoveHeaderCalled != nil {
		return eim.RemoveHeaderCalled(header)
	}
	return nil
}

// RemoveMiniblocks -
func (eim *ElasticProcessorStub) RemoveMiniblocks(header coreData.HeaderHandler, body *block.Body) error {
	if eim.RemoveMiniblocksCalled != nil {
		return eim.RemoveMiniblocksCalled(header, body)
	}
	return nil
}

// RemoveTransactions -
func (eim *ElasticProcessorStub) RemoveTransactions(header coreData.HeaderHandler, body *block.Body) error {
	if eim.RemoveMiniblocksCalled != nil {
		return eim.RemoveTransactionsCalled(header, body)
	}
	return nil
}

// SaveMiniblocks -
func (eim *ElasticProcessorStub) SaveMiniblocks(header coreData.HeaderHandler, body *block.Body) error {
	if eim.SaveMiniblocksCalled != nil {
		return eim.SaveMiniblocksCalled(header, body)
	}
	return nil
}

// SaveTransactions -
func (eim *ElasticProcessorStub) SaveTransactions(body *block.Body, header coreData.HeaderHandler, pool *indexer.Pool) error {
	if eim.SaveTransactionsCalled != nil {
		return eim.SaveTransactionsCalled(body, header, pool)
	}
	return nil
}

// SaveValidatorsRating -
func (eim *ElasticProcessorStub) SaveValidatorsRating(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) error {
	if eim.SaveValidatorsRatingCalled != nil {
		return eim.SaveValidatorsRatingCalled(index, validatorsRatingInfo)
	}
	return nil
}

// SaveRoundsInfo -
func (eim *ElasticProcessorStub) SaveRoundsInfo(info []*data.RoundInfo) error {
	if eim.SaveRoundsInfoCalled != nil {
		return eim.SaveRoundsInfoCalled(info)
	}
	return nil
}

// SaveShardValidatorsPubKeys -
func (eim *ElasticProcessorStub) SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error {
	if eim.SaveShardValidatorsPubKeysCalled != nil {
		return eim.SaveShardValidatorsPubKeysCalled(shardID, epoch, shardValidatorsPubKeys)
	}
	return nil
}

// SaveAccounts -
func (eim *ElasticProcessorStub) SaveAccounts(timestamp uint64, acc []*data.Account) error {
	if eim.SaveAccountsCalled != nil {
		return eim.SaveAccountsCalled(timestamp, acc)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (eim *ElasticProcessorStub) IsInterfaceNil() bool {
	return eim == nil
}
