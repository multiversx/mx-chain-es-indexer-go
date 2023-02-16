package transactions

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
)

type dbTransactionBuilder struct {
	addressPubkeyConverter core.PubkeyConverter
	dataFieldParser        DataFieldParser
	balanceConverter       dataindexer.BalanceConverter
}

func newTransactionDBBuilder(
	addressPubkeyConverter core.PubkeyConverter,
	dataFieldParser DataFieldParser,
	balanceConverter dataindexer.BalanceConverter,
) *dbTransactionBuilder {
	return &dbTransactionBuilder{
		addressPubkeyConverter: addressPubkeyConverter,
		dataFieldParser:        dataFieldParser,
		balanceConverter:       balanceConverter,
	}
}

func (dtb *dbTransactionBuilder) prepareTransaction(
	tx *transaction.Transaction,
	txHash []byte,
	mbHash []byte,
	mb *block.MiniBlock,
	header coreData.HeaderHandler,
	txStatus string,
	fee *big.Int,
	gasUsed uint64,
	initialPaidFee *big.Int,
	numOfShards uint32,
) *data.Transaction {
	isScCall := core.IsSmartContractAddress(tx.RcvAddr)
	res := dtb.dataFieldParser.Parse(tx.Data, tx.SndAddr, tx.RcvAddr, numOfShards)

	receiverAddr := dtb.addressPubkeyConverter.SilentEncode(tx.RcvAddr, log)
	senderAddr := dtb.addressPubkeyConverter.SilentEncode(tx.SndAddr, log)
	receiversAddr, _ := dtb.addressPubkeyConverter.EncodeSlice(res.Receivers)

	return &data.Transaction{
		Hash:              hex.EncodeToString(txHash),
		MBHash:            hex.EncodeToString(mbHash),
		Nonce:             tx.Nonce,
		Round:             header.GetRound(),
		Value:             tx.Value.String(),
		Receiver:          receiverAddr,
		Sender:            senderAddr,
		ReceiverShard:     mb.ReceiverShardID,
		SenderShard:       mb.SenderShardID,
		GasPrice:          tx.GasPrice,
		GasLimit:          tx.GasLimit,
		Data:              tx.Data,
		Signature:         hex.EncodeToString(tx.Signature),
		Timestamp:         time.Duration(header.GetTimeStamp()),
		Status:            txStatus,
		GasUsed:           gasUsed,
		InitialPaidFee:    initialPaidFee.String(),
		Fee:               fee.String(),
		FeeNum:            dtb.balanceConverter.ComputeESDTBalanceAsFloat(fee),
		ReceiverUserName:  tx.RcvUserName,
		SenderUserName:    tx.SndUserName,
		IsScCall:          isScCall,
		Operation:         res.Operation,
		Function:          res.Function,
		ESDTValues:        res.ESDTValues,
		ESDTValuesNum:     dtb.balanceConverter.ComputeSliceOfStringsAsFloat(res.ESDTValues),
		Tokens:            res.Tokens,
		Receivers:         receiversAddr,
		ReceiversShardIDs: res.ReceiversShardID,
		IsRelayed:         res.IsRelayed,
		Version:           tx.Version,
	}
}

func (dtb *dbTransactionBuilder) prepareRewardTransaction(
	rTx *rewardTx.RewardTx,
	txHash []byte,
	mbHash []byte,
	mb *block.MiniBlock,
	header coreData.HeaderHandler,
	txStatus string,
) *data.Transaction {
	receiverAddr := dtb.addressPubkeyConverter.SilentEncode(rTx.RcvAddr, log)

	return &data.Transaction{
		Hash:          hex.EncodeToString(txHash),
		MBHash:        hex.EncodeToString(mbHash),
		Nonce:         0,
		Round:         rTx.Round,
		Value:         rTx.Value.String(),
		Receiver:      receiverAddr,
		Sender:        fmt.Sprintf("%d", core.MetachainShardId),
		ReceiverShard: mb.ReceiverShardID,
		SenderShard:   mb.SenderShardID,
		GasPrice:      0,
		GasLimit:      0,
		Data:          make([]byte, 0),
		Signature:     "",
		Timestamp:     time.Duration(header.GetTimeStamp()),
		Status:        txStatus,
		Operation:     rewardsOperation,
	}
}

func (dtb *dbTransactionBuilder) prepareReceipt(
	recHash string,
	rec *receipt.Receipt,
	header coreData.HeaderHandler,
) *data.Receipt {
	senderAddr := dtb.addressPubkeyConverter.SilentEncode(rec.SndAddr, log)

	return &data.Receipt{
		Hash:      hex.EncodeToString([]byte(recHash)),
		Value:     rec.Value.String(),
		Sender:    senderAddr,
		Data:      string(rec.Data),
		TxHash:    hex.EncodeToString(rec.TxHash),
		Timestamp: time.Duration(header.GetTimeStamp()),
	}
}
