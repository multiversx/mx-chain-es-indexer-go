package check

// ESClientHandler -
type ESClientHandler interface {
	DoScrollRequestAllDocuments(
		index string,
		body []byte,
		handlerFunc func(responseBytes []byte) error,
	) error
	//DoMultiGet(ids []string, index string) ([]byte, error)
}

// RestClientHandler -
type RestClientHandler interface {
	CallGetRestEndPoint(
		path string,
		value interface{},
	) error
}
