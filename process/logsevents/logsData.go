package logsevents

import (
	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process/tags"
)

type logsData struct {
	timestamp uint64
	tokens    data.TokensHandler
	tagsCount tags.CountTags
	accounts  data.AlteredAccountsHandler
	txsMap    map[string]*data.Transaction
	scrsMap   map[string]*data.ScResult
	scDeploys map[string]*data.ScDeployInfo
}

func newLogsData(
	timestamp uint64,
	accounts data.AlteredAccountsHandler,
	txs []*data.Transaction,
	scrs []*data.ScResult,
) *logsData {
	ld := &logsData{}

	ld.txsMap = converters.ConvertTxsSliceIntoMap(txs)
	ld.scrsMap = converters.ConvertScrsSliceIntoMap(scrs)
	ld.tagsCount = tags.NewTagsCount()
	ld.tokens = data.NewTokensInfo()
	ld.accounts = accounts
	ld.timestamp = timestamp
	ld.scDeploys = make(map[string]*data.ScDeployInfo)

	return ld
}