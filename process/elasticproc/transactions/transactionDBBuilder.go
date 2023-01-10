package transactions

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/multiversx/mx-chain-core-go/core"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"
)

type dbTransactionBuilder struct {
	addressPubkeyConverter core.PubkeyConverter
	dataFieldParser        DataFieldParser
}

func newTransactionDBBuilder(
	addressPubkeyConverter core.PubkeyConverter,
	dataFieldParser DataFieldParser,
) *dbTransactionBuilder {
	return &dbTransactionBuilder{
		addressPubkeyConverter: addressPubkeyConverter,
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
	fee *big.Int,
	gasUsed uint64,
	initialPaidFee *big.Int,
	numOfShards uint32,
) *data.Transaction {
	isScCall := core.IsSmartContractAddress(tx.RcvAddr)
	res := dtb.dataFieldParser.Parse(tx.Data, tx.SndAddr, tx.RcvAddr, numOfShards)

	return &data.Transaction{
		Hash:              hex.EncodeToString(txHash),
		MBHash:            hex.EncodeToString(mbHash),
		Nonce:             tx.Nonce,
		Round:             header.GetRound(),
		Value:             tx.Value.String(),
		Receiver:          dtb.addressPubkeyConverter.Encode(tx.RcvAddr),
		Sender:            dtb.addressPubkeyConverter.Encode(tx.SndAddr),
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
		ReceiverUserName:  tx.RcvUserName,
		SenderUserName:    tx.SndUserName,
		IsScCall:          isScCall,
		Operation:         res.Operation,
		Function:          res.Function,
		ESDTValues:        res.ESDTValues,
		Tokens:            res.Tokens,
		Receivers:         datafield.EncodeBytesSlice(dtb.addressPubkeyConverter.Encode, res.Receivers),
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
