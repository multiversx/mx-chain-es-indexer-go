package transactions

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/check"
	"github.com/ElrondNetwork/elrond-go/core"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const (
	// A smart contract action (deploy, call, ...) should have minimum 2 smart contract results
	// exception to this rule are smart contract calls to ESDT contract
	minimumNumberOfSmartContractResults = 2
)

var log = logger.GetOrCreate("indexer/process/transactions")

type ArgsTransactionProcessor struct {
	AddressPubkeyConverter core.PubkeyConverter
	TxFeeCalculator        process.TransactionFeeCalculator
	ShardCoordinator       sharding.Coordinator
	Hasher                 hashing.Hasher
	Marshalizer            marshal.Marshalizer
	IsInImportMode         bool
}

type txDatabaseProcessor struct {
	txFeeCalculator process.TransactionFeeCalculator
	txBuilder       *dbTransactionBuilder
	txsGrouper      *txsGrouper
}

// NewTransactionsProcessor will create a new instance of transactions database processor
func NewTransactionsProcessor(args *ArgsTransactionProcessor) (*txDatabaseProcessor, error) {
	err := checkTxsProcessorArg(args)
	if err != nil {
		return nil, err
	}

	txBuilder := newTransactionDBBuilder(args.AddressPubkeyConverter, args.ShardCoordinator, args.TxFeeCalculator)
	txsDBGrouper := newTxsGrouper(txBuilder, args.IsInImportMode, args.ShardCoordinator.SelfId(), args.Hasher, args.Marshalizer)

	if args.IsInImportMode {
		log.Warn("the node is in import mode! Cross shard transactions and rewards where destination shard is " +
			"not the current node's shard won't be indexed in Elastic Search")
	}

	return &txDatabaseProcessor{
		txFeeCalculator: args.TxFeeCalculator,
		txBuilder:       txBuilder,
		txsGrouper:      txsDBGrouper,
	}, nil
}

// PrepareTransactionsForDatabase will prepare transactions for database
func (tdp *txDatabaseProcessor) PrepareTransactionsForDatabase(
	body *block.Body,
	header nodeData.HeaderHandler,
	pool *indexer.Pool,
) *data.PreparedResults {
	err := checkPrepareTransactionForDatabaseArguments(body, header, pool)
	if err != nil {
		log.Warn("checkPrepareTransactionForDatabaseArguments", "error", err)

		return &data.PreparedResults{
			Transactions:    []*data.Transaction{},
			ScResults:       []*data.ScResult{},
			Receipts:        []*data.Receipt{},
			AlteredAccounts: map[string]*data.AlteredAccount{},
		}
	}

	alteredAddresses := make(map[string]*data.AlteredAccount)
	normalTxs := make(map[string]*data.Transaction)
	rewardsTxs := make(map[string]*data.Transaction)
	invalidTxs := make(map[string]*data.Transaction)

	for _, mb := range body.MiniBlocks {
		switch mb.Type {
		case block.TxBlock:
			mergeTxsMaps(normalTxs, tdp.txsGrouper.groupNormalTxs(mb, header, pool.Txs, alteredAddresses))
		case block.RewardsBlock:
			mergeTxsMaps(rewardsTxs, tdp.txsGrouper.groupRewardsTxs(mb, header, pool.Rewards, alteredAddresses))
		case block.InvalidBlock:
			mergeTxsMaps(invalidTxs, tdp.txsGrouper.groupInvalidTxs(mb, header, pool.Invalid, alteredAddresses))
		default:
			continue
		}
	}

	normalTxs = tdp.setTransactionSearchOrder(normalTxs)
	dbReceipts := tdp.txsGrouper.groupReceipts(header, pool.Receipts)
	dbSCResults, countScResults := tdp.iterateSCRSAndConvert(pool.Scrs, header, normalTxs)

	tdp.txBuilder.addScrsReceiverToAlteredAccounts(alteredAddresses, dbSCResults)
	tdp.setDetailsOfTxsWithSCRS(normalTxs, countScResults)

	sliceNormalTxs := convertMapTxsToSlice(normalTxs)
	sliceRewardsTxs := convertMapTxsToSlice(rewardsTxs)
	sliceInvalidTxs := convertMapTxsToSlice(invalidTxs)
	txsSlice := append(sliceNormalTxs, append(sliceRewardsTxs, sliceInvalidTxs...)...)

	return &data.PreparedResults{
		Transactions:    txsSlice,
		ScResults:       dbSCResults,
		Receipts:        dbReceipts,
		AlteredAccounts: alteredAddresses,
	}
}

func (tdp *txDatabaseProcessor) setDetailsOfTxsWithSCRS(
	transactions map[string]*data.Transaction,
	countScResults map[string]int,
) {
	for hash, nrScResults := range countScResults {
		tx, ok := transactions[hash]
		if !ok {
			continue
		}

		tdp.setDetailsOfATxWithSCRS(tx, nrScResults)
	}
}

func (tdp *txDatabaseProcessor) setDetailsOfATxWithSCRS(tx *data.Transaction, nrScResults int) {
	tx.HasSCR = true

	if isRelayedTx(tx) {
		tx.GasUsed = tx.GasLimit
		fee := tdp.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, tx.GasUsed)
		tx.Fee = fee.String()

		return
	}

	if nrScResults > minimumNumberOfSmartContractResults {
		return
	}

	if len(tx.SmartContractResults) > 0 {
		scResultData := tx.SmartContractResults[0].Data
		if isScResultSuccessful(scResultData) {
			// ESDT contract calls generate only one smart contract result
			return
		}
	}

	tx.Status = transaction.TxStatusFail.String()
	tx.GasUsed = tx.GasLimit
	fee := tdp.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, tx.GasUsed)
	tx.Fee = fee.String()
}

func (tdp *txDatabaseProcessor) iterateSCRSAndConvert(
	txPool map[string]nodeData.TransactionHandler,
	header nodeData.HeaderHandler,
	transactions map[string]*data.Transaction,
) ([]*data.ScResult, map[string]int) {
	// we can not iterate smart contract results directly on the miniblocks contained in the block body
	// as some miniblocks might be missing. Example: intra-shard miniblock that holds smart contract results
	scResults := groupSmartContractResults(txPool)

	dbSCResults := make([]*data.ScResult, 0)
	countScResults := make(map[string]int)
	for scHash, scResult := range scResults {
		dbScResult := tdp.txBuilder.prepareSmartContractResult(scHash, scResult, header)
		dbSCResults = append(dbSCResults, dbScResult)

		tx, ok := transactions[string(scResult.OriginalTxHash)]
		if !ok {
			continue
		}

		tx = tdp.addScResultInfoInTx(dbScResult, tx)
		countScResults[string(scResult.OriginalTxHash)]++
		delete(scResults, scHash)

		// append child smart contract results
		childSCRS := findAllChildScrResults(scHash, scResults)

		tdp.addScResultsInTx(tx, header, childSCRS)

		countScResults[string(scResult.OriginalTxHash)] += len(childSCRS)
	}

	return dbSCResults, countScResults
}

func (tdp *txDatabaseProcessor) addScResultsInTx(tx *data.Transaction, header nodeData.HeaderHandler, scrs map[string]*smartContractResult.SmartContractResult) {
	for childScHash, sc := range scrs {
		childDBScResult := tdp.txBuilder.prepareSmartContractResult(childScHash, sc, header)

		tx = tdp.addScResultInfoInTx(childDBScResult, tx)
	}
}

func findAllChildScrResults(hash string, scrs map[string]*smartContractResult.SmartContractResult) map[string]*smartContractResult.SmartContractResult {
	scrResults := make(map[string]*smartContractResult.SmartContractResult)
	for scrHash, scr := range scrs {
		if string(scr.OriginalTxHash) == hash {
			scrResults[scrHash] = scr
			delete(scrs, scrHash)
		}
	}

	return scrResults
}

func (tdp *txDatabaseProcessor) addScResultInfoInTx(dbScResult *data.ScResult, tx *data.Transaction) *data.Transaction {
	tx.SmartContractResults = append(tx.SmartContractResults, dbScResult)

	if isSCRForSenderWithRefund(dbScResult, tx) {
		refundValue := stringValueToBigInt(dbScResult.Value)
		gasUsed, fee := tdp.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refundValue)
		tx.GasUsed = gasUsed
		tx.Fee = fee.String()
	}

	return tx
}

func (tdp *txDatabaseProcessor) setTransactionSearchOrder(transactions map[string]*data.Transaction) map[string]*data.Transaction {
	currentOrder := uint32(0)
	for _, tx := range transactions {
		tx.SearchOrder = currentOrder
		currentOrder++
	}

	return transactions
}

// GetRewardsTxsHashesHexEncoded will return reward transactions hashes from body hex encoded
func (tdp *txDatabaseProcessor) GetRewardsTxsHashesHexEncoded(header nodeData.HeaderHandler, body *block.Body) []string {
	if body == nil || check.IfNil(header) || len(header.GetMiniBlockHeadersHashes()) == 0 {
		return nil
	}

	selfShardID := header.GetShardID()
	encodedTxsHashes := make([]string, 0)
	for _, miniblock := range body.MiniBlocks {
		if miniblock.Type != block.RewardsBlock {
			continue
		}

		isDstMe := selfShardID == miniblock.ReceiverShardID
		if isDstMe {
			// reward miniblock is always cross-shard
			continue
		}

		txsHashesFromMiniblock := getTxsHashesFromMiniblockHexEncoded(miniblock)
		encodedTxsHashes = append(encodedTxsHashes, txsHashesFromMiniblock...)
	}

	return encodedTxsHashes
}

func getTxsHashesFromMiniblockHexEncoded(miniBlock *block.MiniBlock) []string {
	encodedTxsHashes := make([]string, 0)
	for _, txHash := range miniBlock.TxHashes {
		encodedTxsHashes = append(encodedTxsHashes, hex.EncodeToString(txHash))
	}

	return encodedTxsHashes
}

func mergeTxsMaps(dst, src map[string]*data.Transaction) {
	for key, value := range src {
		dst[key] = value
	}
}
