package client

import (
	"bytes"

	"github.com/tidwall/gjson"
)

// DoCountRequest will get the number of elements that correspond with the provided query
func (esc *esClient) DoCountRequest(index string, body []byte) (uint64, error) {
	res, err := esc.client.Count(
		esc.client.Count.WithIndex(index),
		esc.client.Count.WithBody(bytes.NewBuffer(body)),
	)
	if err != nil {
		return 0, err
	}

	bodyBytes, err := getBytesFromResponse(res)
	if err != nil {
		return 0, err
	}

	countRes := gjson.Get(string(bodyBytes), "count")

	return countRes.Uint(), nil
}
