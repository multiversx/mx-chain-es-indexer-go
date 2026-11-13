package factory

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/process/outputs/rabbitmq"
	"github.com/stretchr/testify/require"
)

func TestPrepareIndices(t *testing.T) {
	t.Parallel()

	available := []string{"index1", "index2"}
	disabled := []string{"index2", "index3"}

	res := prepareIndices(available, disabled)
	require.Equal(t, []string{"index1"}, res)

	available = []string{"index1", "index2"}
	disabled = []string{}

	res = prepareIndices(available, disabled)
	require.Equal(t, []string{"index1", "index2"}, res)
}

func createRabbitClusterConfig(url string) config.ClusterConfig {
	clusterCfg := config.ClusterConfig{}
	clusterCfg.Config.WebSocket.BlockingAckOnError = true
	clusterCfg.Config.RabbitMQ = config.RabbitMQConfig{
		Enabled:         true,
		URL:             url,
		BlocksExchange:  config.RabbitMQExchangeConfig{Name: "blocks", Type: "fanout"},
		RevertsExchange: config.RabbitMQExchangeConfig{Name: "reverts", Type: "fanout"},
	}

	return clusterCfg
}

func TestCreateExtraOutputs(t *testing.T) {
	t.Parallel()

	// nothing listens on this port
	unreachableURL := "amqp://guest:guest@127.0.0.1:1"

	t.Run("RabbitMQ disabled: no extra outputs", func(t *testing.T) {
		t.Parallel()

		extraOutputs, err := createExtraOutputs(config.ClusterConfig{}, nil)
		require.Nil(t, err)
		require.Empty(t, extraOutputs)
	})
	t.Run("invalid url", func(t *testing.T) {
		t.Parallel()

		extraOutputs, err := createExtraOutputs(createRabbitClusterConfig("http://localhost"), nil)
		require.Nil(t, extraOutputs)
		require.Equal(t, rabbitmq.ErrInvalidURL, err)
	})
	t.Run("invalid exchange", func(t *testing.T) {
		t.Parallel()

		clusterCfg := createRabbitClusterConfig(unreachableURL)
		clusterCfg.Config.RabbitMQ.RevertsExchange.Name = ""
		extraOutputs, err := createExtraOutputs(clusterCfg, nil)
		require.Nil(t, extraOutputs)
		require.Equal(t, rabbitmq.ErrInvalidExchangeName, err)
	})
	t.Run("the server cannot be reached: the indexer cannot start if a publish error should fail the block", func(t *testing.T) {
		t.Parallel()

		extraOutputs, err := createExtraOutputs(createRabbitClusterConfig(unreachableURL), nil)
		require.Nil(t, extraOutputs)
		require.True(t, errors.Is(err, rabbitmq.ErrNotConnected))
		require.NotContains(t, err.Error(), "guest:guest")
	})
	t.Run("the server cannot be reached: the indexer starts if the errors are only logged", func(t *testing.T) {
		t.Parallel()

		clusterCfg := createRabbitClusterConfig(unreachableURL)
		clusterCfg.Config.RabbitMQ.ContinueOnError = true
		extraOutputs, err := createExtraOutputs(clusterCfg, nil)
		require.Nil(t, err)
		require.Len(t, extraOutputs, 1)
		require.Equal(t, "rabbitmq", extraOutputs[0].Name)
		require.True(t, extraOutputs[0].ContinueOnError)
		require.Nil(t, extraOutputs[0].Output.Close())
	})
}

type closerStub struct {
	closeCalls int
	err        error
}

func (cs *closerStub) Send(_ []byte, _ string) error {
	return nil
}

func (cs *closerStub) Close() error {
	cs.closeCalls++
	return cs.err
}

func TestWsClientWithIndexer_CloseShouldCloseTheHostAndTheIndexer(t *testing.T) {
	t.Parallel()

	errHost := errors.New("host error")
	errIndexer := errors.New("indexer error")
	host := &closerStub{err: errHost}
	indexer := &closerStub{err: errIndexer}

	wsClient := &wsClientWithIndexer{WSClient: host, indexer: indexer}
	err := wsClient.Close()

	// the web socket host does not close the indexer when it is started in the client mode
	require.Equal(t, 1, host.closeCalls)
	require.Equal(t, 1, indexer.closeCalls)
	require.True(t, errors.Is(err, errHost))
	require.True(t, errors.Is(err, errIndexer))
}
