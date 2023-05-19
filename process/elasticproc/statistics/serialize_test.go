package statistics

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/stretchr/testify/require"
)

func TestStatisticsProcessor_SerializeRoundsInfo(t *testing.T) {
	t.Parallel()

	sp := NewStatisticsProcessor()

	buff := sp.SerializeRoundsInfo(&outport.RoundsInfo{
		RoundsInfo: []*outport.RoundInfo{{
			Epoch: 1,
		}},
	})
	expectedBuff := `{ "index" : { "_id" : "0_0" } }
{"round":0,"signersIndexes":null,"blockWasProposed":false,"shardId":0,"epoch":1,"timestamp":0}
`
	require.Equal(t, expectedBuff, buff.String())
}
