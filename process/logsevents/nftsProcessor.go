package logsevents

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-logger/check"
	"github.com/ElrondNetwork/elrond-go/core"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
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

func (np *nftsProcessor) processLogAndEventsNFTs(
	logsAndEvents map[string]nodeData.LogHandler,
	accounts data.AlteredAccountsHandler,
	timestamp uint64,
) data.TokensHandler {
	if logsAndEvents == nil || accounts == nil {
		return nil
	}

	tokens := data.NewTokensInfo()
	for _, txLog := range logsAndEvents {
		if check.IfNil(txLog) {
			continue
		}

		np.processNFTOperationLog(txLog, accounts, tokens, timestamp)
	}

	return tokens
}

func (np *nftsProcessor) processNFTOperationLog(txLog nodeData.LogHandler, accounts data.AlteredAccountsHandler, tokens data.TokensHandler, timestamp uint64) {
	events := txLog.GetLogEvents()
	if len(events) == 0 {
		return
	}

	for _, event := range events {
		np.processEvent(event, accounts, tokens, timestamp)
	}
}

func (np *nftsProcessor) processEvent(event nodeData.EventHandler, accounts data.AlteredAccountsHandler, tokens data.TokensHandler, timestamp uint64) {
	_, ok := np.nftOperationsIdentifiers[string(event.GetIdentifier())]
	if !ok {
		return
	}
	sender := event.GetAddress()

	if np.shardCoordinator.ComputeId(sender) == np.shardCoordinator.SelfId() {
		np.processNFTEventOnSender(event, accounts, tokens, timestamp)
	}

	// topics contains:
	// [0] -- token identifier
	// [1] -- nonce of the NFT (bytes)
	// [2] -- receiver NFT address -- in case of NFTTransfer OR ESDT token data in case of NFTCreate
	topics := event.GetTopics()
	if string(event.GetIdentifier()) != core.BuiltInFunctionESDTNFTTransfer || len(topics) < 3 {
		return
	}

	token := string(topics[0])
	nonceBig := big.NewInt(0).SetBytes(topics[1])
	receiver := topics[2]
	if np.shardCoordinator.ComputeId(receiver) != np.shardCoordinator.SelfId() {
		return
	}

	encodedReceiver := np.pubKeyConverter.Encode(receiver)
	accounts.Add(encodedReceiver, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: token,
		NFTNonce:        nonceBig.Uint64(),
	})

	return
}

func (np *nftsProcessor) processNFTEventOnSender(event nodeData.EventHandler, accounts data.AlteredAccountsHandler, tokensCreateInfo data.TokensHandler, timestamp uint64) {
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

	if string(event.GetIdentifier()) != core.BuiltInFunctionESDTNFTCreate || len(topics) < 3 {
		return
	}

	esdtTokenBytes := topics[2]
	esdtToken := &esdt.ESDigitalToken{}
	err := np.marshalizer.Unmarshal(esdtToken, esdtTokenBytes)
	if err != nil {
		return
	}

	tokensCreateInfo.Add(&data.TokenInfo{
		Token:      token,
		Identifier: converters.ComputeTokenIdentifier(token, nonceBig.Uint64()),
		Timestamp:  time.Duration(timestamp),
		MetaData:   converters.PrepareTokenMetaData(np.pubKeyConverter, esdtToken),
	})
}
