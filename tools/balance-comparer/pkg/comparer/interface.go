package comparer

// ScrollClient defines what a scroll client should do
type ScrollClient interface {
	DoScrollRequestAllDocuments(
		index string,
		body []byte,
		handlerFunc func(responseBytes []byte) error,
	) error
}
