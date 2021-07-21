package logsevents

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process/tags"
	nodeData "github.com/ElrondNetwork/elrond-go-core/data"
)

type argsProcessEvent struct {
	event     nodeData.EventHandler
	accounts  data.AlteredAccountsHandler
	tokens    data.TokensHandler
	tagsCount tags.CountTags
	timestamp uint64
}

type eventsProcessor interface {
	processEvent(args *argsProcessEvent) (string, bool)
}
