package client

const (
	headerXSRF                       = "kbn-xsrf"
	headerContentType                = "Content-Type"
	numOfErrorsToExtractBulkResponse = 5
)

var headerContentTypeJSON = []string{"application/json"}

// BulkRequestResponse defines the structure of a bulk request response
type BulkRequestResponse struct {
	Errors bool `json:"errors"`
	Items  []struct {
		ItemIndex  *Item `json:"index"`
		ItemUpdate *Item `json:"update"`
	} `json:"items"`
}

// Item defines the structure of a item from a bulk response
type Item struct {
	ID     string `json:"_id"`
	Result string `json:"result"`
	Status int    `json:"status"`
	Error  struct {
		Type   string `json:"type"`
		Reason string `json:"reason"`
		Cause  struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		} `json:"caused_by"`
	} `json:"error"`
}
