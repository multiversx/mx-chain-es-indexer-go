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

	pendingBalanceIdentifier = "pending"
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

func (fep *fungibleESDTProcessor) processEvent(args *argsProcessEvent) (string, string, bool) {
	identifier := args.event.GetIdentifier()
	_, ok := fep.fungibleOperationsIdentifiers[string(identifier)]
	if !ok {
		return "", "", false
	}

	topics := args.event.GetTopics()
	nonceBig := big.NewInt(0).SetBytes(topics[1])
	if nonceBig.Uint64() > 0 {
		// this is a semi-fungible token so we should return
		return "", "", false
	}

	address := args.event.GetAddress()
	if len(topics) < numTopicsWithReceiverAddress-1 {
		return "", "", true
	}

	selfShardID := fep.shardCoordinator.SelfId()
	senderShardID := fep.shardCoordinator.ComputeId(address)
	if senderShardID == selfShardID {
		fep.processEventOnSenderShard(args.event, args.accounts)
	}

	tokenID, valueStr := fep.processEventDestination(args, senderShardID, selfShardID)
	return tokenID, valueStr, true
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

func (fep *fungibleESDTProcessor) processEventDestination(args *argsProcessEvent, senderShardID uint32, selfShardID uint32) (string, string) {
	topics := args.event.GetTopics()
	tokenID := string(topics[0])
	valueBig := big.NewInt(0).SetBytes(topics[2])

	if len(topics) < numTopicsWithReceiverAddress {
		return tokenID, valueBig.String()
	}

	receiverAddr := topics[3]
	receiverShardID := fep.shardCoordinator.ComputeId(receiverAddr)
	encodedAddr := fep.pubKeyConverter.Encode(receiverAddr)
	if receiverShardID != selfShardID {
		args.pendingBalances.addInfo(encodedAddr, tokenID, 0, valueBig.String())
		return tokenID, valueBig.String()
	}

	if senderShardID != receiverShardID {
		args.pendingBalances.addInfo(encodedAddr, tokenID, 0, big.NewInt(0).String())
	}

	args.accounts.Add(encodedAddr, &data.AlteredAccount{
		IsESDTOperation: true,
		TokenIdentifier: tokenID,
	})

	return tokenID, valueBig.String()
}
