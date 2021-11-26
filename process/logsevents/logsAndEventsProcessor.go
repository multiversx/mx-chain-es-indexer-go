package logsevents

import (
	"encoding/hex"
	"time"

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

type logsAndEventsProcessor struct {
	hasher           hashing.Hasher
	pubKeyConverter  core.PubkeyConverter
	eventsProcessors []eventsProcessor

	logsData *logsData
}

// NewLogsAndEventsProcessor will create a new instance for the logsAndEventsProcessor
func NewLogsAndEventsProcessor(
	shardCoordinator elasticIndexer.ShardCoordinator,
	pubKeyConverter core.PubkeyConverter,
	marshalizer marshal.Marshalizer,
	balanceConverter elasticIndexer.BalanceConverter,
	hasher hashing.Hasher,
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
	if check.IfNil(balanceConverter) {
		return nil, elasticIndexer.ErrNilBalanceConverter
	}
	if check.IfNil(hasher) {
		return nil, elasticIndexer.ErrNilHasher
	}

	eventsProcessors := createEventsProcessors(shardCoordinator, pubKeyConverter, marshalizer, balanceConverter)

	return &logsAndEventsProcessor{
		pubKeyConverter:  pubKeyConverter,
		eventsProcessors: eventsProcessors,
		hasher:           hasher,
	}, nil
}

func createEventsProcessors(
	shardCoordinator elasticIndexer.ShardCoordinator,
	pubKeyConverter core.PubkeyConverter,
	marshalizer marshal.Marshalizer,
	balanceConverter elasticIndexer.BalanceConverter,
) []eventsProcessor {
	nftsProc := newNFTsProcessor(shardCoordinator, pubKeyConverter, marshalizer)
	fungibleProc := newFungibleESDTProcessor(pubKeyConverter, shardCoordinator)
	scDeploysProc := newSCDeploysProcessor(pubKeyConverter)

	eventsProcs := []eventsProcessor{
		fungibleProc,
		nftsProc,
		scDeploysProc,
	}

	if shardCoordinator.SelfId() == core.MetachainShardId {
		esdtIssueProc := newESDTIssueProcessor(pubKeyConverter)
		eventsProcs = append(eventsProcs, esdtIssueProc)

		delegatorsProcessor := newDelegatorsProcessor(pubKeyConverter, balanceConverter)
		eventsProcs = append(eventsProcs, delegatorsProcessor)
	}

	return eventsProcs
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
		lep.processEvents(logHash, txLog.GetAddress(), events)
	}

	return &data.PreparedLogsResults{
		Tokens:          lep.logsData.tokens,
		ScDeploys:       lep.logsData.scDeploys,
		TagsCount:       lep.logsData.tagsCount,
		PendingBalances: lep.logsData.pendingBalances.getAll(),
		TokensInfo:      lep.logsData.tokensInfo,
		Delegators:      lep.logsData.delegators,
	}
}

func (lep *logsAndEventsProcessor) processEvents(logHash string, logAddress []byte, events []coreData.EventHandler) {
	for _, event := range events {
		if check.IfNil(event) {
			continue
		}

		lep.processEvent(logHash, logAddress, event)
	}
}

func (lep *logsAndEventsProcessor) processEvent(logHash string, logAddress []byte, event coreData.EventHandler) {
	logHashHexEncoded := hex.EncodeToString([]byte(logHash))
	for _, proc := range lep.eventsProcessors {
		res := proc.processEvent(&argsProcessEvent{
			event:            event,
			txHashHexEncoded: logHashHexEncoded,
			logAddress:       logAddress,
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
		if res.delegator != nil {
			lep.logsData.delegators[res.delegator.Address] = res.delegator
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
