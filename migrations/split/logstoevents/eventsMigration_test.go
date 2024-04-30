package logstoevents

import (
	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDoMigration(t *testing.T) {
	ep, err := NewEventsMigration(config.ClusterInfo{URL: "https://index.multiversx.com"}, config.ClusterInfo{URL: "http://localhost:9200"})
	require.NoError(t, err)

	err = ep.DoMigration(config.Migration{
		Name:             "split-events",
		SourceIndex:      "logs",
		DestinationIndex: "events",
	})
	require.NoError(t, err)

}
