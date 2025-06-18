package logsevents

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/sharding"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
	logger "github.com/multiversx/mx-chain-logger-go"
	"math/big"
)

const (
	numTopicsWithReceiverAddress = 4
)

var log = logger.GetOrCreate("indexer/process/logsevents")

type nftsProcessor struct {
	pubKeyConverter          core.PubkeyConverter
	nftOperationsIdentifiers map[string]struct{}
	marshalizer              marshal.Marshalizer
}

func newNFTsProcessor(
	pubKeyConverter core.PubkeyConverter,
	marshalizer marshal.Marshalizer,
) *nftsProcessor {
	return &nftsProcessor{
		pubKeyConverter: pubKeyConverter,
		marshalizer:     marshalizer,
		nftOperationsIdentifiers: map[string]struct{}{
			core.BuiltInFunctionESDTNFTBurn:   {},
			core.BuiltInFunctionESDTNFTCreate: {},
			core.BuiltInFunctionESDTWipe:      {},
		},
	}
}

func (np *nftsProcessor) processEvent(args *argsProcessEvent) argOutputProcessEvent {
	eventIdentifier := string(args.event.GetIdentifier())
	_, ok := np.nftOperationsIdentifiers[eventIdentifier]
	if !ok {
		return argOutputProcessEvent{}
	}

	// topics contains:
	// [0] --> token identifier
	// [1] --> nonce of the NFT (bytes)
	// [2] --> value
	// [3] --> receiver NFT address in case of NFTTransfer
	//     --> ESDT token data in case of NFTCreate
	topics := args.event.GetTopics()
	nonceBig := big.NewInt(0).SetBytes(topics[1])
	if nonceBig.Uint64() == 0 {
		// this is a fungible token so we should return
		return argOutputProcessEvent{}
	}

	sender := args.event.GetAddress()
	senderShardID := sharding.ComputeShardID(sender, args.numOfShards)
	if senderShardID == args.selfShardID {
		np.processNFTEventOnSender(args.event, args.tokens, args.tokensSupply, args.timestamp, args.timestampMs)
	}

	token := string(topics[0])
	identifier := converters.ComputeTokenIdentifier(token, nonceBig.Uint64())

	if !np.shouldAddReceiverData(args) {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	receiver := args.event.GetTopics()[3]
	receiverShardID := sharding.ComputeShardID(receiver, args.numOfShards)
	if receiverShardID != args.selfShardID {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	if eventIdentifier == core.BuiltInFunctionESDTWipe {
		args.tokensSupply.Add(&data.TokenInfo{
			Token:       token,
			Identifier:  identifier,
			Timestamp:   args.timestamp,
			TimestampMs: args.timestampMs,
			Nonce:       nonceBig.Uint64(),
		})
	}

	return argOutputProcessEvent{
		processed: true,
	}
}

func (np *nftsProcessor) shouldAddReceiverData(args *argsProcessEvent) bool {
	eventIdentifier := string(args.event.GetIdentifier())
	isWrongIdentifier := eventIdentifier != core.BuiltInFunctionESDTNFTTransfer &&
		eventIdentifier != core.BuiltInFunctionMultiESDTNFTTransfer && eventIdentifier != core.BuiltInFunctionESDTWipe

	if isWrongIdentifier || len(args.event.GetTopics()) < numTopicsWithReceiverAddress {
		return false
	}

	return true
}

func (np *nftsProcessor) processNFTEventOnSender(
	event coreData.EventHandler,
	tokensCreateInfo data.TokensHandler,
	tokensSupply data.TokensHandler,
	timestamp uint64,
	timestampMs uint64,
) {
	topics := event.GetTopics()
	token := string(topics[0])
	nonceBig := big.NewInt(0).SetBytes(topics[1])
	eventIdentifier := string(event.GetIdentifier())
	if eventIdentifier == core.BuiltInFunctionESDTNFTBurn || eventIdentifier == core.BuiltInFunctionESDTWipe {
		tokensSupply.Add(&data.TokenInfo{
			Token:       token,
			Identifier:  converters.ComputeTokenIdentifier(token, nonceBig.Uint64()),
			Timestamp:   timestamp,
			TimestampMs: timestampMs,
			Nonce:       nonceBig.Uint64(),
		})
	}

	isNFTCreate := eventIdentifier == core.BuiltInFunctionESDTNFTCreate
	shouldReturn := !isNFTCreate || len(topics) < numTopicsWithReceiverAddress
	if shouldReturn {
		return
	}

	esdtTokenBytes := topics[3]
	esdtToken := &esdt.ESDigitalToken{}
	err := np.marshalizer.Unmarshal(esdtToken, esdtTokenBytes)
	if err != nil {
		log.Warn("nftsProcessor.processNFTEventOnSender() cannot urmarshal", "error", err.Error())
		return
	}

	tokenMetaData := converters.PrepareTokenMetaData(convertMetaData(np.pubKeyConverter, esdtToken.TokenMetaData))
	tokensCreateInfo.Add(&data.TokenInfo{
		Token:       token,
		Identifier:  converters.ComputeTokenIdentifier(token, nonceBig.Uint64()),
		Timestamp:   timestamp,
		TimestampMs: timestampMs,
		Data:        tokenMetaData,
		Nonce:       nonceBig.Uint64(),
	})
}

func convertMetaData(pubKeyConverter core.PubkeyConverter, metaData *esdt.MetaData) *alteredAccount.TokenMetaData {
	if metaData == nil {
		return nil
	}
	encodedCreatorAddr, err := pubKeyConverter.Encode(metaData.Creator)
	if err != nil {
		log.Warn("nftsProcessor.convertMetaData", "cannot encode creator address", "error", err, "address", metaData.Creator)
	}

	return &alteredAccount.TokenMetaData{
		Nonce:      metaData.Nonce,
		Name:       string(metaData.Name),
		Creator:    encodedCreatorAddr,
		Royalties:  metaData.Royalties,
		Hash:       metaData.Hash,
		URIs:       metaData.URIs,
		Attributes: metaData.Attributes,
	}
}
