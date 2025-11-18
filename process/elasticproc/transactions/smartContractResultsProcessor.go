package transactions

import (
	"encoding/hex"
	"strconv"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	indexerData "github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
)

type smartContractResultsProcessor struct {
	pubKeyConverter         core.PubkeyConverter
	hasher                  hashing.Hasher
	marshalizer             marshal.Marshalizer
	dataFieldParser         DataFieldParser
	balanceConverter        dataindexer.BalanceConverter
	relayedV1V2DisableEpoch uint32
}

func newSmartContractResultsProcessor(
	pubKeyConverter core.PubkeyConverter,
	marshalzier marshal.Marshalizer,
	hasher hashing.Hasher,
	dataFieldParser DataFieldParser,
	balanceConverter dataindexer.BalanceConverter,
	relayedV1V2DisableEpoch uint32,
) *smartContractResultsProcessor {
	return &smartContractResultsProcessor{
		pubKeyConverter:         pubKeyConverter,
		marshalizer:             marshalzier,
		hasher:                  hasher,
		dataFieldParser:         dataFieldParser,
		balanceConverter:        balanceConverter,
		relayedV1V2DisableEpoch: relayedV1V2DisableEpoch,
	}
}

func (proc *smartContractResultsProcessor) processSCRs(
	miniBlocks []*block.MiniBlock,
	headerData *indexerData.HeaderData,
	scrs map[string]*outport.SCRInfo,
) []*indexerData.ScResult {
	allSCRs := make([]*indexerData.ScResult, 0, len(scrs))

	// a copy of the SCRS map is needed because proc.processSCRsFromMiniblock would remove items from the original map
	workingSCRSMap := copySCRSMap(scrs)
	for _, mb := range miniBlocks {
		if mb.Type != block.SmartContractResultBlock {
			continue
		}

		indexerSCRs := proc.processSCRsFromMiniblock(headerData, mb, workingSCRSMap)

		allSCRs = append(allSCRs, indexerSCRs...)
	}

	for scrHashHex, noMBScrInfo := range workingSCRSMap {
		indexerScr := proc.prepareSmartContractResult(scrHashHex, nil, noMBScrInfo, headerData, headerData.ShardID, headerData.ShardID)

		allSCRs = append(allSCRs, indexerScr)
	}

	return allSCRs
}

func (proc *smartContractResultsProcessor) processSCRsFromMiniblock(
	headerData *indexerData.HeaderData,
	mb *block.MiniBlock,
	scrs map[string]*outport.SCRInfo,
) []*indexerData.ScResult {
	mbHash, err := core.CalculateHash(proc.marshalizer, proc.hasher, mb)
	if err != nil {
		log.Warn("smartContractResultsProcessor.processSCRsFromMiniblock cannot calculate miniblock hash")
		return []*indexerData.ScResult{}
	}

	indexerSCRs := make([]*indexerData.ScResult, 0, len(mb.TxHashes))
	for _, scrHash := range mb.TxHashes {
		scrHashHex := hex.EncodeToString(scrHash)
		scrInfo, ok := scrs[scrHashHex]
		if !ok {
			log.Warn("smartContractResultsProcessor.processSCRsFromMiniblock scr not found in map",
				"scr hash", scrHashHex,
			)
			continue
		}

		indexerSCR := proc.prepareSmartContractResult(hex.EncodeToString(scrHash), mbHash, scrInfo, headerData, mb.SenderShardID, mb.ReceiverShardID)
		indexerSCRs = append(indexerSCRs, indexerSCR)

		delete(scrs, scrHashHex)
	}

	return indexerSCRs
}

func (proc *smartContractResultsProcessor) prepareSmartContractResult(
	scrHashHex string,
	mbHash []byte,
	scrInfo *outport.SCRInfo,
	headerData *indexerData.HeaderData,
	senderShard uint32,
	receiverShard uint32,
) *indexerData.ScResult {
	scr := scrInfo.SmartContractResult
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

	res := proc.dataFieldParser.Parse(scr.Data, scr.SndAddr, scr.RcvAddr, headerData.NumberOfShards)

	senderAddr := proc.pubKeyConverter.SilentEncode(scr.SndAddr, log)
	receiverAddr := proc.pubKeyConverter.SilentEncode(scr.RcvAddr, log)
	receiversAddr, _ := proc.pubKeyConverter.EncodeSlice(res.Receivers)

	valueNum, err := proc.balanceConverter.ConvertBigValueToFloat(scr.Value)
	if err != nil {
		log.Warn("smartContractResultsProcessor.prepareSmartContractResult cannot compute scr value as num",
			"value", scr.Value, "hash", scrHashHex, "error", err)
	}

	esdtValuesNum, err := proc.balanceConverter.ComputeSliceOfStringsAsFloat(res.ESDTValues)
	if err != nil {
		log.Warn("smartContractResultsProcessor.prepareSmartContractResult cannot compute scr esdt values as num",
			"esdt values", res.ESDTValues, "hash", scrHashHex, "error", err)
	}

	var esdtValues []string
	if areESDTValuesOK(res.ESDTValues) {
		esdtValues = res.ESDTValues
	}

	isRelayed := res.IsRelayed && headerData.ShardID < proc.relayedV1V2DisableEpoch

	feeInfo := getFeeInfo(scrInfo)
	return &indexerData.ScResult{
		Hash:               scrHashHex,
		MBHash:             hexEncodedMBHash,
		Nonce:              scr.Nonce,
		GasLimit:           scr.GasLimit,
		GasPrice:           scr.GasPrice,
		Value:              scr.Value.String(),
		ValueNum:           valueNum,
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
		Timestamp:          headerData.Timestamp,
		SenderAddressBytes: scr.SndAddr,
		SenderShard:        senderShard,
		ReceiverShard:      receiverShard,
		Operation:          res.Operation,
		Function:           converters.TruncateFieldIfExceedsMaxLength(res.Function),
		ESDTValues:         esdtValues,
		ESDTValuesNum:      esdtValuesNum,
		Tokens:             converters.TruncateSliceElementsIfExceedsMaxLength(res.Tokens),
		Receivers:          receiversAddr,
		ReceiversShardIDs:  res.ReceiversShardID,
		IsRelayed:          isRelayed,
		OriginalSender:     originalSenderAddr,
		InitialTxFee:       feeInfo.Fee.String(),
		InitialTxGasUsed:   feeInfo.GasUsed,
		GasRefunded:        feeInfo.GasRefunded,
		ExecutionOrder:     int(scrInfo.ExecutionOrder),
		UUID:               converters.GenerateBase64UUID(),
		Epoch:              headerData.Epoch,
		TimestampMs:        headerData.TimestampMs,
	}
}

func copySCRSMap(initial map[string]*outport.SCRInfo) map[string]*outport.SCRInfo {
	newMap := make(map[string]*outport.SCRInfo)
	for key, value := range initial {
		newMap[key] = value
	}
	return newMap
}
