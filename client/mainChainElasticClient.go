package client

import (
	"github.com/elastic/go-elasticsearch/v7"
)

type mainChainElasticClient struct {
	*elasticClient
	indexingEnabled bool
}

// NewMainChainElasticClient creates a new sovereign elastic client
func NewMainChainElasticClient(cfg elasticsearch.Config, indexingEnabled bool) (*mainChainElasticClient, error) {
	esClient, err := NewElasticClient(cfg)
	if err != nil {
		return nil, err
	}

	return &mainChainElasticClient{
		esClient,
		indexingEnabled,
	}, nil
}

// IsEnabled returns true if main chain elastic client is enabled
func (mcec *mainChainElasticClient) IsEnabled() bool {
	return mcec.indexingEnabled
}
