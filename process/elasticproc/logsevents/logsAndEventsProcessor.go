package logsevents

import (
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

	logsData *logsData
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
	updateNFTProc := newNFTsPropertiesProcessor(args.PubKeyConverter)
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
	lep.logsData = newLogsData(timestamp, preparedResults.Transactions, preparedResults.ScResults)

	for _, txLog := range logsAndEvents {
		if txLog == nil {
			continue
		}

		events := txLog.Log.Events
		lep.processEvents(txLog.TxHash, txLog.Log.Address, events, shardID, numOfShards)

		tx, ok := lep.logsData.txsMap[txLog.TxHash]
		if ok {
			tx.HasLogs = true
			continue
		}
		scr, ok := lep.logsData.scrsMap[txLog.TxHash]
		if ok {
			scr.HasLogs = true
			continue
		}
	}

	return &data.PreparedLogsResults{
		Tokens:                  lep.logsData.tokens,
		ScDeploys:               lep.logsData.scDeploys,
		TokensInfo:              lep.logsData.tokensInfo,
		TokensSupply:            lep.logsData.tokensSupply,
		Delegators:              lep.logsData.delegators,
		NFTsDataUpdates:         lep.logsData.nftsDataUpdates,
		TokenRolesAndProperties: lep.logsData.tokenRolesAndProperties,
		TxHashStatusInfo:        lep.logsData.txHashStatusInfoProc.getAllRecords(),
		ChangeOwnerOperations:   lep.logsData.changeOwnerOperations,
	}
}

func (lep *logsAndEventsProcessor) processEvents(logHashHexEncoded string, logAddress []byte, events []*transaction.Event, shardID uint32, numOfShards uint32) {
	for _, event := range events {
		if check.IfNil(event) {
			continue
		}

		lep.processEvent(logHashHexEncoded, logAddress, event, shardID, numOfShards)
	}
}

func (lep *logsAndEventsProcessor) processEvent(logHashHexEncoded string, logAddress []byte, event coreData.EventHandler, shardID uint32, numOfShards uint32) {
	for _, proc := range lep.eventsProcessors {
		res := proc.processEvent(&argsProcessEvent{
			event:                   event,
			txHashHexEncoded:        logHashHexEncoded,
			logAddress:              logAddress,
			tokens:                  lep.logsData.tokens,
			tokensSupply:            lep.logsData.tokensSupply,
			timestamp:               lep.logsData.timestamp,
			scDeploys:               lep.logsData.scDeploys,
			txs:                     lep.logsData.txsMap,
			scrs:                    lep.logsData.scrsMap,
			tokenRolesAndProperties: lep.logsData.tokenRolesAndProperties,
			txHashStatusInfoProc:    lep.logsData.txHashStatusInfoProc,
			changeOwnerOperations:   lep.logsData.changeOwnerOperations,
			selfShardID:             shardID,
			numOfShards:             numOfShards,
		})
		if res.tokenInfo != nil {
			lep.logsData.tokensInfo = append(lep.logsData.tokensInfo, res.tokenInfo)
		}
		if res.delegator != nil {
			lep.logsData.delegators[res.delegator.Address+res.delegator.Contract] = res.delegator
		}
		if res.updatePropNFT != nil {
			lep.logsData.nftsDataUpdates = append(lep.logsData.nftsDataUpdates, res.updatePropNFT)
		}

		tx, ok := lep.logsData.txsMap[logHashHexEncoded]
		if ok {
			tx.HasOperations = true
			continue
		}
		scr, ok := lep.logsData.scrsMap[logHashHexEncoded]
		if ok {
			scr.HasOperations = true
			continue
		}

		if res.processed {
			return
		}
	}
}

// PrepareLogsForDB will prepare logs for database
func (lep *logsAndEventsProcessor) PrepareLogsForDB(
	logsAndEvents []*outport.LogData,
	timestamp uint64,
) []*data.Logs {
	logs := make([]*data.Logs, 0, len(logsAndEvents))

	for _, txLog := range logsAndEvents {
		if txLog == nil {
			continue
		}

		logs = append(logs, lep.prepareLogsForDB(txLog.TxHash, txLog.Log, timestamp))
	}

	return logs
}

func (lep *logsAndEventsProcessor) prepareLogsForDB(
	logHashHex string,
	logHandler coreData.LogHandler,
	timestamp uint64,
) *data.Logs {
	originalTxHash := ""
	scr, ok := lep.logsData.scrsMap[logHashHex]
	if ok {
		originalTxHash = scr.OriginalTxHash
	}

	events := logHandler.GetLogEvents()

	encodedAddr := lep.pubKeyConverter.SilentEncode(logHandler.GetAddress(), log)

	logsDB := &data.Logs{
		ID:             logHashHex,
		OriginalTxHash: originalTxHash,
		Address:        encodedAddr,
		Timestamp:      time.Duration(timestamp),
		Events:         make([]*data.Event, 0, len(events)),
	}

	for idx, event := range events {
		if check.IfNil(event) {
			continue
		}

		encodedAddress := lep.pubKeyConverter.SilentEncode(event.GetAddress(), log)

		logsDB.Events = append(logsDB.Events, &data.Event{
			Address:    encodedAddress,
			Identifier: string(event.GetIdentifier()),
			Topics:     event.GetTopics(),
			Data:       event.GetData(),
			Order:      idx,
		})
	}

	return logsDB
}
