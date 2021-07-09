package logsevents

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process/tags"
	"github.com/ElrondNetwork/elrond-go/core"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-vm-common/data/esdt"
)

type nftsProcessor struct {
	pubKeyConverter          core.PubkeyConverter
	nftOperationsIdentifiers map[string]struct{}
	shardCoordinator         sharding.Coordinator
	marshalizer              marshal.Marshalizer
}

func newNFTsProcessor(
	shardCoordinator sharding.Coordinator,
	pubKeyConverter core.PubkeyConverter,
	marshalizer marshal.Marshalizer,
) *nftsProcessor {
	return &nftsProcessor{
		shardCoordinator: shardCoordinator,
		pubKeyConverter:  pubKeyConverter,
		marshalizer:      marshalizer,
		nftOperationsIdentifiers: map[string]struct{}{
			core.BuiltInFunctionESDTNFTTransfer:    {},
			core.BuiltInFunctionESDTNFTBurn:        {},
			core.BuiltInFunctionESDTNFTAddQuantity: {},
			core.BuiltInFunctionESDTNFTCreate:      {},
		},
	}
}

func (np *nftsProcessor) processEvent(args *argsProcessEvent) (string, bool) {
	eventIdentifier := string(args.event.GetIdentifier())
	_, ok := np.nftOperationsIdentifiers[eventIdentifier]
	if !ok {
		return "", false
	}

	sender := args.event.GetAddress()
	if np.shardCoordinator.ComputeId(sender) == np.shardCoordinator.SelfId() {
		np.processNFTEventOnSender(args.event, args.accounts, args.tokens, args.timestamp, args.tagsCount)
	}

	// topics contains:
	// [0] -- token identifier
	// [1] -- nonce of the NFT (bytes)
	// [2] -- receiver NFT address -- in case of NFTTransfer OR ESDT token data in case of NFTCreate
	topics := args.event.GetTopics()
	token := string(topics[0])
	nonceBig := big.NewInt(0).SetBytes(topics[1])
	identifier := converters.ComputeTokenIdentifier(token, nonceBig.Uint64())
	shouldReturn := eventIdentifier != core.BuiltInFunctionESDTNFTTransfer || len(topics) < 3
	if shouldReturn {
		return identifier, true
	}

	receiver := topics[2]
	if np.shardCoordinator.ComputeId(receiver) != np.shardCoordinator.SelfId() {
		return identifier, true
	}

	encodedReceiver := np.pubKeyConverter.Encode(receiver)
	args.accounts.Add(encodedReceiver, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: token,
		NFTNonce:        nonceBig.Uint64(),
	})

	return identifier, true
}

func (np *nftsProcessor) processNFTEventOnSender(
	event nodeData.EventHandler,
	accounts data.AlteredAccountsHandler,
	tokensCreateInfo data.TokensHandler,
	timestamp uint64,
	tagsCount tags.CountTags,
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

	shouldReturn := string(event.GetIdentifier()) != core.BuiltInFunctionESDTNFTCreate || len(topics) < 3
	if shouldReturn {
		return
	}

	esdtTokenBytes := topics[2]
	esdtToken := &esdt.ESDigitalToken{}
	err := np.marshalizer.Unmarshal(esdtToken, esdtTokenBytes)
	if err != nil {
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
