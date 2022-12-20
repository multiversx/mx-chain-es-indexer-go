package esclient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/tidwall/gjson"
)

type esClient struct {
	client      *elasticsearch.Client
	countScroll int
}

// NewElasticClient will create a new instance of esClient
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
func (ec *esClient) DoScrollRequestAllDocuments(
	index string,
	body []byte,
	handlerFunc func(responseBytes []byte) error,
) error {
	ec.countScroll++
	res, err := ec.client.Search(
		ec.client.Search.WithSize(9000),
		ec.client.Search.WithScroll(10*time.Minute+time.Duration(ec.countScroll)*time.Millisecond),
		ec.client.Search.WithContext(context.Background()),
		ec.client.Search.WithIndex(index),
		ec.client.Search.WithBody(bytes.NewBuffer(body)),
	)
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("error DoScrollRequestAllDocuments: %s", res.String())
	}

	bodyBytes, errGet := getBytesFromResponse(res)
	if errGet != nil {
		return errGet
	}

	err = handlerFunc(bodyBytes)
	if err != nil {
		return err
	}

	scrollID := gjson.Get(string(bodyBytes), "_scroll_id")
	return ec.iterateScroll(scrollID.String(), handlerFunc)
}

func (ec *esClient) iterateScroll(
	scrollID string,
	handlerFunc func(responseBytes []byte) error,
) error {
	if scrollID == "" {
		return nil
	}
	defer func() {
		_ = ec.clearScroll(scrollID)
	}()

	for {
		scrollBodyBytes, errScroll := ec.getScrollResponse(scrollID)
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

func (ec *esClient) getScrollResponse(scrollID string) ([]byte, error) {
	ec.countScroll++
	res, err := ec.client.Scroll(
		ec.client.Scroll.WithScrollID(scrollID),
		ec.client.Scroll.WithScroll(2*time.Minute+time.Duration(ec.countScroll)*time.Millisecond),
	)
	if err != nil {
		return nil, err
	}

	return getBytesFromResponse(res)
}

func (ec *esClient) clearScroll(scrollID string) error {
	resp, err := ec.client.ClearScroll(
		ec.client.ClearScroll.WithScrollID(scrollID),
	)
	if err != nil {
		return err
	}
	if resp.IsError() && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error response: %s", resp.String())
	}

	defer closeBody(resp)

	return nil
}
