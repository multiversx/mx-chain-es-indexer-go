package checkers

// ESClient defines what an ES client should do
type ESClient interface {
	InitializeScroll(index string, body []byte, response interface{}) (string, bool, error)
	DoScrollRequestV2(scrollID string, response interface{}) (string, bool, error)

	DoCountRequest(index string, body []byte) (uint64, error)
	DoGetRequest(index string, body []byte, response interface{}, size int) error
	DoScrollRequestAllDocuments(
		index string,
		body []byte,
		handlerFunc func(responseBytes []byte) error,
		size int,
	) error
}

type Checker interface {
	CompareIndicesNoTimestamp() error
	CompareIndicesWithTimestamp() error
}
