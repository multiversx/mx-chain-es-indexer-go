package client

import (
	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc"
)

type mainChainElasticClient struct {
	elasticproc.DatabaseClientHandler
	indexingEnabled bool
}

// NewMainChainElasticClient creates a new sovereign elastic client
func NewMainChainElasticClient(esClient elasticproc.DatabaseClientHandler, indexingEnabled bool) (*mainChainElasticClient, error) {
	if check.IfNil(esClient) {
		return nil, dataindexer.ErrNilDatabaseClient
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

// IsInterfaceNil returns true if there is no value under the interface
func (mcec *mainChainElasticClient) IsInterfaceNil() bool {
	return mcec == nil
}
