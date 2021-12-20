package datafield

import (
	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

// ResponseParseData is the response with results after the data field was parsed
type ResponseParseData struct {
	Operation        string
	Function         string
	ESDTValues       []string
	Tokens           []string
	Receivers        []string
	ReceiversShardID []uint32
}

// ArgsOperationDataFieldParser holds all the components required to create a new instance of data field parser
type ArgsOperationDataFieldParser struct {
	PubKeyConverter  core.PubkeyConverter
	Marshalizer      marshal.Marshalizer
	ShardCoordinator indexer.ShardCoordinator
}
