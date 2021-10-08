package logsevents

import (
	"encoding/hex"
	"time"

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

type logsAndEventsProcessor struct {
	pubKeyConverter  core.PubkeyConverter
	eventsProcessors []eventsProcessor

	logsData *logsData
}

// NewLogsAndEventsProcessor will create a new instance for the logsAndEventsProcessor
func NewLogsAndEventsProcessor(
	shardCoordinator elasticIndexer.ShardCoordinator,
	pubKeyConverter core.PubkeyConverter,
	marshalizer marshal.Marshalizer,
) (*logsAndEventsProcessor, error) {
	if check.IfNil(shardCoordinator) {
		return nil, elasticIndexer.ErrNilShardCoordinator
	}
	if check.IfNil(pubKeyConverter) {
		return nil, elasticIndexer.ErrNilPubkeyConverter
	}
	if check.IfNil(marshalizer) {
		return nil, elasticIndexer.ErrNilMarshalizer
	}

	nftsProc := newNFTsProcessor(shardCoordinator, pubKeyConverter, marshalizer)
	fungibleProc := newFungibleESDTProcessor(pubKeyConverter, shardCoordinator)
	scDeploysProc := newSCDeploysProcessor(pubKeyConverter)
	issueESDTProc := newESDTIssueProcessor()

	return &logsAndEventsProcessor{
		pubKeyConverter: pubKeyConverter,
		eventsProcessors: []eventsProcessor{
			fungibleProc,
			nftsProc,
			scDeploysProc,
			issueESDTProc,
		},
	}, nil
}

// ExtractDataFromLogs will extract data from the provided logs and events and put in altered addresses
func (lep *logsAndEventsProcessor) ExtractDataFromLogs(
	logsAndEvents map[string]coreData.LogHandler,
	preparedResults *data.PreparedResults,
	timestamp uint64,
) *data.PreparedLogsResults {
	lep.logsData = newLogsData(timestamp, preparedResults.AlteredAccts, preparedResults.Transactions, preparedResults.ScResults)

	for logHash, txLog := range logsAndEvents {
		if check.IfNil(txLog) {
			continue
		}

		events := txLog.GetLogEvents()
		lep.processEvents(logHash, events)
	}

	return &data.PreparedLogsResults{
		Tokens:          lep.logsData.tokens,
		ScDeploys:       lep.logsData.scDeploys,
		TagsCount:       lep.logsData.tagsCount,
		PendingBalances: lep.logsData.pendingBalances.getAll(),
		TokensInfo:      lep.logsData.tokensInfo,
	}
}

func (lep *logsAndEventsProcessor) processEvents(logHash string, events []coreData.EventHandler) {
	for _, event := range events {
		if check.IfNil(event) {
			continue
		}

		lep.processEvent(logHash, event)
	}
}

func (lep *logsAndEventsProcessor) processEvent(logHash string, events coreData.EventHandler) {
	logHashHexEncoded := hex.EncodeToString([]byte(logHash))
	for _, proc := range lep.eventsProcessors {
		res := proc.processEvent(&argsProcessEvent{
			event:            events,
			txHashHexEncoded: logHashHexEncoded,
			accounts:         lep.logsData.accounts,
			tokens:           lep.logsData.tokens,
			tagsCount:        lep.logsData.tagsCount,
			timestamp:        lep.logsData.timestamp,
			scDeploys:        lep.logsData.scDeploys,
			pendingBalances:  lep.logsData.pendingBalances,
		})
		if res.tokenInfo != nil {
			lep.logsData.tokensInfo = append(lep.logsData.tokensInfo, res.tokenInfo)
		}

		isEmptyIdentifier := res.identifier == ""
		if isEmptyIdentifier && res.processed {
			return
		}

		tx, ok := lep.logsData.txsMap[logHashHexEncoded]
		if ok && !isEmptyIdentifier {
			tx.HasOperations = true
			tx.Tokens = append(tx.Tokens, res.identifier)
			tx.ESDTValues = append(tx.ESDTValues, res.value)
			continue
		}

		scr, ok := lep.logsData.scrsMap[logHashHexEncoded]
		if ok && !isEmptyIdentifier {
			scr.Tokens = append(scr.Tokens, res.identifier)
			scr.ESDTValues = append(scr.ESDTValues, res.value)
			scr.HasOperations = true
			return
		}

		if res.processed {
			return
		}
	}
}

// PrepareLogsForDB will prepare logs for database
func (lep *logsAndEventsProcessor) PrepareLogsForDB(
	logsAndEvents map[string]coreData.LogHandler,
	timestamp uint64,
) []*data.Logs {
	logs := make([]*data.Logs, 0, len(logsAndEvents))

	for txHash, txLog := range logsAndEvents {
		if check.IfNil(txLog) {
			continue
		}

		logs = append(logs, lep.prepareLogsForDB(txHash, txLog, timestamp))
	}

	return logs
}

func (lep *logsAndEventsProcessor) prepareLogsForDB(
	id string,
	logHandler coreData.LogHandler,
	timestamp uint64,
) *data.Logs {
	events := logHandler.GetLogEvents()
	logsDB := &data.Logs{
		ID:        hex.EncodeToString([]byte(id)),
		Address:   lep.pubKeyConverter.Encode(logHandler.GetAddress()),
		Timestamp: time.Duration(timestamp),
		Events:    make([]*data.Event, 0, len(events)),
	}

	for _, event := range events {
		if check.IfNil(event) {
			continue
		}

		logsDB.Events = append(logsDB.Events, &data.Event{
			Address:    lep.pubKeyConverter.Encode(event.GetAddress()),
			Identifier: string(event.GetIdentifier()),
			Topics:     event.GetTopics(),
			Data:       event.GetData(),
		})
	}

	return logsDB
}
