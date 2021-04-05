package statistics

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/stretchr/testify/require"
)

func TestStatisticsProcessor_SerializeStatistics(t *testing.T) {
	t.Parallel()

	sp := NewStatisticsProcessor()

	generalInfo := &data.TPS{}
	shardsInfo := []*data.TPS{
		{}, {},
	}

	buff, err := sp.SerializeStatistics(generalInfo, shardsInfo, "tps")
	require.Nil(t, err)
	expectedBuff := `{ "index" : { "_id" : "meta", "_type" : "tps" } }
{"liveTPS":0,"peakTPS":0,"blockNumber":0,"roundNumber":0,"roundTime":0,"averageBlockTxCount":null,"totalProcessedTxCount":null,"averageTPS":null,"currentBlockNonce":0,"nrOfShards":0,"nrOfNodes":0,"lastBlockTxCount":0,"shardID":0}
{ "index" : { "_id" : "shard0", "_type" : "tps" } }
{"liveTPS":0,"peakTPS":0,"blockNumber":0,"roundNumber":0,"roundTime":0,"averageBlockTxCount":null,"totalProcessedTxCount":null,"averageTPS":null,"currentBlockNonce":0,"nrOfShards":0,"nrOfNodes":0,"lastBlockTxCount":0,"shardID":0}
{ "index" : { "_id" : "shard0", "_type" : "tps" } }
{"liveTPS":0,"peakTPS":0,"blockNumber":0,"roundNumber":0,"roundTime":0,"averageBlockTxCount":null,"totalProcessedTxCount":null,"averageTPS":null,"currentBlockNonce":0,"nrOfShards":0,"nrOfNodes":0,"lastBlockTxCount":0,"shardID":0}
`
	require.Equal(t, expectedBuff, buff.String())
}

func TestStatisticsProcessor_SerializeRoundsInfo(t *testing.T) {
	t.Parallel()

	sp := NewStatisticsProcessor()

	buff := sp.SerializeRoundsInfo([]*data.RoundInfo{{}})
	expectedBuff := `{ "index" : { "_id" : "0_0", "_type" : "_doc" } }
{"round":0,"signersIndexes":null,"blockWasProposed":false,"shardId":0,"timestamp":0}
`
	require.Equal(t, expectedBuff, buff.String())
}
