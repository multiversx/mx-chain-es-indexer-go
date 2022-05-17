package check

import "github.com/ElrondNetwork/elastic-indexer-go/data"

type ResponseTransactions struct {
	Hits struct {
		Hits []struct {
			ID     string           `json:"_id"`
			Source data.Transaction `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

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

type BalancesESDTResponse struct {
	Data struct {
		ESDTS     map[string]*esdtNFTTokenData `json:"esdts"`
		TokenData *esdtNFTTokenData            `json:"tokenData"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

type esdtNFTTokenData struct {
	TokenIdentifier string   `json:"tokenIdentifier"`
	Balance         string   `json:"balance"`
	Properties      string   `json:"properties,omitempty"`
	Name            string   `json:"name,omitempty"`
	Nonce           uint64   `json:"nonce,omitempty"`
	Creator         string   `json:"creator,omitempty"`
	Royalties       string   `json:"royalties,omitempty"`
	Hash            []byte   `json:"hash,omitempty"`
	URIs            [][]byte `json:"uris,omitempty"`
	Attributes      []byte   `json:"attributes,omitempty"`
}
