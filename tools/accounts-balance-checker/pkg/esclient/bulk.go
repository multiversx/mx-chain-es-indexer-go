package esclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("client")

// DoBulkRequest will do a bulk of request to elastic server
func (ec *esClient) DoBulkRequest(buff *bytes.Buffer, index string) error {
	reader := bytes.NewReader(buff.Bytes())

	options := make([]func(*esapi.BulkRequest), 0)
	if index != "" {
		options = append(options, ec.client.Bulk.WithIndex(index))
	}

	res, err := ec.client.Bulk(
		reader,
		options...,
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
	return extractErrorFromResponseBytes(bodyBytes)
}

// bulkRequestResponse defines the structure of a bulk request response
type bulkRequestResponse struct {
	Errors bool `json:"errors"`
	Items  []struct {
		ItemIndex  *item `json:"index"`
		ItemUpdate *item `json:"update"`
	} `json:"items"`
}

type item struct {
	Index  string `json:"_index"`
	ID     string `json:"_id"`
	Status int    `json:"status"`
	Result string `json:"result"`
	Error  struct {
		Type   string `json:"type"`
		Reason string `json:"reason"`
		Cause  struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		} `json:"caused_by"`
	} `json:"error"`
}

func extractErrorFromResponseBytes(responseBytes []byte) error {
	bulkResponse := &bulkRequestResponse{}
	err := json.Unmarshal(responseBytes, bulkResponse)
	if err != nil {
		return err
	}

	count := 0
	errorsString := ""
	for _, i := range bulkResponse.Items {
		var selectedItem item
		switch {
		case i.ItemIndex != nil:
			selectedItem = *i.ItemIndex
		case i.ItemUpdate != nil:
			selectedItem = *i.ItemUpdate
		}

		log.Debug("worked on", "index", selectedItem.Index,
			"_id", selectedItem.ID,
			"result", selectedItem.Result,
			"status", selectedItem.Status,
		)

		if selectedItem.Status < http.StatusBadRequest {
			continue
		}

		count++
		errorsString += fmt.Sprintf(`{ "index": "%s", "id": "%s", "statusCode": %d, "errorType": "%s", "reason": "%s", "causedBy": { "type": "%s", "reason": "%s" }}\n`,
			selectedItem.Index, selectedItem.ID, selectedItem.Status, selectedItem.Error.Type, selectedItem.Error.Reason, selectedItem.Error.Cause.Type, selectedItem.Error.Cause.Reason)
	}
	if errorsString == "" {
		return nil
	}

	return fmt.Errorf("%s", errorsString)
}
