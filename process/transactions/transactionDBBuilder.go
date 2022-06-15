package transactions

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
)

const emptyString = ""

type dbTransactionBuilder struct {
	addressPubkeyConverter core.PubkeyConverter
	shardCoordinator       indexer.ShardCoordinator
	txFeeCalculator        indexer.FeesProcessorHandler
	dataFieldParser        DataFieldParser
}

func newTransactionDBBuilder(
	addressPubkeyConverter core.PubkeyConverter,
	shardCoordinator indexer.ShardCoordinator,
	txFeeCalculator indexer.FeesProcessorHandler,
	dataFieldParser DataFieldParser,
) *dbTransactionBuilder {
	return &dbTransactionBuilder{
		addressPubkeyConverter: addressPubkeyConverter,
		shardCoordinator:       shardCoordinator,
		txFeeCalculator:        txFeeCalculator,
		dataFieldParser:        dataFieldParser,
	}
}

func (dtb *dbTransactionBuilder) prepareTransaction(
	tx *transaction.Transaction,
	txHash []byte,
	mbHash []byte,
	mb *block.MiniBlock,
	header coreData.HeaderHandler,
	txStatus string,
) *data.Transaction {
	gasUsed := dtb.txFeeCalculator.ComputeGasLimit(tx)
	fee := dtb.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, gasUsed)

	isScCall := core.IsSmartContractAddress(tx.RcvAddr)
	res := dtb.dataFieldParser.Parse(tx.Data, tx.SndAddr, tx.RcvAddr)

	return &data.Transaction{
		Hash:                 hex.EncodeToString(txHash),
		MBHash:               hex.EncodeToString(mbHash),
		Nonce:                tx.Nonce,
		Round:                header.GetRound(),
		Value:                tx.Value.String(),
		Receiver:             dtb.addressPubkeyConverter.Encode(tx.RcvAddr),
		Sender:               dtb.addressPubkeyConverter.Encode(tx.SndAddr),
		ReceiverShard:        mb.ReceiverShardID,
		SenderShard:          mb.SenderShardID,
		GasPrice:             tx.GasPrice,
		GasLimit:             tx.GasLimit,
		Data:                 tx.Data,
		Signature:            hex.EncodeToString(tx.Signature),
		Timestamp:            time.Duration(header.GetTimeStamp()),
		Status:               txStatus,
		GasUsed:              gasUsed,
		Fee:                  fee.String(),
		ReceiverUserName:     tx.RcvUserName,
		SenderUserName:       tx.SndUserName,
		ReceiverAddressBytes: tx.RcvAddr,
		IsScCall:             isScCall,
		Operation:            res.Operation,
		Function:             res.Function,
		ESDTValues:           res.ESDTValues,
		Tokens:               res.Tokens,
		Receivers:            encodeBytesSlice(dtb.addressPubkeyConverter.Encode, res.Receivers),
		ReceiversShardIDs:    res.ReceiversShardID,
		IsRelayed:            res.IsRelayed,
		Version:              tx.Version,
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
	return &data.Transaction{
		Hash:          hex.EncodeToString(txHash),
		MBHash:        hex.EncodeToString(mbHash),
		Nonce:         0,
		Round:         rTx.Round,
		Value:         rTx.Value.String(),
		Receiver:      dtb.addressPubkeyConverter.Encode(rTx.RcvAddr),
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
	return &data.Receipt{
		Hash:      hex.EncodeToString([]byte(recHash)),
		Value:     rec.Value.String(),
		Sender:    dtb.addressPubkeyConverter.Encode(rec.SndAddr),
		Data:      string(rec.Data),
		TxHash:    hex.EncodeToString(rec.TxHash),
		Timestamp: time.Duration(header.GetTimeStamp()),
	}
}

func (dtb *dbTransactionBuilder) isInSameShard(sender string) bool {
	senderBytes, err := dtb.addressPubkeyConverter.Decode(sender)
	if err != nil {
		return false
	}

	return dtb.shardCoordinator.ComputeId(senderBytes) == dtb.shardCoordinator.SelfId()
}
