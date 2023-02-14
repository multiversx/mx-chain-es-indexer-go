package transactions

import datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"

// DataFieldParser defines what a data field parser should be able to do
type DataFieldParser interface {
	Parse(dataField []byte, sender, receiver []byte, numOfShards uint32) *datafield.ResponseParseData
}
