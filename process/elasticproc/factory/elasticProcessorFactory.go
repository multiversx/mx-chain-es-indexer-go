package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/accounts"
	blockProc "github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/block"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/logsevents"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/miniblocks"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/operations"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/statistics"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/templatesAndPolicies"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/transactions"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/validators"
)

// ElasticConfig holds the elastic search settings
type ElasticConfig struct {
	Enabled  bool
	Url      string
	UserName string
	Password string
}

// ArgElasticProcessorFactory is struct that is used to store all components that are needed to create an elastic processor factory
type ArgElasticProcessorFactory struct {
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	DBClient                 elasticproc.DatabaseClientHandler
	EnabledIndexes           []string
	Version                  string
	Denomination             int
	BulkRequestMaxSize       int
	UseKibana                bool
	ImportDB                 bool
	TxHashExtractor          transactions.TxHashExtractor
	RewardTxData             transactions.RewardTxDataHandler
	IndexTokensHandler       elasticproc.IndexTokensHandler
}

// CreateElasticProcessor will create a new instance of ElasticProcessor
func CreateElasticProcessor(arguments ArgElasticProcessorFactory) (dataindexer.ElasticProcessor, error) {
	templatesAndPoliciesReader := templatesAndPolicies.CreateTemplatesAndPoliciesReader(arguments.UseKibana)
	indexTemplates, indexPolicies, err := templatesAndPoliciesReader.GetElasticTemplatesAndPolicies()
	if err != nil {
		return nil, err
	}
	extraMappings, err := templatesAndPoliciesReader.GetExtraMappings()
	if err != nil {
		return nil, err
	}

	enabledIndexesMap := make(map[string]struct{})
	for _, index := range arguments.EnabledIndexes {
		enabledIndexesMap[index] = struct{}{}
	}
	if len(enabledIndexesMap) == 0 {
		return nil, dataindexer.ErrEmptyEnabledIndexes
	}

	balanceConverter, err := converters.NewBalanceConverter(arguments.Denomination)
	if err != nil {
		return nil, err
	}

	accountsProc, err := accounts.NewAccountsProcessor(
		arguments.AddressPubkeyConverter,
		balanceConverter,
	)
	if err != nil {
		return nil, err
	}

	blockProcHandler, err := blockProc.NewBlockProcessor(arguments.Hasher, arguments.Marshalizer)
	if err != nil {
		return nil, err
	}

	miniblocksProc, err := miniblocks.NewMiniblocksProcessor(arguments.Hasher, arguments.Marshalizer)
	if err != nil {
		return nil, err
	}
	validatorsProc, err := validators.NewValidatorsProcessor(arguments.ValidatorPubkeyConverter, arguments.BulkRequestMaxSize)
	if err != nil {
		return nil, err
	}

	generalInfoProc := statistics.NewStatisticsProcessor()

	argsTxsProc := &transactions.ArgsTransactionProcessor{
		AddressPubkeyConverter: arguments.AddressPubkeyConverter,
		Hasher:                 arguments.Hasher,
		Marshalizer:            arguments.Marshalizer,
		BalanceConverter:       balanceConverter,
		TxHashExtractor:        arguments.TxHashExtractor,
		RewardTxData:           arguments.RewardTxData,
	}
	txsProc, err := transactions.NewTransactionsProcessor(argsTxsProc)
	if err != nil {
		return nil, err
	}

	argsLogsAndEventsProc := logsevents.ArgsLogsAndEventsProcessor{
		PubKeyConverter:  arguments.AddressPubkeyConverter,
		Marshalizer:      arguments.Marshalizer,
		BalanceConverter: balanceConverter,
		Hasher:           arguments.Hasher,
	}
	logsAndEventsProc, err := logsevents.NewLogsAndEventsProcessor(argsLogsAndEventsProc)
	if err != nil {
		return nil, err
	}

	operationsProc, err := operations.NewOperationsProcessor()
	if err != nil {
		return nil, err
	}

	args := &elasticproc.ArgElasticProcessor{
		BulkRequestMaxSize: arguments.BulkRequestMaxSize,
		TransactionsProc:   txsProc,
		AccountsProc:       accountsProc,
		BlockProc:          blockProcHandler,
		MiniblocksProc:     miniblocksProc,
		ValidatorsProc:     validatorsProc,
		StatisticsProc:     generalInfoProc,
		LogsAndEventsProc:  logsAndEventsProc,
		DBClient:           arguments.DBClient,
		EnabledIndexes:     enabledIndexesMap,
		UseKibana:          arguments.UseKibana,
		IndexTemplates:     indexTemplates,
		IndexPolicies:      indexPolicies,
		ExtraMappings:      extraMappings,
		OperationsProc:     operationsProc,
		ImportDB:           arguments.ImportDB,
		Version:            arguments.Version,
		IndexTokensHandler: arguments.IndexTokensHandler,
	}

	return elasticproc.NewElasticProcessor(args)
}
