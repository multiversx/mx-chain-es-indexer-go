package logsevents

import (
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
)

const (
	numIssueLogTopics = 4

	issueFungibleESDTFunc     = "issue"
	issueSemiFungibleESDTFunc = "issueSemiFungible"
	issueNonFungibleESDTFunc  = "issueNonFungible"
	registerMetaESDTFunc      = "registerMetaESDT"
	changeSFTToMetaESDTFunc   = "changeSFTToMetaESDT"
	transferOwnershipFunc     = "transferOwnership"
	registerAndSetRolesFunc   = "registerAndSetAllRoles"
)

type esdtIssueProcessor struct {
	pubkeyConverter            core.PubkeyConverter
	issueOperationsIdentifiers map[string]struct{}
}

func newESDTIssueProcessor(pubkeyConverter core.PubkeyConverter) *esdtIssueProcessor {
	return &esdtIssueProcessor{
		pubkeyConverter: pubkeyConverter,
		issueOperationsIdentifiers: map[string]struct{}{
			issueFungibleESDTFunc:     {},
			issueSemiFungibleESDTFunc: {},
			issueNonFungibleESDTFunc:  {},
			registerMetaESDTFunc:      {},
			changeSFTToMetaESDTFunc:   {},
			transferOwnershipFunc:     {},
			registerAndSetRolesFunc:   {},
		},
	}
}

func (iep *esdtIssueProcessor) processEvent(args *argsProcessEvent) argOutputProcessEvent {
	identifierStr := string(args.event.GetIdentifier())
	_, ok := iep.issueOperationsIdentifiers[identifierStr]
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
	// topics[4] -- new owner address in case of transferOwnershipFunc
	if len(topics[0]) == 0 {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	encodedAddr := iep.pubkeyConverter.Encode(args.event.GetAddress())

	tokenInfo := &data.TokenInfo{
		Token:        string(topics[0]),
		Name:         string(topics[1]),
		Ticker:       string(topics[2]),
		Type:         string(topics[3]),
		Issuer:       encodedAddr,
		CurrentOwner: encodedAddr,
		Timestamp:    time.Duration(args.timestamp),
		OwnersHistory: []*data.OwnerData{
			{
				Address:   encodedAddr,
				Timestamp: time.Duration(args.timestamp),
			},
		},
	}

	if identifierStr == transferOwnershipFunc && len(topics) >= numIssueLogTopics+1 {
		newOwner := iep.pubkeyConverter.Encode(topics[4])
		tokenInfo.TransferOwnership = true
		tokenInfo.CurrentOwner = newOwner
		tokenInfo.OwnersHistory[0].Address = newOwner
	}

	return argOutputProcessEvent{
		tokenInfo: tokenInfo,
		processed: true,
	}
}
