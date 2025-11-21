package transactions

import (
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/sharding"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
)

type dbTransactionBuilder struct {
	addressPubkeyConverter  core.PubkeyConverter
	dataFieldParser         DataFieldParser
	balanceConverter        dataindexer.BalanceConverter
	relayedV1V2DisableEpoch uint32
}

func newTransactionDBBuilder(
	addressPubkeyConverter core.PubkeyConverter,
	dataFieldParser DataFieldParser,
	balanceConverter dataindexer.BalanceConverter,
	relayedV1V2DisableEpoch uint32,
) *dbTransactionBuilder {
	return &dbTransactionBuilder{
		addressPubkeyConverter:  addressPubkeyConverter,
		dataFieldParser:         dataFieldParser,
		balanceConverter:        balanceConverter,
		relayedV1V2DisableEpoch: relayedV1V2DisableEpoch,
	}
}

func (dtb *dbTransactionBuilder) prepareTransaction(
	txInfo *outport.TxInfo,
	txHash []byte,
	mbHash []byte,
	mb *block.MiniBlock,
	headerData *data.HeaderData,
	txStatus string,
) *data.Transaction {
	tx := txInfo.Transaction

	isScCall := core.IsSmartContractAddress(tx.RcvAddr)
	res := dtb.dataFieldParser.Parse(tx.Data, tx.SndAddr, tx.RcvAddr, headerData.NumberOfShards)

	receiverAddr := dtb.addressPubkeyConverter.SilentEncode(tx.RcvAddr, log)
	senderAddr := dtb.addressPubkeyConverter.SilentEncode(tx.SndAddr, log)
	receiversAddr, _ := dtb.addressPubkeyConverter.EncodeSlice(res.Receivers)

	receiverShardID := mb.ReceiverShardID
	if mb.Type == block.InvalidBlock {
		receiverShardID = sharding.ComputeShardID(tx.RcvAddr, headerData.NumberOfShards)
	}

	valueNum, err := dtb.balanceConverter.ConvertBigValueToFloat(tx.Value)
	if err != nil {
		log.Warn("dbTransactionBuilder.prepareTransaction: cannot compute value as num", "value", tx.Value,
			"hash", txHash, "error", err)
	}

	feeInfo := getFeeInfo(txInfo)
	feeNum, err := dtb.balanceConverter.ConvertBigValueToFloat(feeInfo.Fee)
	if err != nil {
		log.Warn("dbTransactionBuilder.prepareTransaction: cannot compute transaction fee as num", "fee", feeInfo.Fee,
			"hash", txHash, "error", err)
	}
	esdtValuesNum, err := dtb.balanceConverter.ComputeSliceOfStringsAsFloat(res.ESDTValues)
	if err != nil {
		log.Warn("dbTransactionBuilder.prepareTransaction: cannot compute esdt values as num",
			"esdt values", res.ESDTValues, "hash", txHash, "error", err)
	}

	var esdtValues []string
	if areESDTValuesOK(res.ESDTValues) {
		esdtValues = res.ESDTValues
	}
	guardianAddress := ""
	if len(tx.GuardianAddr) > 0 {
		guardianAddress = dtb.addressPubkeyConverter.SilentEncode(tx.GuardianAddr, log)
	}
	relayedAddress := ""
	if len(tx.RelayerAddr) > 0 {
		relayedAddress = dtb.addressPubkeyConverter.SilentEncode(tx.RelayerAddr, log)
	}

	senderUserName := converters.TruncateFieldIfExceedsMaxLengthBase64(string(tx.SndUserName))
	receiverUserName := converters.TruncateFieldIfExceedsMaxLengthBase64(string(tx.RcvUserName))

	eTx := &data.Transaction{
		Hash:              hex.EncodeToString(txHash),
		MBHash:            hex.EncodeToString(mbHash),
		Nonce:             tx.Nonce,
		Round:             headerData.Round,
		Value:             tx.Value.String(),
		Receiver:          receiverAddr,
		Sender:            senderAddr,
		ValueNum:          valueNum,
		ReceiverShard:     receiverShardID,
		SenderShard:       mb.SenderShardID,
		GasPrice:          tx.GasPrice,
		GasLimit:          tx.GasLimit,
		Data:              tx.Data,
		Signature:         hex.EncodeToString(tx.Signature),
		Timestamp:         headerData.Timestamp,
		Status:            txStatus,
		GasUsed:           feeInfo.GasUsed,
		InitialPaidFee:    feeInfo.InitialPaidFee.String(),
		Fee:               feeInfo.Fee.String(),
		FeeNum:            feeNum,
		ReceiverUserName:  []byte(receiverUserName),
		SenderUserName:    []byte(senderUserName),
		IsScCall:          isScCall,
		ESDTValues:        esdtValues,
		ESDTValuesNum:     esdtValuesNum,
		Receivers:         receiversAddr,
		Version:           tx.Version,
		GuardianAddress:   guardianAddress,
		GuardianSignature: hex.EncodeToString(tx.GuardianSignature),
		ExecutionOrder:    int(txInfo.ExecutionOrder),
		Operation:         res.Operation,
		RelayedSignature:  hex.EncodeToString(tx.RelayerSignature),
		RelayedAddr:       relayedAddress,
		HadRefund:         feeInfo.HadRefund,
		UUID:              converters.GenerateBase64UUID(),
		Epoch:             headerData.Epoch,
		TimestampMs:       headerData.TimestampMs,
	}

	hasValidRelayer := len(eTx.RelayedAddr) == len(eTx.Sender) && len(eTx.RelayedAddr) > 0
	hasValidRelayerSignature := len(eTx.RelayedSignature) == len(eTx.Signature) && len(eTx.RelayedSignature) > 0
	isRelayedV3 := hasValidRelayer && hasValidRelayerSignature

	eTx.Function = converters.TruncateFieldIfExceedsMaxLength(res.Function)
	eTx.Tokens = converters.TruncateSliceElementsIfExceedsMaxLength(res.Tokens)
	eTx.ReceiversShardIDs = res.ReceiversShardID

	relayedV1V2Enabled := headerData.Epoch < dtb.relayedV1V2DisableEpoch
	eTx.IsRelayed = res.IsRelayed || isRelayedV3

	if res.IsRelayed && !relayedV1V2Enabled {
		// will be treated as move balance, so reset some fields
		eTx.IsRelayed = false
		eTx.Function = ""
		eTx.RelayedAddr = ""
		eTx.RelayedSignature = ""
		eTx.Receivers = []string{}
		eTx.ReceiversShardIDs = []uint32{}
	}

	return eTx
}

func (dtb *dbTransactionBuilder) prepareRewardTransaction(
	rTxInfo *outport.RewardInfo,
	txHash []byte,
	mbHash []byte,
	mb *block.MiniBlock,
	headerData *data.HeaderData,
	txStatus string,
) *data.Transaction {
	rTx := rTxInfo.Reward
	valueNum, err := dtb.balanceConverter.ConvertBigValueToFloat(rTx.Value)
	if err != nil {
		log.Warn("dbTransactionBuilder.prepareRewardTransaction cannot compute value as num", "value", rTx.Value,
			"hash", txHash, "error", err)
	}

	receiverAddr := dtb.addressPubkeyConverter.SilentEncode(rTx.RcvAddr, log)

	return &data.Transaction{
		Hash:           hex.EncodeToString(txHash),
		MBHash:         hex.EncodeToString(mbHash),
		Nonce:          0,
		Round:          rTx.Round,
		Value:          rTx.Value.String(),
		ValueNum:       valueNum,
		Receiver:       receiverAddr,
		Sender:         fmt.Sprintf("%d", core.MetachainShardId),
		ReceiverShard:  mb.ReceiverShardID,
		SenderShard:    mb.SenderShardID,
		GasPrice:       0,
		GasLimit:       0,
		Data:           make([]byte, 0),
		Signature:      "",
		Timestamp:      headerData.Timestamp,
		Status:         txStatus,
		Operation:      rewardsOperation,
		ExecutionOrder: int(rTxInfo.ExecutionOrder),
		UUID:           converters.GenerateBase64UUID(),
		Epoch:          headerData.Epoch,
		TimestampMs:    headerData.TimestampMs,
	}
}

func (dtb *dbTransactionBuilder) prepareReceipt(
	recHashHex string,
	rec *receipt.Receipt,
	headerData *data.HeaderData,
) *data.Receipt {
	senderAddr := dtb.addressPubkeyConverter.SilentEncode(rec.SndAddr, log)

	return &data.Receipt{
		Hash:        recHashHex,
		Value:       rec.Value.String(),
		Sender:      senderAddr,
		Data:        string(rec.Data),
		TxHash:      hex.EncodeToString(rec.TxHash),
		Timestamp:   headerData.Timestamp,
		TimestampMs: headerData.TimestampMs,
	}
}
