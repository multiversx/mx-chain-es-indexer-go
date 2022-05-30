package transactions

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

const (
	rewardsOperation = "reward"
)

type txsGrouper struct {
	isInImportMode bool
	selfShardID    uint32
	txBuilder      *dbTransactionBuilder
	hasher         hashing.Hasher
	marshalizer    marshal.Marshalizer
}

func newTxsGrouper(
	txBuilder *dbTransactionBuilder,
	isInImportMode bool,
	selfShardID uint32,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) *txsGrouper {
	return &txsGrouper{
		txBuilder:      txBuilder,
		selfShardID:    selfShardID,
		isInImportMode: isInImportMode,
		hasher:         hasher,
		marshalizer:    marshalizer,
	}
}

func (tg *txsGrouper) groupNormalTxs(
	mbIndex int,
	mb *block.MiniBlock,
	header coreData.HeaderHandler,
	txs map[string]coreData.TransactionHandler,
	alteredAccounts data.AlteredAccountsHandler,
) (map[string]*data.Transaction, error) {
	transactions := make(map[string]*data.Transaction)

	mbHash, err := core.CalculateHash(tg.marshalizer, tg.hasher, mb)
	if err != nil {
		return nil, err
	}

	executedTxHashes := extractExecutedTxHashes(mbIndex, mb.TxHashes, header)
	mbStatus := computeStatus(tg.selfShardID, mb.ReceiverShardID)
	for _, txHash := range executedTxHashes {
		dbTx, ok := tg.prepareNormalTxForDB(mbHash, mb, mbStatus, txHash, txs, header)
		if !ok {
			continue
		}

		tg.addToAlteredAddresses(dbTx, alteredAccounts, mb, tg.selfShardID, false)
		if tg.shouldIndex(mb.ReceiverShardID) {
			transactions[string(txHash)] = dbTx
		}
	}

	return transactions, nil
}

func extractExecutedTxHashes(mbIndex int, mbTxHashes [][]byte, header coreData.HeaderHandler) [][]byte {
	miniblockHeaders := header.GetMiniBlockHeaderHandlers()
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
	txs map[string]coreData.TransactionHandler,
	header coreData.HeaderHandler,
) (*data.Transaction, bool) {
	txHandler, okGet := txs[string(txHash)]
	if !okGet {
		return nil, false
	}

	tx, okCast := txHandler.(*transaction.Transaction)
	if !okCast {
		return nil, false
	}

	dbTx := tg.txBuilder.prepareTransaction(tx, txHash, mbHash, mb, header, mbStatus)

	return dbTx, true
}

func (tg *txsGrouper) groupRewardsTxs(
	mbIndex int,
	mb *block.MiniBlock,
	header coreData.HeaderHandler,
	txs map[string]coreData.TransactionHandler,
	alteredAccounts data.AlteredAccountsHandler,
) (map[string]*data.Transaction, error) {
	rewardsTxs := make(map[string]*data.Transaction)
	mbHash, err := core.CalculateHash(tg.marshalizer, tg.hasher, mb)
	if err != nil {
		return nil, err
	}

	mbStatus := computeStatus(tg.selfShardID, mb.ReceiverShardID)
	executedTxHashes := extractExecutedTxHashes(mbIndex, mb.TxHashes, header)
	for _, txHash := range executedTxHashes {
		rewardDBTx, ok := tg.prepareRewardTxForDB(mbHash, mb, mbStatus, txHash, txs, header)
		if !ok {
			continue
		}

		tg.addToAlteredAddresses(rewardDBTx, alteredAccounts, mb, tg.selfShardID, true)
		if tg.shouldIndex(mb.ReceiverShardID) {
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
	txs map[string]coreData.TransactionHandler,
	header coreData.HeaderHandler,
) (*data.Transaction, bool) {
	txHandler, okGet := txs[string(txHash)]
	if !okGet {
		return nil, false
	}

	rtx, okCast := txHandler.(*rewardTx.RewardTx)
	if !okCast {
		return nil, false
	}

	dbTx := tg.txBuilder.prepareRewardTransaction(rtx, txHash, mbHash, mb, header, mbStatus)

	return dbTx, true
}

func (tg *txsGrouper) groupInvalidTxs(
	mbIndex int,
	mb *block.MiniBlock,
	header coreData.HeaderHandler,
	txs map[string]coreData.TransactionHandler,
	alteredAccounts data.AlteredAccountsHandler,
) (map[string]*data.Transaction, error) {
	transactions := make(map[string]*data.Transaction)
	mbHash, err := core.CalculateHash(tg.marshalizer, tg.hasher, mb)
	if err != nil {
		return nil, err
	}

	executedTxHashes := extractExecutedTxHashes(mbIndex, mb.TxHashes, header)
	for _, txHash := range executedTxHashes {
		invalidDBTx, ok := tg.prepareInvalidTxForDB(mbHash, mb, txHash, txs, header)
		if !ok {
			continue
		}

		tg.addToAlteredAddresses(invalidDBTx, alteredAccounts, mb, tg.selfShardID, false)
		transactions[string(txHash)] = invalidDBTx
	}

	return transactions, nil
}

func (tg *txsGrouper) prepareInvalidTxForDB(
	mbHash []byte,
	mb *block.MiniBlock,
	txHash []byte,
	txs map[string]coreData.TransactionHandler,
	header coreData.HeaderHandler,
) (*data.Transaction, bool) {
	txHandler, okGet := txs[string(txHash)]
	if !okGet {
		return nil, false
	}

	tx, okCast := txHandler.(*transaction.Transaction)
	if !okCast {
		return nil, false
	}

	dbTx := tg.txBuilder.prepareTransaction(tx, txHash, mbHash, mb, header, transaction.TxStatusInvalid.String())

	dbTx.GasUsed = dbTx.GasLimit
	fee := tg.txBuilder.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, dbTx.GasUsed)
	dbTx.Fee = fee.String()

	return dbTx, true
}

func (tg *txsGrouper) shouldIndex(destinationShardID uint32) bool {
	if !tg.isInImportMode {
		return true
	}

	return tg.selfShardID == destinationShardID
}

func (tg *txsGrouper) groupReceipts(header coreData.HeaderHandler, txsPool map[string]coreData.TransactionHandler) []*data.Receipt {
	dbReceipts := make([]*data.Receipt, 0)
	for hash, tx := range txsPool {
		rec, ok := tx.(*receipt.Receipt)
		if !ok {
			continue
		}

		dbReceipts = append(dbReceipts, tg.txBuilder.prepareReceipt(hash, rec, header))
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

func (tg *txsGrouper) addToAlteredAddresses(
	tx *data.Transaction,
	alteredAccounts data.AlteredAccountsHandler,
	miniBlock *block.MiniBlock,
	selfShardID uint32,
	isRewardTx bool,
) {
	if selfShardID == miniBlock.SenderShardID && !isRewardTx {
		alteredAccounts.Add(tx.Sender, &data.AlteredAccount{
			IsSender:      true,
			BalanceChange: true,
		})
	}

	ignoreTransactionReceiver := tx.Status == transaction.TxStatusInvalid.String() || tx.Sender == tx.Receiver
	if ignoreTransactionReceiver {
		return
	}

	if selfShardID == miniBlock.ReceiverShardID || miniBlock.ReceiverShardID == core.AllShardId {
		alteredAccounts.Add(tx.Receiver, &data.AlteredAccount{
			IsSender:      false,
			BalanceChange: true,
		})
	}
}
