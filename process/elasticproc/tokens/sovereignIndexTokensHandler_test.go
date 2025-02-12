package tokens

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-es-indexer-go/client/disabled"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
)

const (
	prefix = "sov"
)

func TestSovereignNewIndexTokensHandler(t *testing.T) {
	t.Parallel()

	t.Run("valid disabled config, should work", func(t *testing.T) {
		sith, err := NewSovereignIndexTokensHandler(disabled.NewDisabledElasticClient(), prefix)
		require.NoError(t, err)
		require.Equal(t, "*disabled.elasticClient", fmt.Sprintf("%T", sith.mainChainElasticClient))
	})
	t.Run("valid config, should work", func(t *testing.T) {
		sith, err := NewSovereignIndexTokensHandler(&mock.DatabaseWriterStub{}, prefix)
		require.NoError(t, err)
		require.Equal(t, "*mock.DatabaseWriterStub", fmt.Sprintf("%T", sith.mainChainElasticClient))
	})
}

func TestSovereignIndexTokensHandler_IndexCrossChainTokens(t *testing.T) {
	t.Parallel()

	sith, err := NewSovereignIndexTokensHandler(disabled.NewDisabledElasticClient(), prefix)
	require.NoError(t, err)
	require.NotNil(t, sith)

	// should skip indexing
	err = sith.IndexCrossChainTokens(nil, make([]*data.ScResult, 0), data.NewBufferSlice(0))
	require.NoError(t, err)

	// actual indexing is tested in TestCrossChainTokensIndexingFromMainChain
}
