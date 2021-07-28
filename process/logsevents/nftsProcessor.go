package logsevents

import (
	"math/big"
	"time"

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("indexer/process/logsevents")

type nftsProcessor struct {
	pubKeyConverter          core.PubkeyConverter
	nftOperationsIdentifiers map[string]struct{}
	shardCoordinator         elasticIndexer.ShardCoordinator
	marshalizer              marshal.Marshalizer
}

func newNFTsProcessor(
	shardCoordinator elasticIndexer.ShardCoordinator,
	pubKeyConverter core.PubkeyConverter,
	marshalizer marshal.Marshalizer,
) *nftsProcessor {
	return &nftsProcessor{
		shardCoordinator: shardCoordinator,
		pubKeyConverter:  pubKeyConverter,
		marshalizer:      marshalizer,
		nftOperationsIdentifiers: map[string]struct{}{
			core.BuiltInFunctionESDTNFTTransfer:      {},
			core.BuiltInFunctionESDTNFTBurn:          {},
			core.BuiltInFunctionESDTNFTAddQuantity:   {},
			core.BuiltInFunctionESDTNFTCreate:        {},
			core.BuiltInFunctionMultiESDTNFTTransfer: {},
		},
	}
}

func (np *nftsProcessor) processEvent(args *argsProcessEvent) (string, bool) {
	eventIdentifier := string(args.event.GetIdentifier())
	_, ok := np.nftOperationsIdentifiers[eventIdentifier]
	if !ok {
		return "", false
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
		return "", false
	}

	sender := args.event.GetAddress()
	senderShardID := np.shardCoordinator.ComputeId(sender)
	if senderShardID == np.shardCoordinator.SelfId() {
		np.processNFTEventOnSender(args.event, args.accounts, args.tokens, args.timestamp, args.tagsCount)
	}

	token := string(topics[0])
	identifier := converters.ComputeTokenIdentifier(token, nonceBig.Uint64())

	if !np.shouldAddReceiverData(args) {
		return identifier, true
	}

	valueBig := big.NewInt(0).SetBytes(topics[2])
	receiver := args.event.GetTopics()[3]
	encodedReceiver := np.pubKeyConverter.Encode(topics[3])
	receiverShardID := np.shardCoordinator.ComputeId(receiver)
	if receiverShardID != np.shardCoordinator.SelfId() {
		args.pendingBalances.addInfo(encodedReceiver, token, nonceBig.Uint64(), valueBig.String())
		return identifier, true
	}

	if senderShardID != receiverShardID {
		args.pendingBalances.addInfo(encodedReceiver, token, nonceBig.Uint64(), big.NewInt(0).String())
	}

	args.accounts.Add(encodedReceiver, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: token,
		NFTNonce:        nonceBig.Uint64(),
	})

	return identifier, true
}

func (np *nftsProcessor) shouldAddReceiverData(args *argsProcessEvent) bool {
	eventIdentifier := string(args.event.GetIdentifier())
	isWrongIdentifier := eventIdentifier != core.BuiltInFunctionESDTNFTTransfer &&
		eventIdentifier != core.BuiltInFunctionMultiESDTNFTTransfer

	if isWrongIdentifier || len(args.event.GetTopics()) < numTopicsWithReceiverAddress {
		return false
	}

	return true
}

func (np *nftsProcessor) processNFTEventOnSender(
	event coreData.EventHandler,
	accounts data.AlteredAccountsHandler,
	tokensCreateInfo data.TokensHandler,
	timestamp uint64,
	tagsCount data.CountTags,
) {
	sender := event.GetAddress()
	topics := event.GetTopics()
	token := string(topics[0])
	nonceBig := big.NewInt(0).SetBytes(topics[1])
	bech32Addr := np.pubKeyConverter.Encode(sender)

	alteredAccount := &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: token,
		NFTNonce:        nonceBig.Uint64(),
	}

	accounts.Add(bech32Addr, alteredAccount)

	shouldReturn := string(event.GetIdentifier()) != core.BuiltInFunctionESDTNFTCreate || len(topics) < numTopicsWithReceiverAddress
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

	tokenMetaData := converters.PrepareTokenMetaData(np.pubKeyConverter, esdtToken)
	tokensCreateInfo.Add(&data.TokenInfo{
		Token:      token,
		Identifier: converters.ComputeTokenIdentifier(token, nonceBig.Uint64()),
		Timestamp:  time.Duration(timestamp),
		Data:       tokenMetaData,
	})

	if tokenMetaData != nil {
		tagsCount.ParseTags(tokenMetaData.Tags)
	}
}
