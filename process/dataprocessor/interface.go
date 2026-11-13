package dataprocessor

import (
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

// TransactionsHandler defines what the data processor needs from the transactions processor
type TransactionsHandler interface {
	PrepareTransactionsForDatabase(
		miniBlocks []*block.MiniBlock,
		headerData *data.HeaderData,
		pool *outport.TransactionPool,
		isImportDB bool,
	) *data.PreparedResults
	GetHexEncodedHashesForRemove(headerData *data.HeaderData, body *block.Body) ([]string, []string)
}

// LogsAndEventsHandler defines what the data processor needs from the logs and events processor
type LogsAndEventsHandler interface {
	ExtractDataFromLogs(
		logsAndEvents []*transaction.LogData,
		preparedResults *data.PreparedResults,
		shardID uint32,
		numOfShards uint32,
		timestampMs uint64,
	) *data.PreparedLogsResults
}

// OperationsHandler defines what the data processor needs from the operations processor
type OperationsHandler interface {
	ProcessTransactionsAndSCRs(txs []*data.Transaction, scrs []*data.ScResult, isImportDB bool, shardID uint32) ([]*data.Transaction, []*data.ScResult)
}
