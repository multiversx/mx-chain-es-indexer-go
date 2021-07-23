package logsevents

import (
	"github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
)

const (
	numTopicsWithReceiverAddress = 3
)

type fungibleESDTProcessor struct {
	pubKeyConverter               core.PubkeyConverter
	shardCoordinator              indexer.ShardCoordinator
	fungibleOperationsIdentifiers map[string]struct{}
}

func newFungibleESDTProcessor(pubKeyConverter core.PubkeyConverter, shardCoordinator indexer.ShardCoordinator) *fungibleESDTProcessor {
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

func (fep *fungibleESDTProcessor) processEvent(args *argsProcessEvent) (string, bool) {
	identifier := args.event.GetIdentifier()
	_, ok := fep.fungibleOperationsIdentifiers[string(identifier)]
	if !ok {
		return "", false
	}

	topics := args.event.GetTopics()
	address := args.event.GetAddress()
	if len(topics) < numTopicsWithReceiverAddress-1 {
		return "", true
	}

	selfShardID := fep.shardCoordinator.SelfId()
	senderShardID := fep.shardCoordinator.ComputeId(address)
	if senderShardID == selfShardID {
		fep.processEventOnSenderShard(args.event, args.accounts)
	}

	return fep.processEventDestination(args.event, args.accounts, selfShardID), true
}

func (fep *fungibleESDTProcessor) processEventOnSenderShard(event coreData.EventHandler, accounts data.AlteredAccountsHandler) {
	topics := event.GetTopics()
	tokenID := topics[0]

	encodedAddr := fep.pubKeyConverter.Encode(event.GetAddress())
	accounts.Add(encodedAddr, &data.AlteredAccount{
		IsESDTOperation: true,
		TokenIdentifier: string(tokenID),
	})
}

func (fep *fungibleESDTProcessor) processEventDestination(event coreData.EventHandler, accounts data.AlteredAccountsHandler, selfShardID uint32) string {
	topics := event.GetTopics()
	tokenID := string(topics[0])
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
