package logsevents

import (
	"encoding/hex"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
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
func NewLogsAndEventsProcessor(args *ArgsLogsAndEventsProcessor) (*logsAndEventsProcessor, error) {
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

func checkArgsLogsAndEventsProcessor(args *ArgsLogsAndEventsProcessor) error {
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

func createEventsProcessors(args *ArgsLogsAndEventsProcessor) []eventsProcessor {
	nftsProc := newNFTsProcessor(args.PubKeyConverter, args.Marshalizer)
	fungibleProc := newFungibleESDTProcessor(args.PubKeyConverter)
	scDeploysProc := newSCDeploysProcessor(args.PubKeyConverter)
	informativeProc := newInformativeLogsProcessor()
	updateNFTProc := newNFTsPropertiesProcessor(args.PubKeyConverter)
	esdtPropProc := newEsdtPropertiesProcessor(args.PubKeyConverter)

	eventsProcs := []eventsProcessor{
		fungibleProc,
		nftsProc,
		scDeploysProc,
		informativeProc,
		updateNFTProc,
		esdtPropProc,
	}

	//if args.ShardCoordinator.SelfId() == core.MetachainShardId {
	esdtIssueProc := newESDTIssueProcessor(args.PubKeyConverter)
	eventsProcs = append(eventsProcs, esdtIssueProc)

	delegatorsProcessor := newDelegatorsProcessor(args.PubKeyConverter, args.BalanceConverter)
	eventsProcs = append(eventsProcs, delegatorsProcessor)
	//}

	return eventsProcs
}

// ExtractDataFromLogs will extract data from the provided logs and events and put in altered addresses
func (lep *logsAndEventsProcessor) ExtractDataFromLogs(
	logsAndEvents []*coreData.LogData,
	preparedResults *data.PreparedResults,
	timestamp uint64,
) *data.PreparedLogsResults {
	lep.logsData = newLogsData(timestamp, preparedResults.AlteredAccts, preparedResults.Transactions, preparedResults.ScResults)

	for _, txLog := range logsAndEvents {
		if txLog == nil || check.IfNil(txLog.LogHandler) {
			continue
		}

		events := txLog.LogHandler.GetLogEvents()
		lep.processEvents(txLog.TxHash, txLog.LogHandler.GetAddress(), events)
	}

	return &data.PreparedLogsResults{
		Tokens:                  lep.logsData.tokens,
		ScDeploys:               lep.logsData.scDeploys,
		TokensInfo:              lep.logsData.tokensInfo,
		TokensSupply:            lep.logsData.tokensSupply,
		Delegators:              lep.logsData.delegators,
		NFTsDataUpdates:         lep.logsData.nftsDataUpdates,
		TokenRolesAndProperties: lep.logsData.tokenRolesAndProperties,
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
			event:                   event,
			txHashHexEncoded:        logHashHexEncoded,
			logAddress:              logAddress,
			accounts:                lep.logsData.accounts,
			tokens:                  lep.logsData.tokens,
			tokensSupply:            lep.logsData.tokensSupply,
			timestamp:               lep.logsData.timestamp,
			scDeploys:               lep.logsData.scDeploys,
			txs:                     lep.logsData.txsMap,
			tokenRolesAndProperties: lep.logsData.tokenRolesAndProperties,
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

		isEmptyIdentifier := res.identifier == ""
		if isEmptyIdentifier && res.processed {
			return
		}

		tx, ok := lep.logsData.txsMap[logHashHexEncoded]
		if ok && !isEmptyIdentifier {
			tx.HasOperations = true
			continue
		}
		scr, ok := lep.logsData.scrsMap[logHashHexEncoded]
		if ok && !isEmptyIdentifier {
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
	logsAndEvents []*coreData.LogData,
	timestamp uint64,
) []*data.Logs {
	logs := make([]*data.Logs, 0, len(logsAndEvents))

	for _, txLog := range logsAndEvents {
		if txLog == nil || check.IfNil(txLog.LogHandler) {
			continue
		}

		logs = append(logs, lep.prepareLogsForDB(txLog.TxHash, txLog.LogHandler, timestamp))
	}

	return logs
}

func (lep *logsAndEventsProcessor) prepareLogsForDB(
	id string,
	logHandler coreData.LogHandler,
	timestamp uint64,
) *data.Logs {
	encodedID := hex.EncodeToString([]byte(id))
	originalTxHash := ""
	scr, ok := lep.logsData.scrsMap[encodedID]
	if ok {
		originalTxHash = scr.OriginalTxHash
	}

	events := logHandler.GetLogEvents()
	logsDB := &data.Logs{
		ID:             encodedID,
		OriginalTxHash: originalTxHash,
		Address:        lep.pubKeyConverter.Encode(logHandler.GetAddress()),
		Timestamp:      time.Duration(timestamp),
		Events:         make([]*data.Event, 0, len(events)),
	}

	for idx, event := range events {
		if check.IfNil(event) {
			continue
		}

		logsDB.Events = append(logsDB.Events, &data.Event{
			Address:    lep.pubKeyConverter.Encode(event.GetAddress()),
			Identifier: string(event.GetIdentifier()),
			Topics:     event.GetTopics(),
			Data:       event.GetData(),
			Order:      idx,
		})
	}

	return logsDB
}
