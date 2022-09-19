package transactions

import (
	"encoding/hex"
	"math/big"
	"strconv"
	"time"

	indexerData "github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/sharding"
	"github.com/ElrondNetwork/elrond-go-core/data"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	datafield "github.com/ElrondNetwork/elrond-vm-common/parsers/dataField"
)

type smartContractResultsProcessor struct {
	pubKeyConverter core.PubkeyConverter
	hasher          hashing.Hasher
	marshalizer     marshal.Marshalizer
	dataFieldParser DataFieldParser
}

func newSmartContractResultsProcessor(
	pubKeyConverter core.PubkeyConverter,
	marshalzier marshal.Marshalizer,
	hasher hashing.Hasher,
	dataFieldParser DataFieldParser,
) *smartContractResultsProcessor {
	return &smartContractResultsProcessor{
		pubKeyConverter: pubKeyConverter,
		marshalizer:     marshalzier,
		hasher:          hasher,
		dataFieldParser: dataFieldParser,
	}
}

func (proc *smartContractResultsProcessor) processSCRs(
	body *block.Body,
	header coreData.HeaderHandler,
	txsHandler map[string]data.TransactionHandlerWithGasUsedAndFee,
) []*indexerData.ScResult {
	allSCRs := make([]*indexerData.ScResult, 0, len(txsHandler))
	for _, mb := range body.MiniBlocks {
		if mb.Type != block.SmartContractResultBlock {
			continue
		}

		indexerSCRs := proc.processSCRsFromMiniblock(header, mb, txsHandler)

		allSCRs = append(allSCRs, indexerSCRs...)
	}

	selfShardID := header.GetShardID()
	for scrHash, noMBScr := range txsHandler {
		scr, ok := noMBScr.GetTxHandler().(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}

		indexerScr := proc.prepareSmartContractResult([]byte(scrHash), nil, scr, header, selfShardID, selfShardID, noMBScr.GetFee(), noMBScr.GetGasUsed())

		allSCRs = append(allSCRs, indexerScr)
	}

	return allSCRs
}

func (proc *smartContractResultsProcessor) processSCRsFromMiniblock(
	header coreData.HeaderHandler,
	mb *block.MiniBlock,
	scrs map[string]data.TransactionHandlerWithGasUsedAndFee,
) []*indexerData.ScResult {
	mbHash, err := core.CalculateHash(proc.marshalizer, proc.hasher, mb)
	if err != nil {
		log.Warn("smartContractResultsProcessor.processSCRsFromMiniblock cannot calculate miniblock hash")
		return []*indexerData.ScResult{}
	}

	indexerSCRs := make([]*indexerData.ScResult, 0, len(mb.TxHashes))
	for _, scrHash := range mb.TxHashes {
		scrHandler, ok := scrs[string(scrHash)]
		if !ok {
			log.Warn("smartContractResultsProcessor.processSCRsFromMiniblock scr not found in map",
				"scr hash", hex.EncodeToString(scrHash),
			)
			continue
		}
		scr, ok := scrHandler.GetTxHandler().(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}

		indexerSCR := proc.prepareSmartContractResult(scrHash, mbHash, scr, header, mb.SenderShardID, mb.ReceiverShardID, scrHandler.GetFee(), scrHandler.GetGasUsed())
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
	initialTxFee *big.Int,
	initialTxGasUsed uint64,
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
	originalSenderAddr := ""
	if scr.OriginalSender != nil {
		originalSenderAddr = proc.pubKeyConverter.Encode(scr.OriginalSender)
	}

	res := proc.dataFieldParser.Parse(scr.Data, scr.SndAddr, scr.RcvAddr)

	return &indexerData.ScResult{
		Hash:               hex.EncodeToString(scrHash),
		MBHash:             hexEncodedMBHash,
		Nonce:              scr.Nonce,
		GasLimit:           scr.GasLimit,
		GasPrice:           scr.GasPrice,
		Value:              scr.Value.String(),
		Sender:             proc.pubKeyConverter.Encode(scr.SndAddr),
		Receiver:           proc.pubKeyConverter.Encode(scr.RcvAddr),
		RelayerAddr:        relayerAddr,
		RelayedValue:       relayedValue,
		Code:               string(scr.Code),
		Data:               scr.Data,
		PrevTxHash:         hex.EncodeToString(scr.PrevTxHash),
		OriginalTxHash:     hex.EncodeToString(scr.OriginalTxHash),
		CallType:           strconv.Itoa(int(scr.CallType)),
		CodeMetadata:       scr.CodeMetadata,
		ReturnMessage:      string(scr.ReturnMessage),
		Timestamp:          time.Duration(header.GetTimeStamp()),
		SenderAddressBytes: scr.SndAddr,
		SenderShard:        senderShard,
		ReceiverShard:      receiverShard,
		Operation:          res.Operation,
		Function:           res.Function,
		ESDTValues:         res.ESDTValues,
		Tokens:             res.Tokens,
		Receivers:          datafield.EncodeBytesSlice(proc.pubKeyConverter.Encode, res.Receivers),
		ReceiversShardIDs:  res.ReceiversShardID,
		IsRelayed:          res.IsRelayed,
		OriginalSender:     originalSenderAddr,
		InitialTxFee:       initialTxFee.String(),
		InitialTxGasUsed:   initialTxGasUsed,
	}
}

func (proc *smartContractResultsProcessor) addScrsReceiverToAlteredAccounts(
	alteredAccounts indexerData.AlteredAccountsHandler,
	scrs []*indexerData.ScResult,
	selfShardID uint32,
	numOfShards uint32,
) {
	for _, scr := range scrs {
		receiverAddr, _ := proc.pubKeyConverter.Decode(scr.Receiver)
		if selfShardID != sharding.ComputeShardID(receiverAddr, numOfShards) {
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
