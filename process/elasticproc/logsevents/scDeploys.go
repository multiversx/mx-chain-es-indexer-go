package logsevents

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

const (
	numTopicsChangeOwner   = 1
	minTopicsContractEvent = 3
)

type scDeploysProcessor struct {
	scDeploysIdentifiers map[string]struct{}
	pubKeyConverter      core.PubkeyConverter
}

func newSCDeploysProcessor(pubKeyConverter core.PubkeyConverter) *scDeploysProcessor {
	return &scDeploysProcessor{
		pubKeyConverter: pubKeyConverter,
		scDeploysIdentifiers: map[string]struct{}{
			core.SCDeployIdentifier:                {},
			core.SCUpgradeIdentifier:               {},
			core.BuiltInFunctionChangeOwnerAddress: {},
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
	isChangeOwnerEvent := len(topics) == numTopicsChangeOwner && eventIdentifier == core.BuiltInFunctionChangeOwnerAddress
	if isChangeOwnerEvent {
		return sdp.processChangeOwnerEvent(args)
	}

	if len(topics) < minTopicsContractEvent {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	scAddress := sdp.pubKeyConverter.SilentEncode(topics[0], log)
	creatorAddress := sdp.pubKeyConverter.SilentEncode(topics[1], log)

	args.scDeploys[scAddress] = &data.ScDeployInfo{
		TxHash:       args.txHashHexEncoded,
		Creator:      creatorAddress,
		CurrentOwner: creatorAddress,
		CodeHash:     topics[2],
		Timestamp:    args.timestamp,
		TimestampMs:  args.timestampMs,
	}

	return argOutputProcessEvent{
		processed: true,
	}
}

func (sdp *scDeploysProcessor) processChangeOwnerEvent(args *argsProcessEvent) argOutputProcessEvent {
	scAddress := sdp.pubKeyConverter.SilentEncode(args.event.GetAddress(), log)
	newOwner := sdp.pubKeyConverter.SilentEncode(args.event.GetTopics()[0], log)
	args.changeOwnerOperations[scAddress] = &data.OwnerData{
		TxHash:      args.txHashHexEncoded,
		Address:     newOwner,
		Timestamp:   time.Duration(args.timestamp),
		TimestampMs: time.Duration(args.timestampMs),
	}

	return argOutputProcessEvent{
		processed: true,
	}
}
