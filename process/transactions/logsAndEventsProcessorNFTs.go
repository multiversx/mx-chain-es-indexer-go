package transactions

import (
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/disabled"
	"github.com/ElrondNetwork/elrond-go/core"
	dataElrond "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type logsAndEventsProcessor struct {
	pubKeyConverter          core.PubkeyConverter
	marshalizer              marshal.Marshalizer
	logsAndEventsHandler     process.TransactionLogProcessorDatabase
	nftOperationsIdentifiers map[string]struct{}
	shardCoordinator         sharding.Coordinator
}

func newLogsAndEventsProcessorNFT(
	shardCoordinator sharding.Coordinator,
	pubKeyConverter core.PubkeyConverter,
	marshalizer marshal.Marshalizer,
) *logsAndEventsProcessor {
	return &logsAndEventsProcessor{
		shardCoordinator:     shardCoordinator,
		pubKeyConverter:      pubKeyConverter,
		marshalizer:          marshalizer,
		logsAndEventsHandler: disabled.NewNilTxLogsProcessor(),
		nftOperationsIdentifiers: map[string]struct{}{
			core.BuiltInFunctionESDTNFTTransfer:    {},
			core.BuiltInFunctionESDTNFTBurn:        {},
			core.BuiltInFunctionESDTNFTAddQuantity: {},
			core.BuiltInFunctionESDTNFTCreate:      {},
		},
	}
}

func (lep *logsAndEventsProcessor) setLogsAndEventsHandler(logsAndEventsHandler process.TransactionLogProcessorDatabase) {
	lep.logsAndEventsHandler = logsAndEventsHandler
}

func (lep *logsAndEventsProcessor) processLogsTransactions(txs []*data.Transaction, accounts data.AlteredAccountsHandler) {
	for _, tx := range txs {
		decodedHash, err := hex.DecodeString(tx.Hash)
		if err != nil {
			continue
		}

		lep.processNFTOperationLog(tx, decodedHash, accounts)
	}
}

func (lep *logsAndEventsProcessor) processLogsScrs(scrs []*data.ScResult, accounts data.AlteredAccountsHandler) {
	for _, scr := range scrs {
		decodedHash, err := hex.DecodeString(scr.Hash)
		if err != nil {
			continue
		}

		lep.processNFTOperationLog(scr, decodedHash, accounts)
	}
}

func (lep *logsAndEventsProcessor) processNFTOperationLog(op data.Operation, txHash []byte, accounts data.AlteredAccountsHandler) {
	txLog, ok := lep.logsAndEventsHandler.GetLogFromCache(txHash)
	if txLog == nil || !ok {
		return
	}
	events := txLog.GetLogEvents()
	if len(events) == 0 {
		return
	}

	for _, event := range events {
		tokenID := lep.processEvent(event, accounts)
		if tokenID != "" {
			op.SetToken(tokenID)
		}
	}
}

func (lep *logsAndEventsProcessor) processEvent(event dataElrond.EventHandler, accounts data.AlteredAccountsHandler) (token string) {
	_, ok := lep.nftOperationsIdentifiers[string(event.GetIdentifier())]
	if !ok {
		return
	}

	topics := event.GetTopics()
	if len(topics) < 2 {
		return
	}

	token = string(topics[0])
	nonceBig := big.NewInt(0).SetBytes(topics[1])

	sender := event.GetAddress()
	if lep.shardCoordinator.ComputeId(sender) == lep.shardCoordinator.SelfId() {
		bech32Addr := lep.pubKeyConverter.Encode(sender)
		accounts.Add(bech32Addr, &data.AlteredAccount{
			IsNFTOperation:  true,
			TokenIdentifier: token,
			NFTNonce:        nonceBig.Uint64(),
			IsCreate:        true,
		})
	}

	if len(topics) < 3 {
		return
	}

	receiver := topics[2]
	if lep.shardCoordinator.ComputeId(receiver) != lep.shardCoordinator.SelfId() {
		return
	}

	receiverBech32 := lep.pubKeyConverter.Encode(receiver)
	accounts.Add(receiverBech32, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: token,
		NFTNonce:        nonceBig.Uint64(),
		IsCreate:        true,
	})

	return
}
