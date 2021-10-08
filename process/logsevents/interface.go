package logsevents

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
)

type argsProcessEvent struct {
	txHashHexEncoded string
	pendingBalances  *pendingBalancesProc
	scDeploys        map[string]*data.ScDeployInfo
	event            coreData.EventHandler
	accounts         data.AlteredAccountsHandler
	tokens           data.TokensHandler
	tagsCount        data.CountTags
	timestamp        uint64
	tokensInfo       *[]*data.TokenInfo
}

type eventsProcessor interface {
	processEvent(args *argsProcessEvent) (string, string, bool)
}
