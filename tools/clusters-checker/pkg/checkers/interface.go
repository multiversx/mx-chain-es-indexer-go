package checkers

// ESClient defines what a ES client should do
type ESClient interface {
	DoCountRequest(index string, body []byte) (uint64, error)
	DoScrollRequestAllDocuments(
		index string,
		body []byte,
		handlerFunc func(responseBytes []byte) error,
	) error
}
