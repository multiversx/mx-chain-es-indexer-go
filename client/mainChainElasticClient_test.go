package client

import (
	"fmt"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/stretchr/testify/require"

	indexer "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
)

func TestNewMainChainElasticClient(t *testing.T) {
	t.Run("no url, should error", func(t *testing.T) {
		esClient, err := NewMainChainElasticClient(elasticsearch.Config{
			Addresses: []string{},
		}, true)
		require.Nil(t, esClient)
		require.Equal(t, indexer.ErrNoElasticUrlProvided, err)
	})
	t.Run("should work", func(t *testing.T) {
		esClient, err := NewMainChainElasticClient(elasticsearch.Config{
			Addresses: []string{"http://localhost:9200"},
		}, true)
		require.Nil(t, err)
		require.Equal(t, "*client.mainChainElasticClient", fmt.Sprintf("%T", esClient))
	})
}

func TestMainChainElasticClient_IsEnabled(t *testing.T) {
	esClient, err := NewMainChainElasticClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	}, true)
	require.Nil(t, err)
	require.Equal(t, true, esClient.IsEnabled())
}
