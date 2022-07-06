package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

func (esc *esClient) DoGetRequest(index string, body []byte, response interface{}, size int) error {
	res, err := esc.client.Search(
		esc.client.Search.WithIndex(index),
		esc.client.Search.WithBody(bytes.NewBuffer(body)),
		esc.client.Search.WithRequestCache(false),
		esc.client.Search.WithSize(size),
		esc.client.Search.WithTimeout(10*time.Minute+time.Duration(esc.updateAndGetCount())*time.Millisecond),
	)
	if err != nil {
		return err
	}
	if res.IsError() || res.StatusCode >= 400 {
		return fmt.Errorf("%s", res.String())
	}

	defer closeBody(res)

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bodyBytes, response)
	if err != nil {
		return err
	}

	return nil
}

func (esc *esClient) updateAndGetCount() int {
	esc.mutex.Lock()
	defer esc.mutex.Unlock()

	esc.countSearch++
	return esc.countSearch
}
