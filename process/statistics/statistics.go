package statistics

import (
	"math/big"

	"github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common/statistics"
)

const (
	metachainTpsDocID   = "meta"
	shardTpsDocIDPrefix = "shard"
)

var log = logger.GetOrCreate("indexer/process/statistics")

type statisticsProcessor struct {
}

// NewStatisticsProcessor will create a new instance of a statisticsProcessor
func NewStatisticsProcessor() *statisticsProcessor {
	return &statisticsProcessor{}
}

// PrepareStatistics will prepare the statistics about the chain
func (sp *statisticsProcessor) PrepareStatistics(tpsBenchmark statistics.TPSBenchmark) (*data.TPS, []*data.TPS, error) {
	if check.IfNil(tpsBenchmark) {
		return nil, nil, indexer.ErrNilTPSBenchmark
	}

	generalInfo := &data.TPS{
		LiveTPS:               tpsBenchmark.LiveTPS(),
		PeakTPS:               tpsBenchmark.PeakTPS(),
		NrOfShards:            tpsBenchmark.NrOfShards(),
		BlockNumber:           tpsBenchmark.BlockNumber(),
		RoundNumber:           tpsBenchmark.RoundNumber(),
		RoundTime:             tpsBenchmark.RoundTime(),
		AverageBlockTxCount:   tpsBenchmark.AverageBlockTxCount(),
		LastBlockTxCount:      tpsBenchmark.LastBlockTxCount(),
		TotalProcessedTxCount: tpsBenchmark.TotalProcessedTxCount(),
	}

	shardsInfo := make([]*data.TPS, 0)
	for _, shardInfo := range tpsBenchmark.ShardStatistics() {
		bigTxCount := big.NewInt(int64(shardInfo.AverageBlockTxCount()))
		shardTPS := &data.TPS{
			ShardID:               shardInfo.ShardID(),
			LiveTPS:               shardInfo.LiveTPS(),
			PeakTPS:               shardInfo.PeakTPS(),
			AverageTPS:            shardInfo.AverageTPS(),
			AverageBlockTxCount:   bigTxCount,
			CurrentBlockNonce:     shardInfo.CurrentBlockNonce(),
			LastBlockTxCount:      shardInfo.LastBlockTxCount(),
			TotalProcessedTxCount: shardInfo.TotalProcessedTxCount(),
		}

		shardsInfo = append(shardsInfo, shardTPS)
	}

	return generalInfo, shardsInfo, nil
}
