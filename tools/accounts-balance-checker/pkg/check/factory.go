package check

import (
	"github.com/ElrondNetwork/elastic-indexer-go/tools/accounts-balance-checker/pkg/config"
	"github.com/ElrondNetwork/elastic-indexer-go/tools/accounts-balance-checker/pkg/esclient"
	"github.com/ElrondNetwork/elastic-indexer-go/tools/accounts-balance-checker/pkg/rest"
	"github.com/elastic/go-elasticsearch/v7"
)

func CreateBalanceChecker(cfg *config.Config) (*balanceChecker, error) {
	esClient, err := esclient.NewElasticClient(elasticsearch.Config{
		Addresses: []string{cfg.Elasticsearch.URL},
		Username:  cfg.Elasticsearch.Username,
		Password:  cfg.Elasticsearch.Password,
	})
	if err != nil {
		return nil, err
	}

	restClient, err := rest.NewRestClient(cfg.Proxy.URL)
	if err != nil {
		return nil, err
	}

	return NewBalanceChecker(esClient, restClient)
}
