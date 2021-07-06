package logsevents

import (
	"encoding/hex"

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process/tags"
	"github.com/ElrondNetwork/elrond-go-logger/check"
	"github.com/ElrondNetwork/elrond-go/core"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type logsAndEventsProcessor struct {
	pubKeyConverter core.PubkeyConverter
	nftsProc        *nftsProcessor
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

	return &logsAndEventsProcessor{
		pubKeyConverter: pubKeyConverter,
		nftsProc:        newNFTsProcessor(shardCoordinator, pubKeyConverter, marshalizer),
	}, nil
}

// ExtractDataFromLogsAndPutInAltered will extract data from the provided logs and events and put in altered addresses
func (lep *logsAndEventsProcessor) ExtractDataFromLogsAndPutInAltered(
	logsAndEvents map[string]nodeData.LogHandler,
	accounts data.AlteredAccountsHandler,
	timestamp uint64,
) (data.TokensHandler, tags.CountTags) {
	return lep.nftsProc.processLogAndEventsNFTs(logsAndEvents, accounts, timestamp)
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
