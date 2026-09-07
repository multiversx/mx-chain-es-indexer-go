package dataindexer

const (
	// IndexSuffix is the suffix for the Elasticsearch indexes
	IndexSuffix = "000001"
	// BlockIndex is the Elasticsearch index for the blocks
	BlockIndex = "blocks"
	// MiniblocksIndex is the Elasticsearch index for the miniblocks
	MiniblocksIndex = "miniblocks"
	// TransactionsIndex is the Elasticsearch index for the transactions
	TransactionsIndex = "transactions"
	// ValidatorsIndex is the Elasticsearch index for the validators information
	ValidatorsIndex = "validators"
	// RoundsIndex is the Elasticsearch index for the rounds information
	RoundsIndex = "rounds"
	// RatingIndex is the Elasticsearch index for the rating information
	RatingIndex = "rating"
	// AccountsIndex is the Elasticsearch index for the accounts
	AccountsIndex = "accounts"
	// AccountsHistoryIndex is the Elasticsearch index for the accounts history information
	AccountsHistoryIndex = "accountshistory"
	// ReceiptsIndex is the Elasticsearch index for the receipts
	ReceiptsIndex = "receipts"
	// ScResultsIndex is the Elasticsearch index for the smart contract results
	ScResultsIndex = "scresults"
	// AccountsESDTIndex is the Elasticsearch index for the accounts with ESDT balance
	AccountsESDTIndex = "accountsesdt"
	// AccountsESDTHistoryIndex is the Elasticsearch index for the accounts history information with ESDT balance
	AccountsESDTHistoryIndex = "accountsesdthistory"
	// EpochInfoIndex is the Elasticsearch index for the epoch information
	EpochInfoIndex = "epochinfo"
	// SCDeploysIndex is the Elasticsearch index for the smart contracts deploy information
	SCDeploysIndex = "scdeploys"
	// TokensIndex is the Elasticsearch index for the ESDT tokens
	TokensIndex = "tokens"
	// TagsIndex is the Elasticsearch index for NFTs tags
	TagsIndex = "tags"
	// LogsIndex is the Elasticsearch index for logs
	LogsIndex = "logs"
	// DelegatorsIndex is the Elasticsearch index for delegators
	DelegatorsIndex = "delegators"
	// OperationsIndex is the Elasticsearch index for transactions and smart contract results
	OperationsIndex = "operations"
	// ESDTsIndex is the Elasticsearch index for esdt tokens
	ESDTsIndex = "esdts"
	// ValuesIndex is the Elasticsearch index for extra indexer information
	ValuesIndex = "values"
	// EventsIndex is the Elasticsearch index for log events
	EventsIndex = "events"
	// ExecutionResultsIndex is the Elasticsearch index for execution results
	ExecutionResultsIndex = "executionresults"
)
