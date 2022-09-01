package elasticproc

import (
	"bytes"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process/elasticproc/tokeninfo"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// DatabaseClientHandler defines the actions that a component that handles requests should do
type DatabaseClientHandler interface {
	DoRequest(req *esapi.IndexRequest) error
	DoBulkRequest(buff *bytes.Buffer, index string) error
	DoQueryRemove(index string, buff *bytes.Buffer) error
	DoMultiGet(ids []string, index string, withSource bool, res interface{}) error
	DoScrollRequest(index string, body []byte, withSource bool, handlerFunc func(responseBytes []byte) error) error
	DoCountRequest(index string, body []byte) (uint64, error)

	CheckAndCreateIndex(index string) error
	CheckAndCreateAlias(alias string, index string) error
	CheckAndCreateTemplate(templateName string, template *bytes.Buffer) error
	CheckAndCreatePolicy(policyName string, policy *bytes.Buffer) error

	IsInterfaceNil() bool
}

// DBAccountHandler defines the actions that an accounts' handler should do
type DBAccountHandler interface {
	GetAccounts(alteredAccounts data.AlteredAccountsHandler, coreAlteredAccounts map[string]*outport.AlteredAccount) ([]*data.Account, []*data.AccountESDT)
	PrepareRegularAccountsMap(timestamp uint64, accounts []*data.Account) map[string]*data.AccountInfo
	PrepareAccountsMapESDT(timestamp uint64, accounts []*data.AccountESDT, tagsCount data.CountTags) (map[string]*data.AccountInfo, data.TokensHandler)
	PrepareAccountsHistory(timestamp uint64, accounts map[string]*data.AccountInfo) map[string]*data.AccountBalanceHistory
	PutTokenMedataDataInTokens(tokensData []*data.TokenInfo, coreAlteredAccounts map[string]*outport.AlteredAccount)

	SerializeAccountsHistory(accounts map[string]*data.AccountBalanceHistory, buffSlice *data.BufferSlice, index string) error
	SerializeAccounts(accounts map[string]*data.AccountInfo, buffSlice *data.BufferSlice, index string) error
	SerializeAccountsESDT(accounts map[string]*data.AccountInfo, updateNFTData []*data.NFTDataUpdate, buffSlice *data.BufferSlice, index string) error
	SerializeNFTCreateInfo(tokensInfo []*data.TokenInfo, buffSlice *data.BufferSlice, index string) error
	SerializeTypeForProvidedIDs(ids []string, tokenType string, buffSlice *data.BufferSlice, index string) error
}

// DBBlockHandler defines the actions that a block handler should do
type DBBlockHandler interface {
	PrepareBlockForDB(
		headerHash []byte,
		header coreData.HeaderHandler,
		signersIndexes []uint64,
		body *block.Body,
		notarizedHeadersHashes []string,
		gasConsumptionData outport.HeaderGasConsumption,
		sizeTxs int,
	) (*data.Block, error)
	ComputeHeaderHash(header coreData.HeaderHandler) ([]byte, error)

	SerializeEpochInfoData(header coreData.HeaderHandler, buffSlice *data.BufferSlice, index string) error
	SerializeBlock(elasticBlock *data.Block, buffSlice *data.BufferSlice, index string) error
}

// DBTransactionsHandler defines the actions that a transactions handler should do
type DBTransactionsHandler interface {
	PrepareTransactionsForDatabase(
		body *block.Body,
		header coreData.HeaderHandler,
		pool *outport.Pool,
	) *data.PreparedResults
	GetHexEncodedHashesForRemove(header coreData.HeaderHandler, body *block.Body) ([]string, []string)

	SerializeReceipts(receipts []*data.Receipt, buffSlice *data.BufferSlice, index string) error
	SerializeTransactions(transactions []*data.Transaction, txHashStatus map[string]string, selfShardID uint32, buffSlice *data.BufferSlice, index string) error
	SerializeTransactionsFeeData(txHashRefund map[string]*data.FeeData, buffSlice *data.BufferSlice, index string) error
	SerializeScResults(scResults []*data.ScResult, buffSlice *data.BufferSlice, index string) error
}

// DBMiniblocksHandler defines the actions that a miniblocks handler should do
type DBMiniblocksHandler interface {
	PrepareDBMiniblocks(header coreData.HeaderHandler, body *block.Body) []*data.Miniblock
	GetMiniblocksHashesHexEncoded(header coreData.HeaderHandler, body *block.Body) []string

	SerializeBulkMiniBlocks(bulkMbs []*data.Miniblock, mbsInDB map[string]bool, buffSlice *data.BufferSlice, index string)
}

// DBStatisticsHandler defines the actions that a database statistics handler should do
type DBStatisticsHandler interface {
	SerializeRoundsInfo(roundsInfo []*data.RoundInfo) *bytes.Buffer
}

// DBValidatorsHandler defines the actions that a validators handler should do
type DBValidatorsHandler interface {
	PrepareValidatorsPublicKeys(shardValidatorsPubKeys [][]byte) *data.ValidatorsPublicKeys
	SerializeValidatorsPubKeys(validatorsPubKeys *data.ValidatorsPublicKeys) (*bytes.Buffer, error)
	SerializeValidatorsRating(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) ([]*bytes.Buffer, error)
}

// DBLogsAndEventsHandler defines the actions that a logs and events handler should do
type DBLogsAndEventsHandler interface {
	PrepareLogsForDB(logsAndEvents []*coreData.LogData, timestamp uint64) []*data.Logs
	ExtractDataFromLogs(
		logsAndEvents []*coreData.LogData,
		preparedResults *data.PreparedResults,
		timestamp uint64,
	) *data.PreparedLogsResults

	SerializeLogs(logs []*data.Logs, buffSlice *data.BufferSlice, index string) error
	SerializeSCDeploys(deploysInfo map[string]*data.ScDeployInfo, buffSlice *data.BufferSlice, index string) error
	SerializeTokens(tokens []*data.TokenInfo, updateNFTData []*data.NFTDataUpdate, buffSlice *data.BufferSlice, index string) error
	SerializeDelegators(delegators map[string]*data.Delegator, buffSlice *data.BufferSlice, index string) error
	SerializeSupplyData(tokensSupply data.TokensHandler, buffSlice *data.BufferSlice, index string) error
	SerializeRolesData(
		tokenRolesAndProperties *tokeninfo.TokenRolesAndProperties,
		buffSlice *data.BufferSlice,
		index string,
	) error
}

// OperationsHandler defines the actions that an operations' handler should do
type OperationsHandler interface {
	ProcessTransactionsAndSCRs(txs []*data.Transaction, scrs []*data.ScResult) ([]*data.Transaction, []*data.ScResult)
	SerializeSCRs(scrs []*data.ScResult, buffSlice *data.BufferSlice, index string) error
}
