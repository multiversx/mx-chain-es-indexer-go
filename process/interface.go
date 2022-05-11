package process

import (
	"bytes"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// PostgresClientHandler defines the actions that a component that handles requests should do
type PostgresClientHandler interface {
	CreateTables() error
	CreateTable(entity interface{}) error
	CreateRawTable(sql string) error
	AutoMigrateTables(tables ...interface{}) error
	Insert(entity interface{}) error
	InsertBlock(block *data.Block) error
	Raw(sql string, values ...interface{}) error
	Exec(sql string, values ...interface{}) error
	InsertEpochStartInfo(block *data.Block) error
	InsertValidatorsRating(id string, ratingInfo *data.ValidatorRatingInfo) error
	InsertValidatorsPubKeys(id string, pubKeys *data.ValidatorsPublicKeys) error
	InsertEpochInfo(block *block.MetaBlock) error
	InsertAccount(account *data.AccountInfo) error
	InsertAccountESDT(id string, account *data.AccountInfo) error

	InsertAccountHistory(account *data.AccountBalanceHistory) error
	InsertAccountESDTHistory(account *data.AccountBalanceHistory) error
	IsInterfaceNil() bool
}

// DatabaseClientHandler defines the actions that a component that handles requests should do
type DatabaseClientHandler interface {
	DoRequest(req *esapi.IndexRequest) error
	DoBulkRequest(buff *bytes.Buffer, index string) error
	DoBulkRemove(index string, hashes []string) error
	DoMultiGet(ids []string, index string, withSource bool, res interface{}) error

	CheckAndCreateIndex(index string) error
	CheckAndCreateAlias(alias string, index string) error
	CheckAndCreateTemplate(templateName string, template *bytes.Buffer) error
	CheckAndCreatePolicy(policyName string, policy *bytes.Buffer) error

	IsInterfaceNil() bool
}

// DBAccountHandler defines the actions that an accounts handler should do
type DBAccountHandler interface {
	GetAccounts(alteredAccounts data.AlteredAccountsHandler) ([]*data.Account, []*data.AccountESDT)
	PrepareRegularAccountsMap(accounts []*data.Account) map[string]*data.AccountInfo
	PrepareAccountsMapESDT(accounts []*data.AccountESDT) map[string]*data.AccountInfo
	PrepareAccountsHistory(timestamp uint64, accounts map[string]*data.AccountInfo) map[string]*data.AccountBalanceHistory
	PutTokenMedataDataInTokens(tokensData []*data.TokenInfo)

	SerializeAccountsHistory(accounts map[string]*data.AccountBalanceHistory) ([]*bytes.Buffer, error)
	SerializeAccounts(accounts map[string]*data.AccountInfo, areESDTAccounts bool) ([]*bytes.Buffer, error)
	SerializeNFTCreateInfo(tokensInfo []*data.TokenInfo) ([]*bytes.Buffer, error)
}

// DBBlockHandler defines the actions that a block handler should do
type DBBlockHandler interface {
	PrepareBlockForDB(
		header coreData.HeaderHandler,
		signersIndexes []uint64,
		body *block.Body,
		notarizedHeadersHashes []string,
		gasConsumptionData indexer.HeaderGasConsumption,
		sizeTxs int,
	) (*data.Block, error)
	ComputeHeaderHash(header coreData.HeaderHandler) ([]byte, error)

	SerializeEpochInfoData(header coreData.HeaderHandler) (*bytes.Buffer, error)
	SerializeBlock(elasticBlock *data.Block) (*bytes.Buffer, error)
}

// DBTransactionsHandler defines the actions that a transactions handler should do
type DBTransactionsHandler interface {
	PrepareTransactionsForDatabase(
		body *block.Body,
		header coreData.HeaderHandler,
		pool *indexer.Pool,
	) *data.PreparedResults
	GetRewardsTxsHashesHexEncoded(header coreData.HeaderHandler, body *block.Body) []string

	SerializeReceipts(receipts []*data.Receipt) ([]*bytes.Buffer, error)
	SerializeTransactions(transactions []*data.Transaction, txHashStatus map[string]string, selfShardID uint32) ([]*bytes.Buffer, error)
	SerializeTransactionWithRefund(txs map[string]*data.Transaction, txHashRefund map[string]*data.RefundData) ([]*bytes.Buffer, error)
	SerializeScResults(scResults []*data.ScResult) ([]*bytes.Buffer, error)
}

// DBMiniblocksHandler defines the actions that a miniblocks handler should do
type DBMiniblocksHandler interface {
	PrepareDBMiniblocks(header coreData.HeaderHandler, body *block.Body) []*data.Miniblock
	GetMiniblocksHashesHexEncoded(header coreData.HeaderHandler, body *block.Body) []string

	SerializeBulkMiniBlocks(bulkMbs []*data.Miniblock, mbsInDB map[string]bool) *bytes.Buffer
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
	ValidatorsRatingToPostgres(
		postgresClient PostgresClientHandler,
		index string,
		validatorsRatingInfo []*data.ValidatorRatingInfo,
	) error
}

// DBLogsAndEventsHandler defines the actions that a logs and events handler should do
type DBLogsAndEventsHandler interface {
	PrepareLogsForDB(logsAndEvents []*coreData.LogData, timestamp uint64) []*data.Logs
	ExtractDataFromLogs(
		logsAndEvents []*coreData.LogData,
		preparedResults *data.PreparedResults,
		timestamp uint64,
	) *data.PreparedLogsResults

	SerializeLogs(logs []*data.Logs) ([]*bytes.Buffer, error)
	SerializeSCDeploys(map[string]*data.ScDeployInfo) ([]*bytes.Buffer, error)
	SerializeTokens(tokens []*data.TokenInfo) ([]*bytes.Buffer, error)
	SerializeDelegators(delegators map[string]*data.Delegator) ([]*bytes.Buffer, error)
}
