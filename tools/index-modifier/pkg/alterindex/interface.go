package alterindex

import "bytes"

type ScrollClient interface {
	DoScrollRequestAllDocuments(
		index string,
		body []byte,
		handlerFunc func(responseBytes []byte) error,
	) error
}

type BulkClient interface {
	DoBulkRequest(buff *bytes.Buffer, index string) error
}
