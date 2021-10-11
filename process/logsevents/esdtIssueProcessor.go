package logsevents

import (
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

const (
	numIssueLogTopics = 4

	issueFungibleESDTFunc     = "issue"
	issueSemiFungibleESDTFunc = "issueSemiFungible"
	issueNonFungibleESDTFunc  = "issueNonFungible"
	registerMetaESDTFunc      = "registerMetaESDT"
	changeSFTToMetaESDTFunc   = "changeSFTToMetaESDT"
)

type esdtIssueProcessor struct {
	issueOperationsIdentifiers map[string]struct{}
}

func newESDTIssueProcessor() *esdtIssueProcessor {
	return &esdtIssueProcessor{
		issueOperationsIdentifiers: map[string]struct{}{
			issueFungibleESDTFunc:     {},
			issueSemiFungibleESDTFunc: {},
			issueNonFungibleESDTFunc:  {},
			registerMetaESDTFunc:      {},
			changeSFTToMetaESDTFunc:   {},
		},
	}
}

func (iep *esdtIssueProcessor) processEvent(args *argsProcessEvent) argOutputProcessEvent {
	identifier := args.event.GetIdentifier()
	_, ok := iep.issueOperationsIdentifiers[string(identifier)]
	if !ok {
		return argOutputProcessEvent{}
	}

	topics := args.event.GetTopics()
	if len(topics) < numIssueLogTopics {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	// topics slice contains:
	// topics[0] -- token identifier
	// topics[1] -- token name
	// topics[2] -- token ticker
	// topics[3] -- token type
	if len(topics[0]) == 0 {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	tokenInfo := &data.TokenInfo{
		Token:     string(topics[0]),
		Name:      string(topics[1]),
		Ticker:    string(topics[2]),
		Type:      string(topics[3]),
		Timestamp: time.Duration(args.timestamp),
	}

	return argOutputProcessEvent{
		tokenInfo: tokenInfo,
	}
}
