package logsevents

import (
	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-logger/check"
	"github.com/ElrondNetwork/elrond-go/core"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type logsAndEventsProcessor struct {
	nftsProc *nftsProcessor
}

// NewLogsAndEventsProcessor will create a new instance for the logsAndEventsProcessor
func NewLogsAndEventsProcessor(
	shardCoordinator sharding.Coordinator,
	pubKeyConverter core.PubkeyConverter,
) (*logsAndEventsProcessor, error) {
	if check.IfNil(shardCoordinator) {
		return nil, elasticIndexer.ErrNilShardCoordinator
	}
	if check.IfNil(pubKeyConverter) {
		return nil, elasticIndexer.ErrNilPubkeyConverter
	}

	return &logsAndEventsProcessor{
		nftsProc: newNFTsProcessor(shardCoordinator, pubKeyConverter),
	}, nil
}

// ExtractDataFromLogsAndPutInAltered will extract information from provided logs and events and put in altered address
func (lep *logsAndEventsProcessor) ExtractDataFromLogsAndPutInAltered(
	logsAndEvents map[string]nodeData.LogHandler,
	accounts data.AlteredAccountsHandler,
) {
	lep.nftsProc.processLogAndEventsNFTs(logsAndEvents, accounts)
}

// PrepareLogsForDB will prepare logs for database
func (lep *logsAndEventsProcessor) PrepareLogsForDB(logsAndEvents map[string]nodeData.LogHandler) []*data.Logs {
	return nil
}
