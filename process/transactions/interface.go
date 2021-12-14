package transactions

import "github.com/ElrondNetwork/elastic-indexer-go/process/transactions/datafield"

type DataFieldParser interface {
	Parse(dataField []byte, sender, receiver []byte) *datafield.ResponseParseData
}
