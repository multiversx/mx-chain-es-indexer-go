package logsevents

import (
	"math/big"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go/core"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type nftsProcessor struct {
	pubKeyConverter          core.PubkeyConverter
	nftOperationsIdentifiers map[string]struct{}
	shardCoordinator         sharding.Coordinator
}

func newNFTsProcessor(
	shardCoordinator sharding.Coordinator,
	pubKeyConverter core.PubkeyConverter,
) *nftsProcessor {
	return &nftsProcessor{
		shardCoordinator: shardCoordinator,
		pubKeyConverter:  pubKeyConverter,
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
) {
	if logsAndEvents == nil || accounts == nil {
		return
	}

	for _, txLog := range logsAndEvents {
		np.processNFTOperationLog(txLog, accounts)
	}
}

func (np *nftsProcessor) processNFTOperationLog(txLog nodeData.LogHandler, accounts data.AlteredAccountsHandler) {
	events := txLog.GetLogEvents()
	if len(events) == 0 {
		return
	}

	for _, event := range events {
		np.processEvent(event, accounts)
	}
}

func (np *nftsProcessor) processEvent(event nodeData.EventHandler, accounts data.AlteredAccountsHandler) {
	_, ok := np.nftOperationsIdentifiers[string(event.GetIdentifier())]
	if !ok {
		return
	}
	sender := event.GetAddress()

	if np.shardCoordinator.ComputeId(sender) == np.shardCoordinator.SelfId() {
		np.processNFTEventOnSender(event, accounts)
	}

	// topics contains:
	// [0] -- token identifier
	// [1] -- nonce of the NFT (bytes)
	// [2] -- receiver NFT address -- in case of NFTTransfer OR token type in case of NFTCreate
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
		IsNFTCreate:     false,
	})

	return
}

func (np *nftsProcessor) processNFTEventOnSender(event nodeData.EventHandler, accounts data.AlteredAccountsHandler) {
	sender := event.GetAddress()
	topics := event.GetTopics()
	token := string(topics[0])
	nonceBig := big.NewInt(0).SetBytes(topics[1])
	bech32Addr := np.pubKeyConverter.Encode(sender)

	if string(event.GetIdentifier()) != core.BuiltInFunctionESDTNFTCreate {
		accounts.Add(bech32Addr, &data.AlteredAccount{
			IsNFTOperation:  true,
			TokenIdentifier: token,
			NFTNonce:        nonceBig.Uint64(),
			IsNFTCreate:     false,
		})

		return
	}

	if len(topics) < 3 {
		return
	}

	accounts.Add(bech32Addr, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: token,
		NFTNonce:        nonceBig.Uint64(),
		IsNFTCreate:     true,
		Type:            string(topics[2]),
	})
}
