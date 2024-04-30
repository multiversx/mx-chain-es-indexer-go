package logstoevents

import (
	"github.com/multiversx/mx-chain-es-indexer-go/migrations/dtos"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWriteStatusOfSplit(t *testing.T) {
	ep, err := NewEventsProcessor(ArgsEventsProc{
		SourceCluster: dtos.ClusterSettings{
			URL: "https://index.multiversx.com",
		},
		DestinationCluster: dtos.ClusterSettings{
			URL: "http://127.0.0.1:9200",
		},
	})
	require.NoError(t, err)

	timestamp := uint64(12345)
	err = ep.SplitLogIndexInEvents("splitEvent", "logs", "events")
	require.NoError(t, err)

	migrationInfo, err := ep.checkStatusOfSplit("splitEvent")
	require.NoError(t, err)
	require.Equal(t, timestamp, migrationInfo.Timestamp)

}
