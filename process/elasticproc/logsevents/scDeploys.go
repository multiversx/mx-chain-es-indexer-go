package logsevents

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
)

type scDeploysProcessor struct {
	scDeploysIdentifiers map[string]struct{}
	pubKeyConverter      core.PubkeyConverter
}

func newSCDeploysProcessor(pubKeyConverter core.PubkeyConverter) *scDeploysProcessor {
	return &scDeploysProcessor{
		pubKeyConverter: pubKeyConverter,
		scDeploysIdentifiers: map[string]struct{}{
			core.SCDeployIdentifier:  {},
			core.SCUpgradeIdentifier: {},
		},
	}
}

func (sdp *scDeploysProcessor) processEvent(args *argsProcessEvent) argOutputProcessEvent {
	eventIdentifier := string(args.event.GetIdentifier())
	_, ok := sdp.scDeploysIdentifiers[eventIdentifier]
	if !ok {
		return argOutputProcessEvent{}
	}

	topics := args.event.GetTopics()
	if len(topics) < 2 {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	scAddress, _ := sdp.pubKeyConverter.Encode(topics[0])
	creatorAddress, _ := sdp.pubKeyConverter.Encode(topics[1])

	args.scDeploys[scAddress] = &data.ScDeployInfo{
		TxHash:    args.txHashHexEncoded,
		Creator:   creatorAddress,
		Timestamp: args.timestamp,
	}

	return argOutputProcessEvent{
		processed: true,
	}
}
