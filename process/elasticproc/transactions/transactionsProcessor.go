package transactions

import (
	"encoding/hex"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
)

var log = logger.GetOrCreate("indexer/process/transactions")

// ArgsTransactionProcessor holds all dependencies required by the txsDatabaseProcessor  in order to create
// new instances
type ArgsTransactionProcessor struct {
	AddressPubkeyConverter core.PubkeyConverter
	Hasher                 hashing.Hasher
	Marshalizer            marshal.Marshalizer
	BalanceConverter       dataindexer.BalanceConverter
	TxHashExtractor        TxHashExtractor
	RewardTxData           RewardTxDataHandler
}

type txsDatabaseProcessor struct {
	txBuilder     *dbTransactionBuilder
	txsGrouper    *txsGrouper
	scrsProc      *smartContractResultsProcessor
	scrsDataToTxs *scrsDataToTransactions
}

// NewTransactionsProcessor will create a new instance of transactions database processor
func NewTransactionsProcessor(args *ArgsTransactionProcessor) (*txsDatabaseProcessor, error) {
	err := checkTxsProcessorArg(args)
	if err != nil {
		return nil, err
	}

	argsParser := &datafield.ArgsOperationDataFieldParser{
		AddressLength: args.AddressPubkeyConverter.Len(),
		Marshalizer:   args.Marshalizer,
	}
	operationsDataParser, err := datafield.NewOperationDataFieldParser(argsParser)
	if err != nil {
		return nil, err
	}

	txBuilder := newTransactionDBBuilder(args.AddressPubkeyConverter, operationsDataParser, args.BalanceConverter, args.RewardTxData)
	txsDBGrouper := newTxsGrouper(txBuilder, args.Hasher, args.Marshalizer, args.TxHashExtractor)
	scrProc := newSmartContractResultsProcessor(args.AddressPubkeyConverter, args.Marshalizer, args.Hasher, operationsDataParser, args.BalanceConverter)
	scrsDataToTxs := newScrsDataToTransactions(args.BalanceConverter)

	return &txsDatabaseProcessor{
		txBuilder:     txBuilder,
		txsGrouper:    txsDBGrouper,
		scrsProc:      scrProc,
		scrsDataToTxs: scrsDataToTxs,
	}, nil
}

// PrepareTransactionsForDatabase will prepare transactions for database
func (tdp *txsDatabaseProcessor) PrepareTransactionsForDatabase(
	miniBlocks []*block.MiniBlock,
	header coreData.HeaderHandler,
	pool *outport.TransactionPool,
	isImportDB bool,
	numOfShards uint32,
) *data.PreparedResults {
	err := checkPrepareTransactionForDatabaseArguments(header, pool)
	if err != nil {
		log.Warn("checkPrepareTransactionForDatabaseArguments", "error", err)

		return &data.PreparedResults{
			Transactions: []*data.Transaction{},
			ScResults:    []*data.ScResult{},
			Receipts:     []*data.Receipt{},
		}
	}

	normalTxs := make(map[string]*data.Transaction)
	rewardsTxs := make(map[string]*data.Transaction)

	for mbIndex, mb := range miniBlocks {
		switch mb.Type {
		case block.TxBlock:
			if shouldIgnoreProcessedMBScheduled(header, mbIndex) {
				continue
			}

			txs, errGroup := tdp.txsGrouper.groupNormalTxs(mbIndex, mb, header, pool.Transactions, isImportDB, numOfShards)
			if errGroup != nil {
				log.Warn("txsDatabaseProcessor.groupNormalTxs", "error", errGroup)
				continue
			}
			mergeTxsMaps(normalTxs, txs)
		case block.RewardsBlock:
			txs, errGroup := tdp.txsGrouper.groupRewardsTxs(mbIndex, mb, header, pool.Rewards, isImportDB)
			if errGroup != nil {
				log.Warn("txsDatabaseProcessor.groupRewardsTxs", "error", errGroup)
				continue
			}
			mergeTxsMaps(rewardsTxs, txs)
		case block.InvalidBlock:
			txs, errGroup := tdp.txsGrouper.groupInvalidTxs(mbIndex, mb, header, pool.InvalidTxs, numOfShards)
			if errGroup != nil {
				log.Warn("txsDatabaseProcessor.groupInvalidTxs", "error", errGroup)
				continue
			}
			mergeTxsMaps(normalTxs, txs)
		default:
			continue
		}
	}

	normalTxs = tdp.setTransactionSearchOrder(normalTxs)
	dbReceipts := tdp.txsGrouper.groupReceipts(header, pool.Receipts)
	dbSCResults := tdp.scrsProc.processSCRs(miniBlocks, header, pool.SmartContractResults, numOfShards)

	srcsNoTxInCurrentShard := tdp.scrsDataToTxs.attachSCRsToTransactionsAndReturnSCRsWithoutTx(normalTxs, dbSCResults)
	tdp.scrsDataToTxs.processTransactionsAfterSCRsWereAttached(normalTxs)
	txHashFee := tdp.scrsDataToTxs.processSCRsWithoutTx(srcsNoTxInCurrentShard)

	sliceNormalTxs := convertMapTxsToSlice(normalTxs)
	sliceRewardsTxs := convertMapTxsToSlice(rewardsTxs)
	txsSlice := append(sliceNormalTxs, sliceRewardsTxs...)

	return &data.PreparedResults{
		Transactions: txsSlice,
		ScResults:    dbSCResults,
		Receipts:     dbReceipts,
		TxHashFee:    txHashFee,
	}
}

func (tdp *txsDatabaseProcessor) setTransactionSearchOrder(transactions map[string]*data.Transaction) map[string]*data.Transaction {
	currentOrder := uint32(0)
	for _, tx := range transactions {
		tx.SearchOrder = currentOrder
		currentOrder++
	}

	return transactions
}

// GetHexEncodedHashesForRemove will return hex encoded transaction hashes and smart contract result hashes from body
func (tdp *txsDatabaseProcessor) GetHexEncodedHashesForRemove(header coreData.HeaderHandler, body *block.Body) ([]string, []string) {
	if body == nil || check.IfNil(header) || len(header.GetMiniBlockHeadersHashes()) == 0 {
		return nil, nil
	}

	selfShardID := header.GetShardID()
	encodedTxsHashes := make([]string, 0)
	encodedScrsHashes := make([]string, 0)
	for mbIndex, miniblock := range body.MiniBlocks {
		if shouldIgnoreProcessedMBScheduled(header, mbIndex) {
			continue
		}

		shouldIgnore := isCrossShardAtSourceNormalTx(selfShardID, miniblock)
		if shouldIgnore {
			// ignore cross-shard miniblocks at source with normal txs
			continue
		}

		txsHashesFromMiniblock := getTxsHashesFromMiniblockHexEncoded(miniblock)
		if miniblock.Type == block.SmartContractResultBlock {
			encodedScrsHashes = append(encodedScrsHashes, txsHashesFromMiniblock...)
			continue
		}
		encodedTxsHashes = append(encodedTxsHashes, txsHashesFromMiniblock...)
	}

	return encodedTxsHashes, encodedScrsHashes
}

func isCrossShardAtSourceNormalTx(selfShardID uint32, miniblock *block.MiniBlock) bool {
	isCrossShard := miniblock.SenderShardID != miniblock.ReceiverShardID
	isAtSource := miniblock.SenderShardID == selfShardID
	txBlock := miniblock.Type == block.TxBlock

	return isCrossShard && isAtSource && txBlock
}

func shouldIgnoreProcessedMBScheduled(header coreData.HeaderHandler, mbIndex int) bool {
	miniBlockHeaders := header.GetMiniBlockHeaderHandlers()
	if len(miniBlockHeaders) <= mbIndex {
		return false
	}

	processingType := miniBlockHeaders[mbIndex].GetProcessingType()

	return processingType == int32(block.Processed)
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

func getFeeInfo(txWithFeeInfo feeInfoHandler) *outport.FeeInfo {
	feeInfo := txWithFeeInfo.GetFeeInfo()
	if feeInfo != nil {
		return feeInfo
	}

	return &outport.FeeInfo{
		Fee:            big.NewInt(0),
		InitialPaidFee: big.NewInt(0),
	}
}
