package statistics

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/stretchr/testify/require"
)

func TestStatisticsProcessor_PrepareStatisticsShouldErr(t *testing.T) {
	t.Parallel()

	sp := NewStatisticsProcessor()

	_, _, err := sp.PrepareStatistics(nil)
	require.Equal(t, indexer.ErrNilTPSBenchmark, err)
}

func TestStatisticsProcessor_PrepareStatistics(t *testing.T) {
	t.Parallel()

	sp := NewStatisticsProcessor()

	tps, _ := statistics.NewTPSBenchmark(1, 2)

	generalInfo, shardStatistics, err := sp.PrepareStatistics(tps)
	require.NotNil(t, generalInfo)
	require.NotNil(t, shardStatistics)
	require.Nil(t, err)
}
