package transactions

import (
	"encoding/hex"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

const (
	rewardsOperation = "reward"
)

type txsGrouper struct {
	txBuilder   *dbTransactionBuilder
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
}

func newTxsGrouper(
	txBuilder *dbTransactionBuilder,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) *txsGrouper {
	return &txsGrouper{
		txBuilder:   txBuilder,
		hasher:      hasher,
		marshalizer: marshalizer,
	}
}

func (tg *txsGrouper) groupNormalTxs(
	mbIndex int,
	mb *block.MiniBlock,
	headerData *data.HeaderData,
	txs map[string]*outport.TxInfo,
	isImportDB bool,
	numOfShards uint32,
) (map[string]*data.Transaction, error) {
	transactions := make(map[string]*data.Transaction)

	mbHash, err := core.CalculateHash(tg.marshalizer, tg.hasher, mb)
	if err != nil {
		return nil, err
	}

	executedTxHashes := extractExecutedTxHashes(mbIndex, mb.TxHashes, headerData)
	mbStatus := computeStatus(headerData.ShardID, mb.ReceiverShardID)
	for _, txHash := range executedTxHashes {
		dbTx, ok := tg.prepareNormalTxForDB(mbHash, mb, mbStatus, txHash, txs, headerData, numOfShards)
		if !ok {
			continue
		}

		if tg.shouldIndex(mb.ReceiverShardID, isImportDB, headerData.ShardID) {
			transactions[string(txHash)] = dbTx
		}
	}

	return transactions, nil
}

func extractExecutedTxHashes(mbIndex int, mbTxHashes [][]byte, headerData *data.HeaderData) [][]byte {
	miniblockHeaders := headerData.MiniBlockHeaders
	if len(miniblockHeaders) <= mbIndex {
		return mbTxHashes
	}

	firstProcessed := miniblockHeaders[mbIndex].GetIndexOfFirstTxProcessed()
	lastProcessed := miniblockHeaders[mbIndex].GetIndexOfLastTxProcessed()

	executedTxHashes := make([][]byte, 0)
	for txIndex, txHash := range mbTxHashes {
		if int32(txIndex) < firstProcessed || int32(txIndex) > lastProcessed {
			continue
		}

		executedTxHashes = append(executedTxHashes, txHash)
	}

	return executedTxHashes
}

func (tg *txsGrouper) prepareNormalTxForDB(
	mbHash []byte,
	mb *block.MiniBlock,
	mbStatus string,
	txHash []byte,
	txs map[string]*outport.TxInfo,
	headerData *data.HeaderData,
	numOfShards uint32,
) (*data.Transaction, bool) {
	txInfo, okGet := txs[hex.EncodeToString(txHash)]
	if !okGet {
		return nil, false
	}

	dbTx := tg.txBuilder.prepareTransaction(txInfo, txHash, mbHash, mb, headerData, mbStatus, numOfShards)

	return dbTx, true
}

func (tg *txsGrouper) groupRewardsTxs(
	mbIndex int,
	mb *block.MiniBlock,
	headerData *data.HeaderData,
	txs map[string]*outport.RewardInfo,
	isImportDB bool,
) (map[string]*data.Transaction, error) {
	rewardsTxs := make(map[string]*data.Transaction)
	mbHash, err := core.CalculateHash(tg.marshalizer, tg.hasher, mb)
	if err != nil {
		return nil, err
	}

	mbStatus := computeStatus(headerData.ShardID, mb.ReceiverShardID)
	executedTxHashes := extractExecutedTxHashes(mbIndex, mb.TxHashes, headerData)
	for _, txHash := range executedTxHashes {
		rewardDBTx, ok := tg.prepareRewardTxForDB(mbHash, mb, mbStatus, txHash, txs, headerData)
		if !ok {
			continue
		}

		if tg.shouldIndex(mb.ReceiverShardID, isImportDB, headerData.ShardID) {
			rewardsTxs[string(txHash)] = rewardDBTx
		}
	}

	return rewardsTxs, nil
}

func (tg *txsGrouper) prepareRewardTxForDB(
	mbHash []byte,
	mb *block.MiniBlock,
	mbStatus string,
	txHash []byte,
	txs map[string]*outport.RewardInfo,
	headerData *data.HeaderData,
) (*data.Transaction, bool) {
	rtx, okGet := txs[hex.EncodeToString(txHash)]
	if !okGet {
		return nil, false
	}

	dbTx := tg.txBuilder.prepareRewardTransaction(rtx, txHash, mbHash, mb, headerData, mbStatus)

	return dbTx, true
}

func (tg *txsGrouper) groupInvalidTxs(
	mbIndex int,
	mb *block.MiniBlock,
	headerData *data.HeaderData,
	txs map[string]*outport.TxInfo,
	numOfShards uint32,
) (map[string]*data.Transaction, error) {
	transactions := make(map[string]*data.Transaction)
	mbHash, err := core.CalculateHash(tg.marshalizer, tg.hasher, mb)
	if err != nil {
		return nil, err
	}

	executedTxHashes := extractExecutedTxHashes(mbIndex, mb.TxHashes, headerData)
	for _, txHash := range executedTxHashes {
		invalidDBTx, ok := tg.prepareInvalidTxForDB(mbHash, mb, txHash, txs, headerData, numOfShards)
		if !ok {
			continue
		}

		transactions[string(txHash)] = invalidDBTx
	}

	return transactions, nil
}

func (tg *txsGrouper) prepareInvalidTxForDB(
	mbHash []byte,
	mb *block.MiniBlock,
	txHash []byte,
	txs map[string]*outport.TxInfo,
	headerDta *data.HeaderData,
	numOfShards uint32,
) (*data.Transaction, bool) {
	txInfo, okGet := txs[hex.EncodeToString(txHash)]
	if !okGet {
		return nil, false
	}

	dbTx := tg.txBuilder.prepareTransaction(txInfo, txHash, mbHash, mb, headerDta, transaction.TxStatusInvalid.String(), numOfShards)

	return dbTx, true
}

func (tg *txsGrouper) shouldIndex(destinationShardID uint32, isImportDB bool, selfShardID uint32) bool {
	if !isImportDB {
		return true
	}

	return selfShardID == destinationShardID
}

func (tg *txsGrouper) groupReceipts(headerData *data.HeaderData, txsPool map[string]*receipt.Receipt) []*data.Receipt {
	dbReceipts := make([]*data.Receipt, 0)
	for hashHex, rec := range txsPool {
		dbReceipts = append(dbReceipts, tg.txBuilder.prepareReceipt(hashHex, rec, headerData))
	}

	return dbReceipts
}

func computeStatus(selfShardID uint32, receiverShardID uint32) string {
	if selfShardID == receiverShardID {
		return transaction.TxStatusSuccess.String()
	}

	return transaction.TxStatusPending.String()
}

func convertMapTxsToSlice(txs map[string]*data.Transaction) []*data.Transaction {
	transactions := make([]*data.Transaction, len(txs))
	i := 0
	for _, tx := range txs {
		transactions[i] = tx
		i++
	}
	return transactions
}
