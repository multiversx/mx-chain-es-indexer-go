package mock

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ElasticProcessorStub -
type ElasticProcessorStub struct {
	SaveShardStatisticsCalled        func(tpsBenchmark statistics.TPSBenchmark) error
	SaveHeaderCalled                 func(header nodeData.HeaderHandler, signersIndexes []uint64, body *block.Body, notarizedHeadersHashes []string, txsSize int) error
	RemoveHeaderCalled               func(header nodeData.HeaderHandler) error
	RemoveMiniblocksCalled           func(header nodeData.HeaderHandler, body *block.Body) error
	RemoveTransactionsCalled         func(header nodeData.HeaderHandler, body *block.Body) error
	SaveMiniblocksCalled             func(header nodeData.HeaderHandler, body *block.Body) (map[string]bool, error)
	SaveTransactionsCalled           func(body *block.Body, header nodeData.HeaderHandler, pool *indexer.Pool, mbsInDb map[string]bool) error
	SaveValidatorsRatingCalled       func(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) error
	SaveRoundsInfoCalled             func(infos []*data.RoundInfo) error
	SaveShardValidatorsPubKeysCalled func(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error
	SetTxLogsProcessorCalled         func(txLogsProc process.TransactionLogProcessorDatabase)
	SaveAccountsCalled               func(timestamp uint64, acc []*data.Account) error
}

// SaveShardStatistics -
func (eim *ElasticProcessorStub) SaveShardStatistics(tpsBenchmark statistics.TPSBenchmark) error {
	if eim.SaveShardStatisticsCalled != nil {
		return eim.SaveShardStatisticsCalled(tpsBenchmark)
	}
	return nil
}

// SaveHeader -
func (eim *ElasticProcessorStub) SaveHeader(header nodeData.HeaderHandler, signersIndexes []uint64, body *block.Body, notarizedHeadersHashes []string, txsSize int) error {
	if eim.SaveHeaderCalled != nil {
		return eim.SaveHeaderCalled(header, signersIndexes, body, notarizedHeadersHashes, txsSize)
	}
	return nil
}

// RemoveHeader -
func (eim *ElasticProcessorStub) RemoveHeader(header nodeData.HeaderHandler) error {
	if eim.RemoveHeaderCalled != nil {
		return eim.RemoveHeaderCalled(header)
	}
	return nil
}

// RemoveMiniblocks -
func (eim *ElasticProcessorStub) RemoveMiniblocks(header nodeData.HeaderHandler, body *block.Body) error {
	if eim.RemoveMiniblocksCalled != nil {
		return eim.RemoveMiniblocksCalled(header, body)
	}
	return nil
}

// RemoveTransactions -
func (eim *ElasticProcessorStub) RemoveTransactions(header nodeData.HeaderHandler, body *block.Body) error {
	if eim.RemoveMiniblocksCalled != nil {
		return eim.RemoveTransactionsCalled(header, body)
	}
	return nil
}

// SaveMiniblocks -
func (eim *ElasticProcessorStub) SaveMiniblocks(header nodeData.HeaderHandler, body *block.Body) (map[string]bool, error) {
	if eim.SaveMiniblocksCalled != nil {
		return eim.SaveMiniblocksCalled(header, body)
	}
	return nil, nil
}

// SaveTransactions -
func (eim *ElasticProcessorStub) SaveTransactions(body *block.Body, header nodeData.HeaderHandler, pool *indexer.Pool, mbsInDb map[string]bool) error {
	if eim.SaveTransactionsCalled != nil {
		return eim.SaveTransactionsCalled(body, header, pool, mbsInDb)
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

// SetTxLogsProcessor -
func (eim *ElasticProcessorStub) SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase) {
	if eim.SetTxLogsProcessorCalled != nil {
		eim.SetTxLogsProcessorCalled(txLogsProc)
	}
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
