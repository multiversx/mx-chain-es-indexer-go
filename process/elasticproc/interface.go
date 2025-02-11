package elasticproc

import (
	"bytes"
	"context"

	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/tokeninfo"
)

// MainChainDatabaseClientHandler defines the actions that sovereign database client handler should do
type MainChainDatabaseClientHandler interface {
	DatabaseClientHandler
	IsEnabled() bool
	IsInterfaceNil() bool
}

// DatabaseClientHandler defines the actions that a component that handles requests should do
type DatabaseClientHandler interface {
	DoBulkRequest(ctx context.Context, buff *bytes.Buffer, index string) error
	DoQueryRemove(ctx context.Context, index string, buff *bytes.Buffer) error
	DoMultiGet(ctx context.Context, ids []string, index string, withSource bool, res interface{}) error
	DoScrollRequest(ctx context.Context, index string, body []byte, withSource bool, handlerFunc func(responseBytes []byte) error) error
	DoCountRequest(ctx context.Context, index string, body []byte) (uint64, error)
	UpdateByQuery(ctx context.Context, index string, buff *bytes.Buffer) error

	PutMappings(indexName string, mappings *bytes.Buffer) error
	CheckAndCreateIndex(index string) error
	CheckAndCreateAlias(alias string, index string) error
	CheckAndCreateTemplate(templateName string, template *bytes.Buffer) error
	CheckAndCreatePolicy(policyName string, policy *bytes.Buffer) error

	IsInterfaceNil() bool
}

// DBAccountHandler defines the actions that an accounts' handler should do
type DBAccountHandler interface {
	GetAccounts(coreAlteredAccounts map[string]*alteredAccount.AlteredAccount) ([]*data.Account, []*data.AccountESDT)
	PrepareRegularAccountsMap(timestamp uint64, accounts []*data.Account, shardID uint32) map[string]*data.AccountInfo
	PrepareAccountsMapESDT(timestamp uint64, accounts []*data.AccountESDT, tagsCount data.CountTags, shardID uint32) (map[string]*data.AccountInfo, data.TokensHandler)
	PrepareAccountsHistory(timestamp uint64, accounts map[string]*data.AccountInfo, shardID uint32) map[string]*data.AccountBalanceHistory
	PutTokenMedataDataInTokens(tokensData []*data.TokenInfo, coreAlteredAccounts map[string]*alteredAccount.AlteredAccount)

	SerializeAccountsHistory(accounts map[string]*data.AccountBalanceHistory, buffSlice *data.BufferSlice, index string) error
	SerializeAccounts(accounts map[string]*data.AccountInfo, buffSlice *data.BufferSlice, index string) error
	SerializeAccountsESDT(accounts map[string]*data.AccountInfo, updateNFTData []*data.NFTDataUpdate, buffSlice *data.BufferSlice, index string) error
	SerializeNFTCreateInfo(tokensInfo []*data.TokenInfo, buffSlice *data.BufferSlice, index string) error
	SerializeTypeForProvidedIDs(ids []string, tokenType string, buffSlice *data.BufferSlice, index string) error
}

// DBBlockHandler defines the actions that a block handler should do
type DBBlockHandler interface {
	PrepareBlockForDB(obh *outport.OutportBlockWithHeader) (*data.Block, error)
	ComputeHeaderHash(header coreData.HeaderHandler) ([]byte, error)

	SerializeEpochInfoData(header coreData.HeaderHandler, buffSlice *data.BufferSlice, index string) error
	SerializeBlock(elasticBlock *data.Block, buffSlice *data.BufferSlice, index string) error
}

// DBTransactionsHandler defines the actions that a transactions handler should do
type DBTransactionsHandler interface {
	PrepareTransactionsForDatabase(
		miniBlocks []*block.MiniBlock,
		header coreData.HeaderHandler,
		pool *outport.TransactionPool,
		isImportDB bool,
		numOfShards uint32,
	) *data.PreparedResults
	GetHexEncodedHashesForRemove(header coreData.HeaderHandler, body *block.Body) ([]string, []string)

	SerializeReceipts(receipts []*data.Receipt, buffSlice *data.BufferSlice, index string) error
	SerializeTransactions(transactions []*data.Transaction, txHashStatusInfo map[string]*outport.StatusInfo, selfShardID uint32, buffSlice *data.BufferSlice, index string) error
	SerializeTransactionsFeeData(txHashRefund map[string]*data.FeeData, buffSlice *data.BufferSlice, index string) error
	SerializeScResults(scResults []*data.ScResult, buffSlice *data.BufferSlice, index string) error
}

// DBMiniblocksHandler defines the actions that a miniblocks handler should do
type DBMiniblocksHandler interface {
	PrepareDBMiniblocks(header coreData.HeaderHandler, miniBlocks []*block.MiniBlock) []*data.Miniblock
	GetMiniblocksHashesHexEncoded(header coreData.HeaderHandler, body *block.Body) []string

	SerializeBulkMiniBlocks(bulkMbs []*data.Miniblock, buffSlice *data.BufferSlice, index string, shardID uint32)
}

// DBStatisticsHandler defines the actions that a database statistics handler should do
type DBStatisticsHandler interface {
	SerializeRoundsInfo(rounds *outport.RoundsInfo) *bytes.Buffer
}

// DBValidatorsHandler defines the actions that a validators handler should do
type DBValidatorsHandler interface {
	PrepareAnSerializeValidatorsPubKeys(validatorsPubKeys *outport.ValidatorsPubKeys) ([]*bytes.Buffer, error)
	SerializeValidatorsRating(ratingData *outport.ValidatorsRating) ([]*bytes.Buffer, error)
}

// DBLogsAndEventsHandler defines the actions that a logs and events handler should do
type DBLogsAndEventsHandler interface {
	ExtractDataFromLogs(
		logsAndEvents []*outport.LogData,
		preparedResults *data.PreparedResults,
		timestamp uint64,
		shardID uint32,
		numOfShards uint32,
	) *data.PreparedLogsResults

	SerializeEvents(events []*data.LogEvent, buffSlice *data.BufferSlice, index string) error
	SerializeLogs(logs []*data.Logs, buffSlice *data.BufferSlice, index string) error
	SerializeSCDeploys(deploysInfo map[string]*data.ScDeployInfo, buffSlice *data.BufferSlice, index string) error
	SerializeChangeOwnerOperations(changeOwnerOperations map[string]*data.OwnerData, buffSlice *data.BufferSlice, index string) error
	SerializeTokens(tokens []*data.TokenInfo, updateNFTData []*data.NFTDataUpdate, buffSlice *data.BufferSlice, index string) error
	SerializeDelegators(delegators map[string]*data.Delegator, buffSlice *data.BufferSlice, index string) error
	SerializeSupplyData(tokensSupply data.TokensHandler, buffSlice *data.BufferSlice, index string) error
	SerializeRolesData(
		tokenRolesAndProperties *tokeninfo.TokenRolesAndProperties,
		buffSlice *data.BufferSlice,
		index string,
	) error
	PrepareDelegatorsQueryInCaseOfRevert(timestamp uint64) *bytes.Buffer
}

// OperationsHandler defines the actions that an operations' handler should do
type OperationsHandler interface {
	ProcessTransactionsAndSCRs(txs []*data.Transaction, scrs []*data.ScResult, isImportDB bool, shardID uint32) ([]*data.Transaction, []*data.ScResult)
	SerializeSCRs(scrs []*data.ScResult, buffSlice *data.BufferSlice, index string, shardID uint32) error
}

// IndexTokensHandler defines what index tokens handler should be able to do
type IndexTokensHandler interface {
	IndexCrossChainTokens(handler DatabaseClientHandler, scrs []*data.ScResult, buffSlice *data.BufferSlice) error
	IsInterfaceNil() bool
}
