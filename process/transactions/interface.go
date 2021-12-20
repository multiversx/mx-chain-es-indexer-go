package transactions

import "github.com/ElrondNetwork/elastic-indexer-go/process/transactions/datafield"

// DataFieldParser defines what a data field parser should be able to do
type DataFieldParser interface {
	Parse(dataField []byte, sender, receiver []byte) *datafield.ResponseParseData
}
