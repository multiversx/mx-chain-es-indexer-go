package transactions

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go/core"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
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
	mb *block.MiniBlock,
	header nodeData.HeaderHandler,
	txs map[string]nodeData.TransactionHandler,
	alteredAddresses map[string]*data.AlteredAccount,
) (map[string]*data.Transaction, error) {
	transactions := make(map[string]*data.Transaction)

	mbHash, err := core.CalculateHash(tg.marshalizer, tg.hasher, mb)
	if err != nil {
		return nil, err
	}

	mbStatus := computeStatus(tg.selfShardID, mb.ReceiverShardID)
	for _, txHash := range mb.TxHashes {
		dbTx, ok := tg.prepareNormalTxForDB(mbHash, mb, mbStatus, txHash, txs, header)
		if !ok {
			continue
		}

		tg.addToAlteredAddresses(dbTx, alteredAddresses, mb, tg.selfShardID, false)
		if tg.shouldIndex(mb.ReceiverShardID) {
			transactions[string(txHash)] = dbTx
		}
	}

	return transactions, nil
}

func (tg *txsGrouper) prepareNormalTxForDB(
	mbHash []byte,
	mb *block.MiniBlock,
	mbStatus string,
	txHash []byte,
	txs map[string]nodeData.TransactionHandler,
	header nodeData.HeaderHandler,
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
	mb *block.MiniBlock,
	header nodeData.HeaderHandler,
	txs map[string]nodeData.TransactionHandler,
	alteredAddresses map[string]*data.AlteredAccount,
) (map[string]*data.Transaction, error) {
	rewardsTxs := make(map[string]*data.Transaction)
	mbHash, err := core.CalculateHash(tg.marshalizer, tg.hasher, mb)
	if err != nil {
		return nil, err
	}

	mbStatus := computeStatus(tg.selfShardID, mb.ReceiverShardID)
	for _, txHash := range mb.TxHashes {
		rewardDBTx, ok := tg.prepareRewardTxForDB(mbHash, mb, mbStatus, txHash, txs, header)
		if !ok {
			continue
		}

		tg.addToAlteredAddresses(rewardDBTx, alteredAddresses, mb, tg.selfShardID, true)
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
	txs map[string]nodeData.TransactionHandler,
	header nodeData.HeaderHandler,
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
	mb *block.MiniBlock,
	header nodeData.HeaderHandler,
	txs map[string]nodeData.TransactionHandler,
	alteredAddresses map[string]*data.AlteredAccount,
) (map[string]*data.Transaction, error) {
	transactions := make(map[string]*data.Transaction)
	mbHash, err := core.CalculateHash(tg.marshalizer, tg.hasher, mb)
	if err != nil {
		return nil, err
	}

	for _, txHash := range mb.TxHashes {
		invalidDBTx, ok := tg.prepareInvalidTxForDB(mbHash, mb, txHash, txs, header)
		if !ok {
			continue
		}

		tg.addToAlteredAddresses(invalidDBTx, alteredAddresses, mb, tg.selfShardID, false)
		transactions[string(txHash)] = invalidDBTx
	}

	return transactions, nil
}

func (tg *txsGrouper) prepareInvalidTxForDB(
	mbHash []byte,
	mb *block.MiniBlock,
	txHash []byte,
	txs map[string]nodeData.TransactionHandler,
	header nodeData.HeaderHandler,
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

func (tg *txsGrouper) groupReceipts(header nodeData.HeaderHandler, txsPool map[string]nodeData.TransactionHandler) []*data.Receipt {
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

func groupSmartContractResults(txsPool map[string]nodeData.TransactionHandler) map[string]*smartContractResult.SmartContractResult {
	scResults := make(map[string]*smartContractResult.SmartContractResult)
	for hash, tx := range txsPool {
		scResult, ok := tx.(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}
		scResults[hash] = scResult
	}

	return scResults
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
	alteredAddresses map[string]*data.AlteredAccount,
	miniBlock *block.MiniBlock,
	selfShardID uint32,
	isRewardTx bool,
) {
	isESDTTx, isNFTTx, nftNonceSTR := tg.txBuilder.computeESDTInfo(tx.Data, tx.EsdtTokenIdentifier)
	if selfShardID == miniBlock.SenderShardID && !isRewardTx {
		alteredAddresses[tx.Sender] = &data.AlteredAccount{
			IsSender:        true,
			IsESDTOperation: isESDTTx,
			IsNFTOperation:  isNFTTx,
			TokenIdentifier: tx.EsdtTokenIdentifier,
			NFTNonceString:  nftNonceSTR,
		}
	}

	ignoreTransactionReceiver := tx.Status == transaction.TxStatusInvalid.String() || tx.Sender == tx.Receiver
	if ignoreTransactionReceiver {
		return
	}

	if selfShardID == miniBlock.ReceiverShardID || miniBlock.ReceiverShardID == core.AllShardId {
		alteredAddresses[tx.Receiver] = &data.AlteredAccount{
			IsSender:        false,
			IsESDTOperation: isESDTTx,
			IsNFTOperation:  isNFTTx,
			TokenIdentifier: tx.EsdtTokenIdentifier,
			NFTNonceString:  nftNonceSTR,
		}
	}
}
