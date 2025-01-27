package tokens

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/factory"
)

const (
	prefix = "sov"
)

func createElasticConfig() factory.ElasticConfig {
	return factory.ElasticConfig{
		Enabled:  true,
		Url:      "http://localhost:9200",
		UserName: "",
		Password: "",
	}
}

func TestSovereignNewIndexTokensHandler(t *testing.T) {
	t.Parallel()

	t.Run("no url, should error", func(t *testing.T) {
		esConfig := createElasticConfig()
		esConfig.Url = "http://bad url"
		sith, err := NewSovereignIndexTokensHandler(esConfig, prefix)
		require.ErrorContains(t, err, "cannot parse url")
		require.Nil(t, sith)
	})
	t.Run("not enabled should not create main chain elastic client", func(t *testing.T) {
		esConfig := createElasticConfig()
		esConfig.Enabled = false
		sith, err := NewSovereignIndexTokensHandler(esConfig, prefix)
		require.NoError(t, err)
		require.NotNil(t, sith)
		require.Nil(t, sith.mainChainElasticClient)
	})
	t.Run("valid config, should work", func(t *testing.T) {
		esConfig := createElasticConfig()
		sith, err := NewSovereignIndexTokensHandler(esConfig, prefix)
		require.NoError(t, err)
		require.NotNil(t, sith)
	})
}

func TestSovereignIndexTokensHandler_IndexCrossChainTokens(t *testing.T) {
	t.Parallel()

	esConfig := createElasticConfig()
	esConfig.Enabled = false
	sith, err := NewSovereignIndexTokensHandler(esConfig, prefix)
	require.NoError(t, err)
	require.NotNil(t, sith)

	// should skip indexing
	err = sith.IndexCrossChainTokens(nil, make([]*data.ScResult, 0), data.NewBufferSlice(0))
	require.NoError(t, err)

	// actual indexing is tested in TestCrossChainTokensIndexingFromMainChain
}
