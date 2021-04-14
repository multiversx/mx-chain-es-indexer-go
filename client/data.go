package client

const (
	headerXSRF                       = "kbn-xsrf"
	headerContentType                = "Content-Type"
	kibanaPluginPath                 = "_plugin/kibana/api"
	numOfErrorsToExtractBulkResponse = 5
)

var headerContentTypeJSON = []string{"application/json"}

// BulkRequestResponse defines the structure of a bulk request response
type BulkRequestResponse struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Index struct {
			Status int `json:"status"`
			Error  struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
			} `json:"error"`
		} `json:"index"`
	} `json:"items"`
}
