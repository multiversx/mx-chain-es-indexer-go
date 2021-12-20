package transactions

import (
	"encoding/hex"
	"strconv"
	"time"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	indexerData "github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

type smartContractResultsProcessor struct {
	pubKeyConverter  core.PubkeyConverter
	shardCoordinator indexer.ShardCoordinator
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	dataFieldParser  DataFieldParser
}

func newSmartContractResultsProcessor(
	pubKeyConverter core.PubkeyConverter,
	shardCoordinator indexer.ShardCoordinator,
	marshalzier marshal.Marshalizer,
	hasher hashing.Hasher,
	dataFieldParser DataFieldParser,
) *smartContractResultsProcessor {
	return &smartContractResultsProcessor{
		pubKeyConverter:  pubKeyConverter,
		shardCoordinator: shardCoordinator,
		marshalizer:      marshalzier,
		hasher:           hasher,
		dataFieldParser:  dataFieldParser,
	}
}

func (proc *smartContractResultsProcessor) processSCRs(
	body *block.Body,
	header coreData.HeaderHandler,
	txsHandler map[string]data.TransactionHandler,
) []*indexerData.ScResult {
	allSCRs := make([]*indexerData.ScResult, 0, len(txsHandler))
	scrs := convertHandlerMap(txsHandler)
	for _, mb := range body.MiniBlocks {
		if mb.Type != block.SmartContractResultBlock {
			continue
		}

		indexerSCRs := proc.processSCRsFromMiniblock(header, mb, scrs)

		allSCRs = append(allSCRs, indexerSCRs...)
	}

	selfShardID := proc.shardCoordinator.SelfId()
	for scrHash, noMBScr := range scrs {
		indexerScr := proc.prepareSmartContractResult([]byte(scrHash), nil, noMBScr, header, selfShardID, selfShardID)

		allSCRs = append(allSCRs, indexerScr)
	}

	return allSCRs
}

func convertHandlerMap(txsHandler map[string]data.TransactionHandler) map[string]*smartContractResult.SmartContractResult {
	scrs := make(map[string]*smartContractResult.SmartContractResult, len(txsHandler))
	for txHandlerHash, txHandler := range txsHandler {
		scr, ok := txHandler.(*smartContractResult.SmartContractResult)
		if !ok {
			log.Warn("smartContractResultsProcessor.processSCRsFromMiniblock cannot convert TransactionHandler to scr",
				"scr hash", hex.EncodeToString([]byte(txHandlerHash)),
			)
			continue
		}

		scrs[txHandlerHash] = scr
	}

	return scrs
}

func (proc *smartContractResultsProcessor) processSCRsFromMiniblock(
	header coreData.HeaderHandler,
	mb *block.MiniBlock,
	scrs map[string]*smartContractResult.SmartContractResult,
) []*indexerData.ScResult {
	mbHash, err := core.CalculateHash(proc.marshalizer, proc.hasher, mb)
	if err != nil {
		log.Warn("smartContractResultsProcessor.processSCRsFromMiniblock cannot calculate miniblock hash")
		return []*indexerData.ScResult{}
	}

	indexerSCRs := make([]*indexerData.ScResult, 0, len(mb.TxHashes))
	for _, scrHash := range mb.TxHashes {
		scr, ok := scrs[string(scrHash)]
		if !ok {
			log.Warn("smartContractResultsProcessor.processSCRsFromMiniblock scr not found in map",
				"scr hash", hex.EncodeToString(scrHash),
			)
			continue
		}

		indexerSCR := proc.prepareSmartContractResult(scrHash, mbHash, scr, header, mb.SenderShardID, mb.ReceiverShardID)
		indexerSCRs = append(indexerSCRs, indexerSCR)

		delete(scrs, string(scrHash))
	}

	return indexerSCRs
}

func (proc *smartContractResultsProcessor) prepareSmartContractResult(
	scrHash []byte,
	mbHash []byte,
	scr *smartContractResult.SmartContractResult,
	header coreData.HeaderHandler,
	senderShard uint32,
	receiverShard uint32,
) *indexerData.ScResult {
	hexEncodedMBHash := ""
	if len(mbHash) > 0 {
		hexEncodedMBHash = hex.EncodeToString(mbHash)
	}

	relayerAddr := ""
	if len(scr.RelayerAddr) > 0 {
		relayerAddr = proc.pubKeyConverter.Encode(scr.RelayerAddr)
	}

	relayedValue := ""
	if scr.RelayedValue != nil {
		relayedValue = scr.RelayedValue.String()
	}

	res := proc.dataFieldParser.Parse(scr.Data, scr.SndAddr, scr.RcvAddr)

	return &indexerData.ScResult{
		Hash:              hex.EncodeToString(scrHash),
		MBHash:            hexEncodedMBHash,
		Nonce:             scr.Nonce,
		GasLimit:          scr.GasLimit,
		GasPrice:          scr.GasPrice,
		Value:             scr.Value.String(),
		Sender:            proc.pubKeyConverter.Encode(scr.SndAddr),
		Receiver:          proc.pubKeyConverter.Encode(scr.RcvAddr),
		RelayerAddr:       relayerAddr,
		RelayedValue:      relayedValue,
		Code:              string(scr.Code),
		Data:              scr.Data,
		PrevTxHash:        hex.EncodeToString(scr.PrevTxHash),
		OriginalTxHash:    hex.EncodeToString(scr.OriginalTxHash),
		CallType:          strconv.Itoa(int(scr.CallType)),
		CodeMetadata:      scr.CodeMetadata,
		ReturnMessage:     string(scr.ReturnMessage),
		Timestamp:         time.Duration(header.GetTimeStamp()),
		SenderShard:       senderShard,
		ReceiverShard:     receiverShard,
		Operation:         res.Operation,
		Function:          res.Function,
		ESDTValues:        res.ESDTValues,
		Tokens:            res.Tokens,
		Receivers:         res.Receivers,
		ReceiversShardIDs: res.ReceiversShardID,
	}
}

func (proc *smartContractResultsProcessor) addScrsReceiverToAlteredAccounts(
	alteredAccounts indexerData.AlteredAccountsHandler,
	scrs []*indexerData.ScResult,
) {
	for _, scr := range scrs {
		receiverAddr, _ := proc.pubKeyConverter.Decode(scr.Receiver)
		shardID := proc.shardCoordinator.ComputeId(receiverAddr)
		if shardID != proc.shardCoordinator.SelfId() {
			continue
		}

		balanceNotChanged := scr.Value == emptyString || scr.Value == "0"
		if balanceNotChanged {
			// the smart contract results that don't alter the balance of the receiver address should be ignored
			continue
		}

		alteredAccounts.Add(scr.Receiver, &indexerData.AlteredAccount{
			IsSender:      false,
			BalanceChange: true,
		})
	}
}
