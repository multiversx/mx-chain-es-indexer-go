package logsevents

import (
	"math/big"

	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
)

const minTopicsUpdate = 5

type nftsPropertiesProc struct {
	pubKeyConverter            core.PubkeyConverter
	propertiesChangeOperations map[string]struct{}
}

func newNFTsPropertiesProcessor(pubKeyConverter core.PubkeyConverter) *nftsPropertiesProc {
	return &nftsPropertiesProc{
		pubKeyConverter: pubKeyConverter,
		propertiesChangeOperations: map[string]struct{}{
			core.BuiltInFunctionESDTNFTAddURI:           {},
			core.BuiltInFunctionESDTNFTUpdateAttributes: {},
		},
	}
}

func (npp *nftsPropertiesProc) processEvent(args *argsProcessEvent) argOutputProcessEvent {
	eventIdentifier := string(args.event.GetIdentifier())
	_, ok := npp.propertiesChangeOperations[eventIdentifier]
	if !ok {
		return argOutputProcessEvent{}
	}

	// topics contains:
	// [0] --> token identifier
	// [1] --> nonce of the NFT (bytes)
	// [2] --> value
	// [3] --> caller address
	// [4:] --> modified data
	topics := args.event.GetTopics()
	if len(topics) < minTopicsUpdate {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	callerAddress := npp.pubKeyConverter.Encode(topics[3])
	if callerAddress == "" {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	nonceBig := big.NewInt(0).SetBytes(topics[1])
	if nonceBig.Uint64() == 0 {
		// this is a fungible token so we should return
		return argOutputProcessEvent{}
	}

	token := string(topics[0])
	identifier := converters.ComputeTokenIdentifier(token, nonceBig.Uint64())

	updateNFT := &data.UpdateNFTData{
		ID:      identifier,
		Address: callerAddress,
	}

	switch eventIdentifier {
	case core.BuiltInFunctionESDTNFTUpdateAttributes:
		updateNFT.NewAttributes = topics[4]
	case core.BuiltInFunctionESDTNFTAddURI:
		updateNFT.URIsToAdd = topics[4:]
	}

	return argOutputProcessEvent{
		processed:     true,
		identifier:    identifier,
		updatePropNFT: updateNFT,
	}
}
