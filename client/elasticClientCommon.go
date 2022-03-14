package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func exists(res *esapi.Response, err error) bool {
	defer func() {
		if res != nil && res.Body != nil {
			err = res.Body.Close()
			if err != nil {
				log.Warn("elasticClient.exists", "could not close body: ", err.Error())
			}
		}
	}()

	if err != nil {
		log.Warn("elasticClient.IndexExists", "could not check index on the elastic nodes:", err.Error())
		return false
	}

	switch res.StatusCode {
	case http.StatusOK:
		return true
	case http.StatusNotFound:
		return false
	default:
		log.Warn("elasticClient.exists", "invalid status code returned by the elastic nodes:", res.StatusCode)
		return false
	}
}

func newRequest(method, path string, body *bytes.Buffer) *http.Request {
	r := http.Request{
		Method:     method,
		URL:        &url.URL{Path: path},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}

	if body != nil {
		r.Body = ioutil.NopCloser(body)
		r.ContentLength = int64(body.Len())
	}

	return &r
}

func isErrAlreadyExists(resString string) bool {
	aliasExistsMessage := "invalid_alias_name_exception"
	alreadyExistsAlias := "resource_already_exists_exception"

	return strings.Contains(resString, aliasExistsMessage) || strings.Contains(resString, alreadyExistsAlias)
}

func parseResponse(res *esapi.Response, dstBody ...interface{}) error {
	if res == nil {
		return nil
	}

	if res.IsError() {
		resStr := res.String()
		if isErrAlreadyExists(resStr) {
			return nil
		}

		return fmt.Errorf("%s", res.String())
	}

	defer closeBody(res)

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if dstBody != nil && len(dstBody) > 0 {
		return json.Unmarshal(bodyBytes, &dstBody[0])
	}

	responseBody := map[string]interface{}{}
	err = json.Unmarshal(bodyBytes, &responseBody)
	if err != nil {
		errToReturn := err
		isBackOffError := strings.Contains(string(bodyBytes), fmt.Sprintf("%d", http.StatusForbidden)) ||
			strings.Contains(string(bodyBytes), fmt.Sprintf("%d", http.StatusTooManyRequests))
		if isBackOffError {
			errToReturn = indexer.ErrBackOff
		}

		return fmt.Errorf("%w, cannot unmarshal elastic response body to map[string]interface{}, "+
			"decode error: %s, body response: %s", errToReturn, err.Error(), string(bodyBytes))
	}

	bulkResponse := &BulkRequestResponse{}
	err = json.Unmarshal(bodyBytes, bulkResponse)
	if err != nil {
		return err
	}

	if bulkResponse.Errors {
		return extractErrorFromBulkBodyResponseBytes(bulkResponse)
	}

	return nil
}

func extractErrorFromBulkBodyResponseBytes(response *BulkRequestResponse) error {
	count := 0
	errorsString := ""
	for _, item := range response.Items {
		var selectedItem Item

		switch {
		case item.ItemIndex != nil:
			selectedItem = *item.ItemIndex
		case item.ItemUpdate != nil:
			selectedItem = *item.ItemUpdate
		}

		if selectedItem.Status < http.StatusBadRequest {
			continue
		}

		count++
		errorsString += fmt.Sprintf(`{ "id": "%s", "statusCode": %d, "errorType": "%s", "reason": "%s", "causedBy": { "type": "%s", "reason": "%s" }}\n`,
			selectedItem.ID, selectedItem.Status, selectedItem.Error.Type, selectedItem.Error.Reason, selectedItem.Error.Cause.Type, selectedItem.Error.Cause.Reason)

		if count == numOfErrorsToExtractBulkResponse {
			break
		}
	}

	return fmt.Errorf("%s", errorsString)
}
