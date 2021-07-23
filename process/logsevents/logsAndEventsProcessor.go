package logsevents

import (
	"encoding/hex"

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process/tags"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type logsAndEventsProcessor struct {
	pubKeyConverter  core.PubkeyConverter
	eventsProcessors []eventsProcessor

	logsData *logsData
}

// NewLogsAndEventsProcessor will create a new instance for the logsAndEventsProcessor
func NewLogsAndEventsProcessor(
	shardCoordinator sharding.Coordinator,
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

	return &logsAndEventsProcessor{
		pubKeyConverter: pubKeyConverter,
		eventsProcessors: []eventsProcessor{
			fungibleProc,
			nftsProc,
			scDeploysProc,
		},
	}, nil
}

// ExtractDataFromLogsAndPutInAltered will extract data from the provided logs and events and put in altered addresses
func (lep *logsAndEventsProcessor) ExtractDataFromLogsAndPutInAltered(
	logsAndEvents map[string]nodeData.LogHandler,
	preparedResults *data.PreparedResults,
	timestamp uint64,
) (data.TokensHandler, tags.CountTags, map[string]*data.ScDeployInfo) {
	lep.logsData = newLogsData(timestamp, preparedResults.AlteredAccts, preparedResults.Transactions, preparedResults.ScResults)

	for logHash, log := range logsAndEvents {
		if check.IfNil(log) {
			continue
		}

		events := log.GetLogEvents()
		lep.processEvents(logHash, events)
	}

	return lep.logsData.tokens, lep.logsData.tagsCount, lep.logsData.scDeploys
}

func (lep *logsAndEventsProcessor) processEvents(logHash string, events []nodeData.EventHandler) {
	for _, event := range events {
		if check.IfNil(event) {
			continue
		}

		lep.processEvent(logHash, event)
	}
}

func (lep *logsAndEventsProcessor) processEvent(logHash string, events nodeData.EventHandler) {
	logHashHexEncoded := hex.EncodeToString([]byte(logHash))
	for _, proc := range lep.eventsProcessors {
		identifier, processed := proc.processEvent(&argsProcessEvent{
			event:            events,
			accounts:         lep.logsData.accounts,
			tokens:           lep.logsData.tokens,
			tagsCount:        lep.logsData.tagsCount,
			timestamp:        lep.logsData.timestamp,
			scDeploys:        lep.logsData.scDeploys,
			txHashHexEncoded: logHashHexEncoded,
		})
		isEmptyIdentifier := identifier == ""
		if isEmptyIdentifier && processed {
			return
		}

		tx, ok := lep.logsData.txsMap[logHashHexEncoded]
		if ok && !isEmptyIdentifier {
			tx.EsdtTokenIdentifier = identifier
			continue
		}

		scr, ok := lep.logsData.scrsMap[logHashHexEncoded]
		if ok && !isEmptyIdentifier {
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
