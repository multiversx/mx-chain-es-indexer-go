package logsevents

import (
	"math/big"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process/elasticproc/converters"
	"github.com/ElrondNetwork/elrond-go-core/core"
)

const minTopicsUpdate = 4

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
	// [3:] --> modified data
	topics := args.event.GetTopics()
	if len(topics) < minTopicsUpdate {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	callerAddress, _ := npp.pubKeyConverter.Encode(args.event.GetAddress())
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

	updateNFT := &data.NFTDataUpdate{
		Identifier: identifier,
		Address:    callerAddress,
	}

	switch eventIdentifier {
	case core.BuiltInFunctionESDTNFTUpdateAttributes:
		updateNFT.NewAttributes = topics[3]
	case core.BuiltInFunctionESDTNFTAddURI:
		updateNFT.URIsToAdd = topics[3:]
	}

	return argOutputProcessEvent{
		processed:     true,
		updatePropNFT: updateNFT,
	}
}
