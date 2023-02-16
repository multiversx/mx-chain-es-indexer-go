package transactions

import (
	"encoding/hex"
	"math/big"
	"strconv"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	indexerData "github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"
)

type smartContractResultsProcessor struct {
	pubKeyConverter  core.PubkeyConverter
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	dataFieldParser  DataFieldParser
	balanceConverter dataindexer.BalanceConverter
}

func newSmartContractResultsProcessor(
	pubKeyConverter core.PubkeyConverter,
	marshalzier marshal.Marshalizer,
	hasher hashing.Hasher,
	dataFieldParser DataFieldParser,
	balanceConverter dataindexer.BalanceConverter,
) *smartContractResultsProcessor {
	return &smartContractResultsProcessor{
		pubKeyConverter:  pubKeyConverter,
		marshalizer:      marshalzier,
		hasher:           hasher,
		dataFieldParser:  dataFieldParser,
		balanceConverter: balanceConverter,
	}
}

func (proc *smartContractResultsProcessor) processSCRs(
	body *block.Body,
	header coreData.HeaderHandler,
	txsHandler map[string]data.TransactionHandlerWithGasUsedAndFee,
	numOfShards uint32,
) []*indexerData.ScResult {
	allSCRs := make([]*indexerData.ScResult, 0, len(txsHandler))

	// a copy of the SCRS map is needed because proc.processSCRsFromMiniblock would remove items from the original map
	workingSCRSMap := copySCRSMap(txsHandler)
	for _, mb := range body.MiniBlocks {
		if mb.Type != block.SmartContractResultBlock {
			continue
		}

		indexerSCRs := proc.processSCRsFromMiniblock(header, mb, workingSCRSMap, numOfShards)

		allSCRs = append(allSCRs, indexerSCRs...)
	}

	selfShardID := header.GetShardID()
	for scrHash, noMBScr := range workingSCRSMap {
		scr, ok := noMBScr.GetTxHandler().(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}

		indexerScr := proc.prepareSmartContractResult([]byte(scrHash), nil, scr, header, selfShardID, selfShardID, noMBScr.GetFee(), noMBScr.GetGasUsed(), numOfShards)

		allSCRs = append(allSCRs, indexerScr)
	}

	return allSCRs
}

func (proc *smartContractResultsProcessor) processSCRsFromMiniblock(
	header coreData.HeaderHandler,
	mb *block.MiniBlock,
	scrs map[string]data.TransactionHandlerWithGasUsedAndFee,
	numOfShards uint32,
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

		indexerSCR := proc.prepareSmartContractResult(scrHash, mbHash, scr, header, mb.SenderShardID, mb.ReceiverShardID, scrHandler.GetFee(), scrHandler.GetGasUsed(), numOfShards)
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
	numOfShards uint32,
) *indexerData.ScResult {
	hexEncodedMBHash := ""
	if len(mbHash) > 0 {
		hexEncodedMBHash = hex.EncodeToString(mbHash)
	}

	relayerAddr := ""
	if len(scr.RelayerAddr) > 0 {
		relayerAddr = proc.pubKeyConverter.SilentEncode(scr.RelayerAddr, log)
	}

	relayedValue := ""
	if scr.RelayedValue != nil {
		relayedValue = scr.RelayedValue.String()
	}
	originalSenderAddr := ""
	if scr.OriginalSender != nil {
		originalSenderAddr = proc.pubKeyConverter.SilentEncode(scr.OriginalSender, log)
	}

	res := proc.dataFieldParser.Parse(scr.Data, scr.SndAddr, scr.RcvAddr, numOfShards)

	senderAddr := proc.pubKeyConverter.SilentEncode(scr.SndAddr, log)
	receiverAddr := proc.pubKeyConverter.SilentEncode(scr.RcvAddr, log)
	receiversAddr, _ := proc.pubKeyConverter.EncodeSlice(res.Receivers)

	return &indexerData.ScResult{
		Hash:               hex.EncodeToString(scrHash),
		MBHash:             hexEncodedMBHash,
		Nonce:              scr.Nonce,
		GasLimit:           scr.GasLimit,
		GasPrice:           scr.GasPrice,
		Value:              scr.Value.String(),
		Sender:             senderAddr,
		Receiver:           receiverAddr,
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
		ESDTValuesNum:      proc.balanceConverter.ComputeSliceOfStringsAsFloat(res.ESDTValues),
		Tokens:             res.Tokens,
		Receivers:          receiversAddr,
		ReceiversShardIDs:  res.ReceiversShardID,
		IsRelayed:          res.IsRelayed,
		OriginalSender:     originalSenderAddr,
		InitialTxFee:       initialTxFee.String(),
		InitialTxGasUsed:   initialTxGasUsed,
	}
}

func copySCRSMap(initial map[string]data.TransactionHandlerWithGasUsedAndFee) map[string]data.TransactionHandlerWithGasUsedAndFee {
	newMap := make(map[string]data.TransactionHandlerWithGasUsedAndFee)
	for key, value := range initial {
		newMap[key] = value
	}
	return newMap
}
