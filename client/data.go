package client

const (
	headerXSRF                       = "kbn-xsrf"
	headerContentType                = "Content-Type"
	kibanaPluginPath                 = "_plugin/kibana/api"
	numOfErrorsToExtractBulkResponse = 5
)

var headerContentTypeJSON = []string{"application/json"}

// BulkRequestResponse defines the structure of a bulk request response index
type BulkRequestResponseIndex struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Item Item `json:"index"`
	} `json:"items"`
}

// BulkRequestResponse defines the structure of a bulk request response update
type BulkRequestResponseUpdate struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Item Item `json:"update"`
	} `json:"items"`
}

// Item defines the structure of a item from a bulk response
type Item struct {
	Status int `json:"status"`
	Error  struct {
		Type   string `json:"type"`
		Reason string `json:"reason"`
	} `json:"error"`
}
