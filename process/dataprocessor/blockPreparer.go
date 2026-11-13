package dataprocessor

import (
	"encoding/hex"

	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/converters"
)

// PrepareBlock prepares, without any access to a database, the transactions, smart contract results, logs, events
// and operations of the provided block. The header is the first unit, followed by one unit for every execution result
func (dp *dataProcessor) PrepareBlock(obh *outport.OutportBlockWithHeader) *data.PreparedBlock {
	isImportDB := dp.isImportDB()

	miniBlocks := append(obh.BlockData.Body.MiniBlocks, obh.BlockData.IntraShardMiniBlocks...)
	headerData := &data.HeaderData{
		Timestamp:        converters.MillisecondsToSeconds(obh.BlockData.TimestampMs),
		TimestampMs:      obh.BlockData.TimestampMs,
		Round:            obh.Header.GetRound(),
		ShardID:          obh.Header.GetShardID(),
		Epoch:            obh.Header.GetEpoch(),
		MiniBlockHeaders: obh.Header.GetMiniBlockHeaderHandlers(),
		NumberOfShards:   obh.NumberOfShards,
	}

	executionResults := obh.Header.GetExecutionResultsHandlers()
	units := make([]*data.PreparedBlockUnit, 0, len(executionResults)+1)
	// the block of a header v3 is executed later, the node will send its data as an execution result of another header
	isHeaderExecuted := !obh.Header.IsHeaderV3()
	units = append(units, dp.prepareBlockUnit(obh.BlockData.HeaderHash, isHeaderExecuted, headerData, miniBlocks, obh.TransactionPool, obh.AlteredAccounts, isImportDB))

	for _, executionResult := range executionResults {
		executionResultData, found := obh.BlockData.Results[hex.EncodeToString(executionResult.GetHeaderHash())]
		if !found {
			log.Warn("dataProcessor.PrepareBlock: cannot find execution result data", "hash", executionResult.GetHeaderHash())
			continue
		}

		headerData = &data.HeaderData{
			Timestamp:        converters.MillisecondsToSeconds(executionResultData.TimestampMs),
			TimestampMs:      executionResultData.TimestampMs,
			Round:            executionResult.GetHeaderRound(),
			ShardID:          obh.Header.GetShardID(),
			NumberOfShards:   obh.NumberOfShards,
			Epoch:            executionResult.GetHeaderEpoch(),
			MiniBlockHeaders: converters.GetMiniBlocksHeaderHandlersFromExecResult(executionResult),
		}

		miniBlocks = append(executionResultData.Body.MiniBlocks, executionResultData.IntraShardMiniBlocks...)
		units = append(units, dp.prepareBlockUnit(
			executionResult.GetHeaderHash(),
			true,
			headerData,
			miniBlocks,
			executionResultData.TransactionPool,
			executionResultData.AlteredAccounts,
			isImportDB,
		))
	}

	return &data.PreparedBlock{
		OutportBlock: obh,
		Units:        units,
	}
}

func (dp *dataProcessor) prepareBlockUnit(
	blockHash []byte,
	executed bool,
	headerData *data.HeaderData,
	miniBlocks []*block.MiniBlock,
	pool *outport.TransactionPool,
	alteredAccounts map[string]*alteredAccount.AlteredAccount,
	isImportDB bool,
) *data.PreparedBlockUnit {
	// the order is important: the logs processor updates the prepared transactions and smart contract results
	// and the operations are copies of them
	preparedResults := dp.transactionsProc.PrepareTransactionsForDatabase(miniBlocks, headerData, pool, isImportDB)
	logsData := dp.logsAndEventsProc.ExtractDataFromLogs(pool.Logs, preparedResults, headerData.ShardID, headerData.NumberOfShards, headerData.TimestampMs)
	operationsTxs, operationsSCRs := dp.operationsProc.ProcessTransactionsAndSCRs(preparedResults.Transactions, preparedResults.ScResults, isImportDB, headerData.ShardID)

	return &data.PreparedBlockUnit{
		BlockHash:       blockHash,
		Executed:        executed,
		HeaderData:      headerData,
		Results:         preparedResults,
		Logs:            logsData,
		OperationsTxs:   operationsTxs,
		OperationsSCRs:  operationsSCRs,
		AlteredAccounts: alteredAccounts,
	}
}

// PrepareRevertedBlock prepares the data needed by the outputs in order to revert the provided block
func (dp *dataProcessor) PrepareRevertedBlock(header coreData.HeaderHandler, blockData *outport.BlockData) *data.RevertedBlock {
	headerData := &data.HeaderData{
		Timestamp:        header.GetTimeStamp(),
		TimestampMs:      blockData.TimestampMs,
		Round:            header.GetRound(),
		ShardID:          header.GetShardID(),
		Epoch:            header.GetEpoch(),
		MiniBlockHeaders: header.GetMiniBlockHeaderHandlers(),
	}
	txHashes, scrHashes := dp.transactionsProc.GetHexEncodedHashesForRemove(headerData, blockData.Body)

	return &data.RevertedBlock{
		Header:    header,
		BlockData: blockData,
		TxHashes:  txHashes,
		ScrHashes: scrHashes,
	}
}
