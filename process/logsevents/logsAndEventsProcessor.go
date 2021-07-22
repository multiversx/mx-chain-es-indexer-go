package logsevents

import (
	"encoding/hex"

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process/tags"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	nodeData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

type logsAndEventsProcessor struct {
	pubKeyConverter  core.PubkeyConverter
	eventsProcessors []eventsProcessor
}

// NewLogsAndEventsProcessor will create a new instance for the logsAndEventsProcessor
func NewLogsAndEventsProcessor(
	shardCoordinator elasticIndexer.Coordinator,
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

	return &logsAndEventsProcessor{
		pubKeyConverter: pubKeyConverter,
		eventsProcessors: []eventsProcessor{
			fungibleProc, nftsProc,
		},
	}, nil
}

// ExtractDataFromLogsAndPutInAltered will extract data from the provided logs and events and put in altered addresses
func (lep *logsAndEventsProcessor) ExtractDataFromLogsAndPutInAltered(
	logsAndEvents map[string]nodeData.LogHandler,
	preparedResults *data.PreparedResults,
	timestamp uint64,
) (data.TokensHandler, tags.CountTags) {
	txsMap := converters.ConvertTxsSliceIntoMap(preparedResults.Transactions)
	scrsMap := converters.ConvertScrsSliceIntoMap(preparedResults.ScResults)

	tagsCount := tags.NewTagsCount()
	tokens := data.NewTokensInfo()
	for logHash, log := range logsAndEvents {
		if check.IfNil(log) {
			continue
		}

		events := log.GetLogEvents()
		lep.processEvents(logHash, timestamp, events, tokens, tagsCount, preparedResults.AlteredAccts, txsMap, scrsMap)
	}

	return tokens, tagsCount
}

func (lep *logsAndEventsProcessor) processEvents(
	logHash string,
	timestamp uint64,
	events []nodeData.EventHandler,
	tokens data.TokensHandler,
	tagsCount tags.CountTags,
	accounts data.AlteredAccountsHandler,
	txsMap map[string]*data.Transaction,
	scrsMap map[string]*data.ScResult,
) {
	for _, event := range events {
		if check.IfNil(event) {
			continue
		}

		lep.processEvent(logHash, timestamp, event, tokens, tagsCount, accounts, txsMap, scrsMap)
	}
}

func (lep *logsAndEventsProcessor) processEvent(
	logHash string,
	timestamp uint64,
	events nodeData.EventHandler,
	tokens data.TokensHandler,
	tagsCount tags.CountTags,
	accounts data.AlteredAccountsHandler,
	txsMap map[string]*data.Transaction,
	scrsMap map[string]*data.ScResult,
) {
	logHashHexEncoded := hex.EncodeToString([]byte(logHash))
	for _, proc := range lep.eventsProcessors {
		identifier, processed := proc.processEvent(&argsProcessEvent{
			event:     events,
			accounts:  accounts,
			tokens:    tokens,
			tagsCount: tagsCount,
			timestamp: timestamp,
		})
		if identifier == "" {
			continue
		}

		tx, ok := txsMap[logHashHexEncoded]
		if ok {
			tx.EsdtTokenIdentifier = identifier
			continue
		}

		scr, ok := scrsMap[logHashHexEncoded]
		if ok {
			scr.EsdtTokenIdentifier = identifier
			return
		}

		if processed {
			return
		}
	}
}

// PrepareLogsForDB will prepare logs for database
func (lep *logsAndEventsProcessor) PrepareLogsForDB(logsAndEvents map[string]nodeData.LogHandler) []*data.Logs {
	logs := make([]*data.Logs, 0, len(logsAndEvents))

	for txHash, log := range logsAndEvents {
		if check.IfNil(log) {
			continue
		}

		logs = append(logs, lep.prepareLogsForDB(txHash, log))
	}

	return logs
}

func (lep *logsAndEventsProcessor) prepareLogsForDB(id string, logHandler nodeData.LogHandler) *data.Logs {
	events := logHandler.GetLogEvents()
	logsDB := &data.Logs{
		ID:      hex.EncodeToString([]byte(id)),
		Address: lep.pubKeyConverter.Encode(logHandler.GetAddress()),
		Events:  make([]*data.Event, 0, len(events)),
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
