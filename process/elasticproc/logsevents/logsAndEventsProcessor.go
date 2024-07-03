package logsevents

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
)

const eventIDFormat = "%s-%d-%d"

// ArgsLogsAndEventsProcessor  holds all dependencies required to create new instances of logsAndEventsProcessor
type ArgsLogsAndEventsProcessor struct {
	PubKeyConverter  core.PubkeyConverter
	Marshalizer      marshal.Marshalizer
	BalanceConverter dataindexer.BalanceConverter
	Hasher           hashing.Hasher
}

type logsAndEventsProcessor struct {
	hasher           hashing.Hasher
	pubKeyConverter  core.PubkeyConverter
	eventsProcessors []eventsProcessor
}

// NewLogsAndEventsProcessor will create a new instance for the logsAndEventsProcessor
func NewLogsAndEventsProcessor(args ArgsLogsAndEventsProcessor) (*logsAndEventsProcessor, error) {
	err := checkArgsLogsAndEventsProcessor(args)
	if err != nil {
		return nil, err
	}

	eventsProcessors := createEventsProcessors(args)

	return &logsAndEventsProcessor{
		pubKeyConverter:  args.PubKeyConverter,
		eventsProcessors: eventsProcessors,
		hasher:           args.Hasher,
	}, nil
}

func checkArgsLogsAndEventsProcessor(args ArgsLogsAndEventsProcessor) error {
	if check.IfNil(args.PubKeyConverter) {
		return dataindexer.ErrNilPubkeyConverter
	}
	if check.IfNil(args.Marshalizer) {
		return dataindexer.ErrNilMarshalizer
	}
	if check.IfNil(args.BalanceConverter) {
		return dataindexer.ErrNilBalanceConverter
	}
	if check.IfNil(args.Hasher) {
		return dataindexer.ErrNilHasher
	}

	return nil
}

func createEventsProcessors(args ArgsLogsAndEventsProcessor) []eventsProcessor {
	nftsProc := newNFTsProcessor(args.PubKeyConverter, args.Marshalizer)
	scDeploysProc := newSCDeploysProcessor(args.PubKeyConverter)
	informativeProc := newInformativeLogsProcessor()
	updateNFTProc := newNFTsPropertiesProcessor(args.PubKeyConverter, args.Marshalizer)
	esdtPropProc := newEsdtPropertiesProcessor(args.PubKeyConverter)
	esdtIssueProc := newESDTIssueProcessor(args.PubKeyConverter)
	delegatorsProcessor := newDelegatorsProcessor(args.PubKeyConverter, args.BalanceConverter)

	eventsProcs := []eventsProcessor{
		scDeploysProc,
		informativeProc,
		updateNFTProc,
		esdtPropProc,
		esdtIssueProc,
		delegatorsProcessor,
		nftsProc,
	}

	return eventsProcs
}

// ExtractDataFromLogs will extract data from the provided logs and events and put in altered addresses
func (lep *logsAndEventsProcessor) ExtractDataFromLogs(
	logsAndEvents []*outport.LogData,
	preparedResults *data.PreparedResults,
	timestamp uint64,
	shardID uint32,
	numOfShards uint32,
) *data.PreparedLogsResults {
	lgData := newLogsData(timestamp, preparedResults.Transactions, preparedResults.ScResults)
	for _, txLog := range logsAndEvents {
		if txLog == nil {
			continue
		}

		events := txLog.Log.Events
		lep.processEvents(lgData, txLog.TxHash, txLog.Log.Address, events, shardID, numOfShards)

		tx, ok := lgData.txsMap[txLog.TxHash]
		if ok {
			tx.HasLogs = true
			continue
		}
		scr, ok := lgData.scrsMap[txLog.TxHash]
		if ok {
			scr.HasLogs = true
			continue
		}
	}

	dbLogs, dbEvents := lep.prepareLogsForDB(lgData, logsAndEvents, timestamp, shardID)

	return &data.PreparedLogsResults{
		Tokens:                  lgData.tokens,
		ScDeploys:               lgData.scDeploys,
		TokensInfo:              lgData.tokensInfo,
		TokensSupply:            lgData.tokensSupply,
		Delegators:              lgData.delegators,
		NFTsDataUpdates:         lgData.nftsDataUpdates,
		TokenRolesAndProperties: lgData.tokenRolesAndProperties,
		TxHashStatusInfo:        lgData.txHashStatusInfoProc.getAllRecords(),
		ChangeOwnerOperations:   lgData.changeOwnerOperations,
		DBLogs:                  dbLogs,
		DBEvents:                dbEvents,
	}
}

func (lep *logsAndEventsProcessor) processEvents(lgData *logsData, logHashHexEncoded string, logAddress []byte, events []*transaction.Event, shardID uint32, numOfShards uint32) {
	for _, event := range events {
		if check.IfNil(event) {
			continue
		}

		lep.processEvent(lgData, logHashHexEncoded, logAddress, event, shardID, numOfShards)
	}
}

func (lep *logsAndEventsProcessor) processEvent(lgData *logsData, logHashHexEncoded string, logAddress []byte, event coreData.EventHandler, shardID uint32, numOfShards uint32) {
	for _, proc := range lep.eventsProcessors {
		res := proc.processEvent(&argsProcessEvent{
			event:                   event,
			txHashHexEncoded:        logHashHexEncoded,
			logAddress:              logAddress,
			tokens:                  lgData.tokens,
			tokensSupply:            lgData.tokensSupply,
			timestamp:               lgData.timestamp,
			scDeploys:               lgData.scDeploys,
			txs:                     lgData.txsMap,
			scrs:                    lgData.scrsMap,
			tokenRolesAndProperties: lgData.tokenRolesAndProperties,
			txHashStatusInfoProc:    lgData.txHashStatusInfoProc,
			changeOwnerOperations:   lgData.changeOwnerOperations,
			selfShardID:             shardID,
			numOfShards:             numOfShards,
		})
		if res.tokenInfo != nil {
			lgData.tokensInfo = append(lgData.tokensInfo, res.tokenInfo)
		}
		if res.delegator != nil {
			lgData.delegators[res.delegator.Address+res.delegator.Contract] = res.delegator
		}
		if res.updatePropNFT != nil {
			lgData.nftsDataUpdates = append(lgData.nftsDataUpdates, res.updatePropNFT)
		}

		tx, ok := lgData.txsMap[logHashHexEncoded]
		if ok {
			tx.HasOperations = true
			continue
		}
		scr, ok := lgData.scrsMap[logHashHexEncoded]
		if ok {
			scr.HasOperations = true
			continue
		}

		if res.processed {
			return
		}
	}
}

func (lep *logsAndEventsProcessor) prepareLogsForDB(
	lgData *logsData,
	logsAndEvents []*outport.LogData,
	timestamp uint64,
	shardID uint32,
) ([]*data.Logs, []*data.LogEvent) {
	logs := make([]*data.Logs, 0, len(logsAndEvents))
	events := make([]*data.LogEvent, 0)

	for _, txLog := range logsAndEvents {
		if txLog == nil {
			continue
		}

		dbLog, logEvents := lep.prepareLog(lgData, txLog.TxHash, txLog.Log, timestamp, shardID)

		logs = append(logs, dbLog)
		events = append(events, logEvents...)
	}

	return logs, events
}

func (lep *logsAndEventsProcessor) prepareLog(
	lgData *logsData,
	logHashHex string,
	eventLogs *transaction.Log,
	timestamp uint64,
	shardID uint32,
) (*data.Logs, []*data.LogEvent) {
	originalTxHash := lep.getOriginalTxHash(lgData, logHashHex)
	encodedAddr := lep.pubKeyConverter.SilentEncode(eventLogs.GetAddress(), log)
	logsDB := &data.Logs{
		ID:             logHashHex,
		OriginalTxHash: originalTxHash,
		Address:        encodedAddr,
		Timestamp:      time.Duration(timestamp),
		Events:         make([]*data.Event, 0, len(eventLogs.Events)),
	}

	dbEvents := make([]*data.LogEvent, 0, len(eventLogs.Events))
	for idx, event := range eventLogs.Events {
		if check.IfNil(event) {
			continue
		}

		logEvent := &data.Event{
			Address:        lep.pubKeyConverter.SilentEncode(event.GetAddress(), log),
			Identifier:     string(event.GetIdentifier()),
			Topics:         event.GetTopics(),
			Data:           event.GetData(),
			AdditionalData: event.GetAdditionalData(),
			Order:          idx,
		}
		logsDB.Events = append(logsDB.Events, logEvent)

		executionOrder := lep.getExecutionOrder(lgData, logHashHex)
		dbEvents = append(dbEvents, lep.prepareLogEvent(logsDB, logEvent, shardID, executionOrder))
	}

	return logsDB, dbEvents
}

func (lep *logsAndEventsProcessor) prepareLogEvent(dbLog *data.Logs, event *data.Event, shardID uint32, execOrder int) *data.LogEvent {
	dbEvent := &data.LogEvent{
		TxHash:         dbLog.ID,
		LogAddress:     dbLog.Address,
		Address:        event.Address,
		Identifier:     event.Identifier,
		Data:           hex.EncodeToString(event.Data),
		AdditionalData: hexEncodeSlice(event.AdditionalData),
		Topics:         hexEncodeSlice(event.Topics),
		Order:          event.Order,
		ShardID:        shardID,
		TxOrder:        execOrder,
		OriginalTxHash: dbLog.OriginalTxHash,
		Timestamp:      dbLog.Timestamp,
		ID:             fmt.Sprintf(eventIDFormat, dbLog.ID, shardID, event.Order),
	}

	return dbEvent
}

func (lep *logsAndEventsProcessor) getOriginalTxHash(lgData *logsData, logHashHex string) string {
	if lgData.scrsMap == nil {
		return ""
	}

	scr, ok := lgData.scrsMap[logHashHex]
	if ok {
		return scr.OriginalTxHash
	}

	return ""
}

func (lep *logsAndEventsProcessor) getExecutionOrder(lgData *logsData, logHashHex string) int {
	tx, ok := lgData.txsMap[logHashHex]
	if ok {
		return tx.ExecutionOrder
	}

	scr, ok := lgData.scrsMap[logHashHex]
	if ok {
		return scr.ExecutionOrder
	}

	log.Warn("cannot find hash in the txs map or scrs map", "hash", logHashHex)

	return -1
}

func hexEncodeSlice(input [][]byte) []string {
	hexEncoded := make([]string, 0, len(input))
	for idx := 0; idx < len(input); idx++ {
		hexEncoded = append(hexEncoded, hex.EncodeToString(input[idx]))
	}
	if len(hexEncoded) == 0 {
		return nil
	}

	return hexEncoded
}
