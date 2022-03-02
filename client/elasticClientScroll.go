package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/tidwall/gjson"
)

// DoScrollRequest will perform a documents request using scroll api
func (ec *elasticClient) DoScrollRequest(
	index string,
	body []byte,
	handlerFunc func(responseBytes []byte) error,
) error {
	ec.countScroll++
	res, err := ec.es.Search(
		ec.es.Search.WithSize(9000),
		ec.es.Search.WithScroll(10*time.Minute+time.Duration(ec.countScroll)*time.Millisecond),
		ec.es.Search.WithContext(context.Background()),
		ec.es.Search.WithIndex(index),
		ec.es.Search.WithBody(bytes.NewBuffer(body)),
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
	return ec.iterateScroll(scrollID.String(), handlerFunc)
}

func (ec *elasticClient) iterateScroll(
	scrollID string,
	handlerFunc func(responseBytes []byte) error,
) error {
	if scrollID == "" {
		return nil
	}
	defer func() {
		err := ec.clearScroll(scrollID)
		if err != nil {
			log.Warn("cannot clear scroll", "error", err)
		}
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

func (ec *elasticClient) getScrollResponse(scrollID string) ([]byte, error) {
	ec.countScroll++
	res, err := ec.es.Scroll(
		ec.es.Scroll.WithScrollID(scrollID),
		ec.es.Scroll.WithScroll(2*time.Minute+time.Duration(ec.countScroll)*time.Millisecond),
	)
	if err != nil {
		return nil, err
	}

	return getBytesFromResponse(res)
}

func (ec *elasticClient) clearScroll(scrollID string) error {
	resp, err := ec.es.ClearScroll(
		ec.es.ClearScroll.WithScrollID(scrollID),
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
