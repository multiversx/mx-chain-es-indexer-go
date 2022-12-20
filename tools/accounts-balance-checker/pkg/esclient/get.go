package esclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

var countGet = 0

// DoGetRequest will do a get request
func (ec *esClient) DoGetRequest(buff *bytes.Buffer, index string, response interface{}, size int) error {
	countGet++
	res, err := ec.client.Search(
		ec.client.Search.WithIndex(index),
		ec.client.Search.WithBody(buff),
		ec.client.Search.WithRequestCache(false),
		ec.client.Search.WithSize(size),
		ec.client.Search.WithTimeout(10*time.Minute+time.Duration(countGet)*time.Millisecond),
	)
	if err != nil {
		return err
	}
	if res.IsError() {
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
