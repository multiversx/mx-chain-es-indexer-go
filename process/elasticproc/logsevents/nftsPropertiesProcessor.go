package logsevents

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
)

const minTopicsUpdate = 4

type nftsPropertiesProc struct {
	marshaller                 marshal.Marshalizer
	pubKeyConverter            core.PubkeyConverter
	propertiesChangeOperations map[string]struct{}
}

func newNFTsPropertiesProcessor(pubKeyConverter core.PubkeyConverter, marshaller marshal.Marshalizer) *nftsPropertiesProc {
	return &nftsPropertiesProc{
		marshaller:      marshaller,
		pubKeyConverter: pubKeyConverter,
		propertiesChangeOperations: map[string]struct{}{
			core.BuiltInFunctionESDTNFTAddURI:           {},
			core.BuiltInFunctionESDTNFTUpdateAttributes: {},
			core.BuiltInFunctionESDTFreeze:              {},
			core.BuiltInFunctionESDTUnFreeze:            {},
			core.BuiltInFunctionESDTPause:               {},
			core.BuiltInFunctionESDTUnPause:             {},
			core.ESDTMetaDataRecreate:                   {},
			core.ESDTMetaDataUpdate:                     {},
			core.ESDTSetNewURIs:                         {},
			core.ESDTModifyCreator:                      {},
			core.ESDTModifyRoyalties:                    {},
		},
	}
}

func (npp *nftsPropertiesProc) processEvent(args *argsProcessEvent) argOutputProcessEvent {
	//nolint
	eventIdentifier := string(args.event.GetIdentifier())
	_, ok := npp.propertiesChangeOperations[eventIdentifier]
	if !ok {
		return argOutputProcessEvent{}
	}

	callerAddress := npp.pubKeyConverter.SilentEncode(args.event.GetAddress(), log)
	if callerAddress == "" {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	topics := args.event.GetTopics()
	if len(topics) == 1 {
		return npp.processPauseAndUnPauseEvent(eventIdentifier, string(topics[0]))
	}

	// topics contains:
	// [0] --> token identifier
	// [1] --> nonce of the NFT (bytes)
	// [2] --> value
	// [3:] --> modified data
	// [3] --> ESDT token data in case of ESDTMetaDataRecreate

	isModifyCreator := len(topics) == minTopicsUpdate-1 && eventIdentifier == core.ESDTModifyCreator
	if len(topics) < minTopicsUpdate && !isModifyCreator {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	callerAddress = npp.pubKeyConverter.SilentEncode(args.event.GetAddress(), log)
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
	case core.ESDTSetNewURIs:
		updateNFT.SetURIs = true
		updateNFT.URIsToAdd = topics[3:]
	case core.BuiltInFunctionESDTFreeze:
		updateNFT.Freeze = true
	case core.BuiltInFunctionESDTUnFreeze:
		updateNFT.UnFreeze = true
	case core.ESDTMetaDataRecreate, core.ESDTMetaDataUpdate:
		npp.processMetaDataUpdate(updateNFT, topics[3])
	case core.ESDTModifyCreator:
		updateNFT.NewCreator = callerAddress
	case core.ESDTModifyRoyalties:
		newRoyalties := uint32(big.NewInt(0).SetBytes(topics[3]).Uint64())
		updateNFT.NewRoyalties = core.OptionalUint32{
			Value:    newRoyalties,
			HasValue: true,
		}
	}

	return argOutputProcessEvent{
		processed:     true,
		updatePropNFT: updateNFT,
	}
}

func (npp *nftsPropertiesProc) processMetaDataUpdate(updateNFT *data.NFTDataUpdate, esdtTokenBytes []byte) {
	esdtToken := &esdt.ESDigitalToken{}
	err := npp.marshaller.Unmarshal(esdtToken, esdtTokenBytes)
	if err != nil {
		log.Warn("nftsPropertiesProc.processMetaDataRecreate() cannot urmarshal", "error", err.Error())
		return
	}

	tokenMetaData := converters.PrepareTokenMetaData(convertMetaData(npp.pubKeyConverter, esdtToken.TokenMetaData))
	updateNFT.NewMetaData = tokenMetaData
}

func (npp *nftsPropertiesProc) processPauseAndUnPauseEvent(eventIdentifier string, token string) argOutputProcessEvent {
	var updateNFT *data.NFTDataUpdate

	switch eventIdentifier {
	case core.BuiltInFunctionESDTPause:
		updateNFT = &data.NFTDataUpdate{
			Identifier: token,
			Pause:      true,
		}
	case core.BuiltInFunctionESDTUnPause:
		updateNFT = &data.NFTDataUpdate{
			Identifier: token,
			UnPause:    true,
		}
	}

	return argOutputProcessEvent{
		processed:     true,
		updatePropNFT: updateNFT,
	}
}
