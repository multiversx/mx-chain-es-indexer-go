package check

import "github.com/ElrondNetwork/elastic-indexer-go/data"

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

type ResponseAccounts struct {
	Hits struct {
		Hits []struct {
			ID     string           `json:"_id"`
			Source data.AccountInfo `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// AccountResponse holds the account endpoint response
type AccountResponse struct {
	Data struct {
		Balance string `json:"balance"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}
