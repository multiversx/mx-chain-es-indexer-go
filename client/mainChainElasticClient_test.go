package client

import (
	"fmt"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/stretchr/testify/require"
)

func TestNewMainChainElasticClient(t *testing.T) {
	esClient, err := NewElasticClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	})
	require.Nil(t, err)
	require.NotNil(t, esClient)

	mainChainESClient, err := NewMainChainElasticClient(esClient, true)
	require.NoError(t, err)
	require.Equal(t, "*client.mainChainElasticClient", fmt.Sprintf("%T", mainChainESClient))
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
