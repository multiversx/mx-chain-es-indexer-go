package transactions

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
)

const emptyString = ""

type dbTransactionBuilder struct {
	addressPubkeyConverter core.PubkeyConverter
	shardCoordinator       indexer.ShardCoordinator
	txFeeCalculator        indexer.FeesProcessorHandler
}

func newTransactionDBBuilder(
	addressPubkeyConverter core.PubkeyConverter,
	shardCoordinator indexer.ShardCoordinator,
	txFeeCalculator indexer.FeesProcessorHandler,
) *dbTransactionBuilder {
	return &dbTransactionBuilder{
		addressPubkeyConverter: addressPubkeyConverter,
		shardCoordinator:       shardCoordinator,
		txFeeCalculator:        txFeeCalculator,
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
	}
}

func (dtb *dbTransactionBuilder) prepareSmartContractResult(
	scHash string,
	sc *smartContractResult.SmartContractResult,
	header coreData.HeaderHandler,
) *data.ScResult {
	relayerAddr := ""
	if len(sc.RelayerAddr) > 0 {
		relayerAddr = dtb.addressPubkeyConverter.Encode(sc.RelayerAddr)
	}

	relayedValue := ""
	if sc.RelayedValue != nil {
		relayedValue = sc.RelayedValue.String()
	}

	return &data.ScResult{
		Hash:           hex.EncodeToString([]byte(scHash)),
		Nonce:          sc.Nonce,
		GasLimit:       sc.GasLimit,
		GasPrice:       sc.GasPrice,
		Value:          sc.Value.String(),
		Sender:         dtb.addressPubkeyConverter.Encode(sc.SndAddr),
		Receiver:       dtb.addressPubkeyConverter.Encode(sc.RcvAddr),
		RelayerAddr:    relayerAddr,
		RelayedValue:   relayedValue,
		Code:           string(sc.Code),
		Data:           sc.Data,
		PrevTxHash:     hex.EncodeToString(sc.PrevTxHash),
		OriginalTxHash: hex.EncodeToString(sc.OriginalTxHash),
		CallType:       strconv.Itoa(int(sc.CallType)),
		CodeMetadata:   sc.CodeMetadata,
		ReturnMessage:  string(sc.ReturnMessage),
		Timestamp:      time.Duration(header.GetTimeStamp()),
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

func (dtb *dbTransactionBuilder) addScrsReceiverToAlteredAccounts(
	alteredAccounts data.AlteredAccountsHandler,
	scrs []*data.ScResult,
) {
	for _, scr := range scrs {
		receiverAddr, _ := dtb.addressPubkeyConverter.Decode(scr.Receiver)
		shardID := dtb.shardCoordinator.ComputeId(receiverAddr)
		if shardID != dtb.shardCoordinator.SelfId() {
			continue
		}

		egldBalanceNotChanged := scr.Value == emptyString || scr.Value == "0"
		if egldBalanceNotChanged {
			// the smart contract results that don't alter the balance of the receiver address should be ignored
			continue
		}

		alteredAccounts.Add(scr.Receiver, &data.AlteredAccount{
			IsSender: false,
		})
	}
}

func (dtb *dbTransactionBuilder) isInSameShard(sender string) bool {
	senderBytes, err := dtb.addressPubkeyConverter.Decode(sender)
	if err != nil {
		return false
	}

	return dtb.shardCoordinator.ComputeId(senderBytes) == dtb.shardCoordinator.SelfId()
}
