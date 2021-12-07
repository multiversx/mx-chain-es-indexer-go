package logsevents

import (
	"math/big"

	"github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
)

const (
	numTopicsWithReceiverAddress = 4
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
			core.BuiltInFunctionESDTTransfer:         {},
			core.BuiltInFunctionESDTBurn:             {},
			core.BuiltInFunctionESDTLocalMint:        {},
			core.BuiltInFunctionESDTLocalBurn:        {},
			core.BuiltInFunctionESDTWipe:             {},
			core.BuiltInFunctionMultiESDTNFTTransfer: {},
		},
	}
}

func (fep *fungibleESDTProcessor) processEvent(args *argsProcessEvent) argOutputProcessEvent {
	identifier := args.event.GetIdentifier()
	_, ok := fep.fungibleOperationsIdentifiers[string(identifier)]
	if !ok {
		return argOutputProcessEvent{}
	}

	topics := args.event.GetTopics()
	nonceBig := big.NewInt(0).SetBytes(topics[1])
	if nonceBig.Uint64() > 0 {
		// this is a semi-fungible token so we should return
		return argOutputProcessEvent{}
	}

	address := args.event.GetAddress()
	if len(topics) < numTopicsWithReceiverAddress-1 {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	selfShardID := fep.shardCoordinator.SelfId()
	senderShardID := fep.shardCoordinator.ComputeId(address)
	if senderShardID == selfShardID {
		fep.processEventOnSenderShard(args.event, args.accounts)
	}

	tokenID, valueStr, receiver, receiverShardID := fep.processEventDestination(args, selfShardID)
	return argOutputProcessEvent{
		identifier:      tokenID,
		value:           valueStr,
		processed:       true,
		receiver:        receiver,
		receiverShardID: receiverShardID,
	}
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

func (fep *fungibleESDTProcessor) processEventDestination(args *argsProcessEvent, selfShardID uint32) (string, string, string, uint32) {
	topics := args.event.GetTopics()
	tokenID := string(topics[0])
	valueBig := big.NewInt(0).SetBytes(topics[2])

	if len(topics) < numTopicsWithReceiverAddress {
		return tokenID, valueBig.String(), "", 0
	}

	receiverAddr := topics[3]
	receiverShardID := fep.shardCoordinator.ComputeId(receiverAddr)
	encodedReceiver := fep.pubKeyConverter.Encode(receiverAddr)
	if receiverShardID != selfShardID {
		return tokenID, valueBig.String(), encodedReceiver, receiverShardID
	}

	args.accounts.Add(encodedReceiver, &data.AlteredAccount{
		IsESDTOperation: true,
		TokenIdentifier: tokenID,
	})

	return tokenID, valueBig.String(), encodedReceiver, receiverShardID
}
