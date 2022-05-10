package check

import "bytes"

// ESClientHandler -
type ESClientHandler interface {
	DoScrollRequestAllDocuments(
		index string,
		body []byte,
		handlerFunc func(responseBytes []byte) error,
	) error
	DoGetRequest(buff *bytes.Buffer, index string, response interface{}, size int) error
}

// RestClientHandler -
type RestClientHandler interface {
	CallGetRestEndPoint(
		path string,
		value interface{},
	) error
}
