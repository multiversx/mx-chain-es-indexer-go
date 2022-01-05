package client

import (
	"bytes"
	"context"
	"fmt"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/tidwall/gjson"
)

var (
	log = logger.GetOrCreate("index-modifier/pkg/client")
)

type esClient struct {
	client *elasticsearch.Client
	// countScroll is used to be incremented after each scroll so the scroll duration is different each time,
	// bypassing any possible caching based on the same request
	countScroll int
}

// NewElasticClient will create a new instance of an esClient
func NewElasticClient(cfg elasticsearch.Config) (*esClient, error) {
	elasticClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &esClient{
		client:      elasticClient,
		countScroll: 0,
	}, nil
}

// DoScrollRequestAllDocuments will perform a documents request using scroll api
func (esc *esClient) DoScrollRequestAllDocuments(
	index string,
	body []byte,
	handlerFunc func(responseBytes []byte) error,
) error {
	esc.countScroll++
	res, err := esc.client.Search(
		esc.client.Search.WithSize(9000),
		esc.client.Search.WithScroll(10*time.Minute+time.Duration(esc.countScroll)*time.Millisecond),
		esc.client.Search.WithContext(context.Background()),
		esc.client.Search.WithIndex(index),
		esc.client.Search.WithBody(bytes.NewBuffer(body)),
	)
	if err != nil {
		return err
	}

	bodyBytes, err := getBytesFromResponse(res)
	if err != nil {
		return err
	}

	err = handlerFunc(bodyBytes)
	if err != nil {
		return err
	}

	scrollID := gjson.Get(string(bodyBytes), "_scroll_id")
	return esc.iterateScroll(scrollID.String(), handlerFunc)
}

func (esc *esClient) iterateScroll(
	scrollID string,
	handlerFunc func(responseBytes []byte) error,
) error {
	if scrollID == "" {
		return nil
	}
	defer func() {
		err := esc.clearScroll(scrollID)
		if err != nil {
			log.Warn("cannot clear scroll", "error", err)
		}
	}()

	for {
		scrollBodyBytes, errScroll := esc.getScrollResponse(scrollID)
		if errScroll != nil {
			return errScroll
		}

		numberOfHits := gjson.Get(string(scrollBodyBytes), "hits.hits.#")
		if numberOfHits.Int() < 1 {
			return nil
		}
		err := handlerFunc(scrollBodyBytes)
		if err != nil {
			return err
		}
	}
}

func (esc *esClient) getScrollResponse(scrollID string) ([]byte, error) {
	esc.countScroll++
	res, err := esc.client.Scroll(
		esc.client.Scroll.WithScrollID(scrollID),
		esc.client.Scroll.WithScroll(2*time.Minute+time.Duration(esc.countScroll)*time.Millisecond),
	)
	if err != nil {
		return nil, err
	}

	return getBytesFromResponse(res)
}

func (esc *esClient) clearScroll(scrollID string) error {
	resp, err := esc.client.ClearScroll(
		esc.client.ClearScroll.WithScrollID(scrollID),
	)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	if resp.IsError() && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error response: %s", resp)
	}

	return nil
}

func getBytesFromResponse(res *esapi.Response) ([]byte, error) {
	if res.IsError() {
		return nil, fmt.Errorf("error response: %s", res)
	}
	defer closeBody(res)

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}

func closeBody(res *esapi.Response) {
	if res != nil && res.Body != nil {
		_ = res.Body.Close()
	}
}
