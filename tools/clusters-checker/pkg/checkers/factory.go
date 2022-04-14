package checkers

import (
	"fmt"
	"github.com/ElrondNetwork/elastic-indexer-go/tools/clusters-checker/pkg/client"
	"github.com/ElrondNetwork/elastic-indexer-go/tools/clusters-checker/pkg/config"
	"github.com/elastic/go-elasticsearch/v7"
)

func CreateClusterChecker(cfg *config.Config) (*clusterChecker, error) {
	clientSource, err := client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{cfg.SourceCluster.URL},
		Username:  cfg.SourceCluster.User,
		Password:  cfg.SourceCluster.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create source client %s", err.Error())
	}

	clientDestination, err := client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{cfg.DestinationCluster.URL},
		Username:  cfg.DestinationCluster.User,
		Password:  cfg.DestinationCluster.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create destination client %s", err.Error())
	}

	return &clusterChecker{
		clientSource:      clientSource,
		clientDestination: clientDestination,
		indices:           cfg.Compare.Indices,
	}, nil
}
