package statistics

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/stretchr/testify/require"
)

func TestStatisticsProcessor_SerializeRoundsInfo(t *testing.T) {
	t.Parallel()

	sp := NewStatisticsProcessor()

	buff := sp.SerializeRoundsInfo([]*data.RoundInfo{{
		Epoch: 1,
	}})
	expectedBuff := `{ "index" : { "_id" : "0_0" } }
{"round":0,"signersIndexes":null,"blockWasProposed":false,"shardId":0,"epoch":1,"timestamp":0}
`
	require.Equal(t, expectedBuff, buff.String())
}
