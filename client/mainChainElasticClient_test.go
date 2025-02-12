package client

import (
	"fmt"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
)

func TestNewMainChainElasticClient(t *testing.T) {
	t.Run("nil elastic client, should error", func(t *testing.T) {
		mainChainESClient, err := NewMainChainElasticClient(nil, true)
		require.Error(t, err, dataindexer.ErrNilDatabaseClient)
		require.True(t, mainChainESClient.IsInterfaceNil())
	})
	t.Run("valid elastic client, should work", func(t *testing.T) {
		esClient, err := NewElasticClient(elasticsearch.Config{
			Addresses: []string{"http://localhost:9200"},
		})
		require.Nil(t, err)
		require.NotNil(t, esClient)

		mainChainESClient, err := NewMainChainElasticClient(esClient, true)
		require.NoError(t, err)
		require.Equal(t, "*client.mainChainElasticClient", fmt.Sprintf("%T", mainChainESClient))
	})
}

func TestMainChainElasticClient_IsEnabled(t *testing.T) {
	esClient, err := NewElasticClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	})
	require.Nil(t, err)
	require.NotNil(t, esClient)

	mainChainESClient, err := NewMainChainElasticClient(esClient, true)
	require.NoError(t, err)
	require.Equal(t, true, mainChainESClient.IsEnabled())
}
