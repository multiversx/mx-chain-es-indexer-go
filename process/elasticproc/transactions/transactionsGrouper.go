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
	header coreData.HeaderHandler,
	txs map[string]coreData.TransactionHandlerWithGasUsedAndFee,
	alteredAccounts data.AlteredAccountsHandler,
	isImportDB bool,
) (map[string]*data.Transaction, error) {
	transactions := make(map[string]*data.Transaction)

	mbHash, err := core.CalculateHash(tg.marshalizer, tg.hasher, mb)
	if err != nil {
		return nil, err
	}

	selfShardID := header.GetShardID()
	executedTxHashes := extractExecutedTxHashes(mbIndex, mb.TxHashes, header)
	mbStatus := computeStatus(selfShardID, mb.ReceiverShardID)
	for _, txHash := range executedTxHashes {
		dbTx, ok := tg.prepareNormalTxForDB(mbHash, mb, mbStatus, txHash, txs, header)
		if !ok {
			continue
		}

		tg.addToAlteredAddresses(dbTx, alteredAccounts, mb, selfShardID, false)
		if tg.shouldIndex(mb.ReceiverShardID, isImportDB, selfShardID) {
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
	txs map[string]coreData.TransactionHandlerWithGasUsedAndFee,
	header coreData.HeaderHandler,
) (*data.Transaction, bool) {
	txHandler, okGet := txs[string(txHash)]
	if !okGet {
		return nil, false
	}

	tx, okCast := txHandler.GetTxHandler().(*transaction.Transaction)
	if !okCast {
		return nil, false
	}

	dbTx := tg.txBuilder.prepareTransaction(tx, txHash, mbHash, mb, header, mbStatus, txHandler.GetFee(), txHandler.GetGasUsed(), txHandler.GetInitialPaidFee())

	return dbTx, true
}

func (tg *txsGrouper) groupRewardsTxs(
	mbIndex int,
	mb *block.MiniBlock,
	header coreData.HeaderHandler,
	txs map[string]coreData.TransactionHandlerWithGasUsedAndFee,
	alteredAccounts data.AlteredAccountsHandler,
	isImportDB bool,
) (map[string]*data.Transaction, error) {
	rewardsTxs := make(map[string]*data.Transaction)
	mbHash, err := core.CalculateHash(tg.marshalizer, tg.hasher, mb)
	if err != nil {
		return nil, err
	}

	selfShardID := header.GetShardID()
	mbStatus := computeStatus(selfShardID, mb.ReceiverShardID)
	executedTxHashes := extractExecutedTxHashes(mbIndex, mb.TxHashes, header)
	for _, txHash := range executedTxHashes {
		rewardDBTx, ok := tg.prepareRewardTxForDB(mbHash, mb, mbStatus, txHash, txs, header)
		if !ok {
			continue
		}

		tg.addToAlteredAddresses(rewardDBTx, alteredAccounts, mb, selfShardID, true)
		if tg.shouldIndex(mb.ReceiverShardID, isImportDB, selfShardID) {
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
	txs map[string]coreData.TransactionHandlerWithGasUsedAndFee,
	header coreData.HeaderHandler,
) (*data.Transaction, bool) {
	txHandler, okGet := txs[string(txHash)]
	if !okGet {
		return nil, false
	}

	rtx, okCast := txHandler.GetTxHandler().(*rewardTx.RewardTx)
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
	txs map[string]coreData.TransactionHandlerWithGasUsedAndFee,
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

		tg.addToAlteredAddresses(invalidDBTx, alteredAccounts, mb, header.GetShardID(), false)
		transactions[string(txHash)] = invalidDBTx
	}

	return transactions, nil
}

func (tg *txsGrouper) prepareInvalidTxForDB(
	mbHash []byte,
	mb *block.MiniBlock,
	txHash []byte,
	txs map[string]coreData.TransactionHandlerWithGasUsedAndFee,
	header coreData.HeaderHandler,
) (*data.Transaction, bool) {
	txHandler, okGet := txs[string(txHash)]
	if !okGet {
		return nil, false
	}

	tx, okCast := txHandler.GetTxHandler().(*transaction.Transaction)
	if !okCast {
		return nil, false
	}

	dbTx := tg.txBuilder.prepareTransaction(tx, txHash, mbHash, mb, header, transaction.TxStatusInvalid.String(), txHandler.GetFee(), txHandler.GetGasUsed(), txHandler.GetInitialPaidFee())

	return dbTx, true
}

func (tg *txsGrouper) shouldIndex(destinationShardID uint32, isImportDB bool, selfShardID uint32) bool {
	if !isImportDB {
		return true
	}

	return selfShardID == destinationShardID
}

func (tg *txsGrouper) groupReceipts(header coreData.HeaderHandler, txsPool map[string]coreData.TransactionHandlerWithGasUsedAndFee) []*data.Receipt {
	dbReceipts := make([]*data.Receipt, 0)
	for hash, tx := range txsPool {
		rec, ok := tx.GetTxHandler().(*receipt.Receipt)
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
