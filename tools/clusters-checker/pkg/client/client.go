package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/tidwall/gjson"
)

var (
	log = logger.GetOrCreate("clusters-checker/pkg/client")

	httpStatusesForRetry = []int{429, 502, 503, 504}
)

type esClient struct {
	client *elasticsearch.Client
	// countScroll is used to be incremented after each scroll so the scroll duration is different each time,
	// bypassing any possible caching based on the same request
	countScroll int
	countSearch int
	mutex       sync.Mutex
}

// NewElasticClient will create a new instance of an esClient
func NewElasticClient(cfg elasticsearch.Config) (*esClient, error) {
	if len(cfg.RetryOnStatus) == 0 {
		cfg.RetryOnStatus = httpStatusesForRetry
		cfg.RetryBackoff = func(i int) time.Duration {
			// A simple exponential delay
			d := time.Duration(math.Exp2(float64(i))) * time.Second
			log.Info("elastic: retry backoff", "attempt", i, "sleep duration", d)
			return d
		}
		cfg.MaxRetries = 5
	}

	elasticClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &esClient{
		client:      elasticClient,
		countScroll: 0,
		mutex:       sync.Mutex{},
	}, nil
}

func (esc *esClient) InitializeScroll(index string, body []byte, response interface{}) (string, bool, error) {
	res, err := esc.client.Search(
		esc.client.Search.WithSize(9000),
		esc.client.Search.WithScroll(10*time.Minute+time.Duration(esc.updateAndGetCountScroll())*time.Millisecond),
		esc.client.Search.WithIndex(index),
		esc.client.Search.WithBody(bytes.NewBuffer(body)),
	)
	if err != nil {
		return "", false, err
	}
	if res.IsError() || res.StatusCode >= 400 {
		return "", false, fmt.Errorf("%s", res.String())
	}

	bodyBytes, err := getBytesFromResponse(res)
	if err != nil {
		return "", false, err
	}
	scrollID := gjson.Get(string(bodyBytes), "_scroll_id").String()
	numberOfHits := gjson.Get(string(bodyBytes), "hits.hits.#")
	isDone := numberOfHits.Int() == 0

	if isDone {
		defer func() {
			errC := esc.clearScroll(scrollID)
			if errC != nil {
				log.Warn("cannot clear scroll", "error", errC)
			}
		}()
	}

	err = json.Unmarshal(bodyBytes, response)
	if err != nil {
		return "", false, err
	}

	return scrollID, isDone, nil
}

func (esc *esClient) DoScrollRequestV2(scrollID string, response interface{}) (string, bool, error) {
	defer logExecutionTime(time.Now(), "esClient.DoScrollRequestV2")

	res, err := esc.client.Scroll(
		esc.client.Scroll.WithScrollID(scrollID),
		esc.client.Scroll.WithScroll(2*time.Minute+time.Duration(esc.updateAndGetCountScroll())*time.Millisecond),
	)
	if err != nil {
		return "", false, err
	}

	bodyBytes, err := getBytesFromResponse(res)
	if err != nil {
		return "", false, err
	}

	nextScrollID := gjson.Get(string(bodyBytes), "_scroll_id").String()
	numberOfHits := gjson.Get(string(bodyBytes), "hits.hits.#")
	isDone := numberOfHits.Int() == 0

	if isDone {
		defer func() {
			errC := esc.clearScroll(scrollID)
			if errC != nil {
				log.Warn("cannot clear scroll", "error", errC)
			}
		}()
	}

	err = json.Unmarshal(bodyBytes, response)
	if err != nil {
		return "", false, err
	}

	return nextScrollID, isDone, nil
}

func logExecutionTime(start time.Time, message string) {
	log.Debug(message, "duration in seconds", time.Since(start).Seconds())
}

// DoScrollRequestAllDocuments will perform a documents request using scroll api
func (esc *esClient) DoScrollRequestAllDocuments(
	index string,
	body []byte,
	handlerFunc func(responseBytes []byte) error,
) error {
	res, err := esc.client.Search(
		esc.client.Search.WithSize(9000),
		esc.client.Search.WithScroll(10*time.Minute+time.Duration(esc.updateAndGetCountScroll())*time.Millisecond),
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
	res, err := esc.client.Scroll(
		esc.client.Scroll.WithScrollID(scrollID),
		esc.client.Scroll.WithScroll(2*time.Minute+time.Duration(esc.updateAndGetCountScroll())*time.Millisecond),
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

func (esc *esClient) updateAndGetCountScroll() int {
	esc.mutex.Lock()
	defer esc.mutex.Unlock()

	esc.countScroll += 1 + rand.Intn(10)
	return esc.countScroll
}
