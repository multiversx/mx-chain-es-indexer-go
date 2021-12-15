package alterindex

import (
	"bytes"
	"fmt"
	"math"
	"time"

	client2 "github.com/ElrondNetwork/elastic-indexer-go/client"
	"github.com/ElrondNetwork/elastic-indexer-go/tools/index-modifier/pkg/client"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/elastic/go-elasticsearch/v7"
)

const (
	queryMatchAll = `{"query":{"match_all": {}}}`
)

var log = logger.GetOrCreate("index-modifier/pkg/alterindex")

type indexModifier struct {
	scrollClient ScrollClient
	bulkClient   BulkClient
}

func backOff(i int) time.Duration {
	// A simple exponential delay
	d := time.Duration(math.Exp2(float64(i))) * time.Second
	log.Info("elastic: retry backoff", "attempt", i, "sleep duration", d)
	return d
}

func CreateIndexModifier(scrollClientAddress, bulkClientAddress string) (*indexModifier, error) {
	cfg := elasticsearch.Config{
		Addresses:     []string{scrollClientAddress},
		MaxRetries:    0,
		RetryBackoff:  backOff,
		RetryOnStatus: []int{429, 502, 503, 504},
	}
	scrollClient, err := client.NewElasticClient(cfg)
	if err != nil {
		return nil, err
	}

	cfg.Addresses = []string{bulkClientAddress}
	bulkClient, err := client2.NewElasticClient(cfg)
	if err != nil {
		return nil, err
	}

	return &indexModifier{
		scrollClient: scrollClient,
		bulkClient:   bulkClient,
	}, nil
}

func (im *indexModifier) AlterIndex(index string, modifier func(responseBytes []byte) ([]*bytes.Buffer, error)) error {
	count := 0
	handlerFunc := func(responseBytes []byte) error {
		count++
		dataBuffers, err := modifier(responseBytes)
		if err != nil {
			return fmt.Errorf("%w while preparing data for indexing", err)
		}

		for i := 0; i < len(dataBuffers); i++ {
			err = im.bulkClient.DoBulkRequest(dataBuffers[i], index)
			if err != nil {
				return fmt.Errorf("%w while r.destinationElastic.DoBulkRequest", err)
			}
		}

		log.Info("Do bulk request...", "count", count)

		return nil
	}

	err := im.scrollClient.DoScrollRequestAllDocuments(index, []byte(queryMatchAll), handlerFunc)
	if err != nil {
		return fmt.Errorf("%w while r.sourceElastic.DoScrollRequestAllDocuments", err)
	}

	return nil
}
