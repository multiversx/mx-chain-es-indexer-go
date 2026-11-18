package config

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestClusterConfig_ShippedPreferencesFile(t *testing.T) {
	t.Parallel()

	cfg := ClusterConfig{}
	err := core.LoadTomlFile(&cfg, "../cmd/elasticindexer/config/prefs.toml")
	require.Nil(t, err)

	require.Equal(t, "http://localhost:9200", cfg.Config.ElasticCluster.URL)

	// the RabbitMQ output is disabled by default and a publish error fails the block
	rabbitCfg := cfg.Config.RabbitMQ
	require.False(t, rabbitCfg.Enabled)
	require.False(t, rabbitCfg.ContinueOnError)
	require.Equal(t, "amqp://guest:guest@localhost:5672", rabbitCfg.URL)
	require.Equal(t, uint32(10), rabbitCfg.PublishTimeoutInSec)
	require.Equal(t, RabbitMQExchangeConfig{Name: "indexer_blocks", Type: "fanout"}, rabbitCfg.BlocksExchange)
	require.Equal(t, RabbitMQExchangeConfig{Name: "indexer_reverts", Type: "fanout"}, rabbitCfg.RevertsExchange)
}

func TestClusterConfig_PreferencesFileWithoutRabbitMQ(t *testing.T) {
	t.Parallel()

	// a preferences file created before the RabbitMQ output was added
	prefs := `
[config]
    disabled-indices = []
    [config.elastic-cluster]
        url = "http://localhost:9200"
`
	cfg := ClusterConfig{}
	err := core.LoadTomlFile(&cfg, writeTempFile(t, prefs))
	require.Nil(t, err)

	require.False(t, cfg.Config.RabbitMQ.Enabled)
	require.False(t, cfg.Config.RabbitMQ.ContinueOnError)
}
