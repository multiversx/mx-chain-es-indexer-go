package alterindex

import "bytes"

// ScrollClient defines what a scroll client should do
type ScrollClient interface {
	DoScrollRequestAllDocuments(
		index string,
		body []byte,
		handlerFunc func(responseBytes []byte) error,
	) error
}

// BulkClient defines what s bulk client should do
type BulkClient interface {
	DoBulkRequest(buff *bytes.Buffer, index string) error
}
