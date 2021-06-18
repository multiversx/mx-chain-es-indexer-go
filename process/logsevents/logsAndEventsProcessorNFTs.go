package logsevents

import (
	"math/big"

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-logger/check"
	"github.com/ElrondNetwork/elrond-go/core"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type logsAndEventsProcessor struct {
	pubKeyConverter          core.PubkeyConverter
	nftOperationsIdentifiers map[string]struct{}
	shardCoordinator         sharding.Coordinator
}

// NewLogsAndEventsProcessorNFT will create a new instance for the logsAndEventsProcessor
func NewLogsAndEventsProcessorNFT(
	shardCoordinator sharding.Coordinator,
	pubKeyConverter core.PubkeyConverter,
) (*logsAndEventsProcessor, error) {
	if check.IfNil(shardCoordinator) {
		return nil, elasticIndexer.ErrNilShardCoordinator
	}
	if check.IfNil(pubKeyConverter) {
		return nil, elasticIndexer.ErrNilPubkeyConverter
	}

	return &logsAndEventsProcessor{
		shardCoordinator: shardCoordinator,
		pubKeyConverter:  pubKeyConverter,
		nftOperationsIdentifiers: map[string]struct{}{
			core.BuiltInFunctionESDTNFTTransfer:    {},
			core.BuiltInFunctionESDTNFTBurn:        {},
			core.BuiltInFunctionESDTNFTAddQuantity: {},
			core.BuiltInFunctionESDTNFTCreate:      {},
		},
	}, nil
}

// ProcessLogsAndEvents will process provided logs and events
func (lep *logsAndEventsProcessor) ProcessLogsAndEvents(
	logsAndEvents map[string]nodeData.LogHandler,
	accounts data.AlteredAccountsHandler,
) {
	if logsAndEvents == nil || accounts == nil {
		return
	}

	for _, txLog := range logsAndEvents {
		lep.processNFTOperationLog(txLog, accounts)
	}
}

func (lep *logsAndEventsProcessor) processNFTOperationLog(txLog nodeData.LogHandler, accounts data.AlteredAccountsHandler) {
	events := txLog.GetLogEvents()
	if len(events) == 0 {
		return
	}

	for _, event := range events {
		lep.processEvent(event, accounts)
	}
}

func (lep *logsAndEventsProcessor) processEvent(event nodeData.EventHandler, accounts data.AlteredAccountsHandler) {
	_, ok := lep.nftOperationsIdentifiers[string(event.GetIdentifier())]
	if !ok {
		return
	}
	sender := event.GetAddress()

	if lep.shardCoordinator.ComputeId(sender) == lep.shardCoordinator.SelfId() {
		lep.processNFTEventOnSender(event, accounts)
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
	if lep.shardCoordinator.ComputeId(receiver) != lep.shardCoordinator.SelfId() {
		return
	}

	encodedReceiver := lep.pubKeyConverter.Encode(receiver)
	accounts.Add(encodedReceiver, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: token,
		NFTNonce:        nonceBig.Uint64(),
		IsNFTCreate:     false,
	})

	return
}

func (lep *logsAndEventsProcessor) processNFTEventOnSender(event nodeData.EventHandler, accounts data.AlteredAccountsHandler) {
	sender := event.GetAddress()
	topics := event.GetTopics()
	token := string(topics[0])
	nonceBig := big.NewInt(0).SetBytes(topics[1])
	bech32Addr := lep.pubKeyConverter.Encode(sender)

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
