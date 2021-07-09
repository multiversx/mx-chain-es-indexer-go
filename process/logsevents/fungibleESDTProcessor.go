package logsevents

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const (
	numTopicsWithReceiverAddress = 3
)

type fungibleESDTProcessor struct {
	pubKeyConverter               core.PubkeyConverter
	shardCoordinator              sharding.Coordinator
	fungibleOperationsIdentifiers map[string]struct{}
}

func newFungibleESDTProcessor(pubKeyConverter core.PubkeyConverter, shardCoordinator sharding.Coordinator) *fungibleESDTProcessor {
	return &fungibleESDTProcessor{
		pubKeyConverter:  pubKeyConverter,
		shardCoordinator: shardCoordinator,
		fungibleOperationsIdentifiers: map[string]struct{}{
			core.BuiltInFunctionESDTTransfer:  {},
			core.BuiltInFunctionESDTBurn:      {},
			core.BuiltInFunctionESDTLocalMint: {},
			core.BuiltInFunctionESDTLocalBurn: {},
			core.BuiltInFunctionESDTWipe:      {},
		},
	}
}

func (fep *fungibleESDTProcessor) processLogsAndEventsESDT(logsAndEvents map[string]nodeData.LogHandler, accounts data.AlteredAccountsHandler, txsMap map[string]*data.Transaction, scrsMap map[string]*data.ScResult) {
	for logHash, log := range logsAndEvents {
		if check.IfNil(log) {
			continue
		}

		fep.processEventsESDT(logHash, log.GetLogEvents(), accounts, txsMap, scrsMap)
	}
}

func (fep *fungibleESDTProcessor) processEventsESDT(
	logHash string,
	events []nodeData.EventHandler,
	accounts data.AlteredAccountsHandler,
	txsMap map[string]*data.Transaction,
	scrsMap map[string]*data.ScResult,
) {
	logHashHexEncoded := hex.EncodeToString([]byte(logHash))
	for _, event := range events {
		if check.IfNil(event) {
			continue
		}

		tokenIdentifier := fep.processEvent(event, accounts)
		tx, ok := txsMap[logHashHexEncoded]
		if ok {
			tx.EsdtTokenIdentifier = tokenIdentifier
			continue
		}

		scr, ok := scrsMap[logHashHexEncoded]
		if ok {
			scr.EsdtTokenIdentifier = tokenIdentifier
			continue
		}
	}
}

func (fep *fungibleESDTProcessor) processEvent(event nodeData.EventHandler, accounts data.AlteredAccountsHandler) string {
	identifier := event.GetIdentifier()
	if _, ok := fep.fungibleOperationsIdentifiers[string(identifier)]; !ok {
		return ""
	}

	topics := event.GetTopics()
	address := event.GetAddress()
	if len(topics) < numTopicsWithReceiverAddress-1 {
		return ""
	}
	tokenID := string(topics[0])

	selfShardID := fep.shardCoordinator.SelfId()
	shardIDSender := fep.shardCoordinator.ComputeId(address)
	if shardIDSender == selfShardID {
		fep.processEventOnSenderShard(event, accounts)
	}

	if len(topics) < numTopicsWithReceiverAddress {
		return tokenID
	}

	receiverAddr := topics[2]
	receiverShard := fep.shardCoordinator.ComputeId(receiverAddr)
	if receiverShard != selfShardID {
		return tokenID
	}

	encodedAddr := fep.pubKeyConverter.Encode(receiverAddr)
	accounts.Add(encodedAddr, &data.AlteredAccount{
		IsESDTOperation: true,
		TokenIdentifier: tokenID,
	})

	return tokenID
}

func (fep *fungibleESDTProcessor) processEventOnSenderShard(event nodeData.EventHandler, accounts data.AlteredAccountsHandler) {
	topics := event.GetTopics()
	tokenID := topics[0]

	encodedAddr := fep.pubKeyConverter.Encode(event.GetAddress())
	accounts.Add(encodedAddr, &data.AlteredAccount{
		IsESDTOperation: true,
		TokenIdentifier: string(tokenID),
	})
}
