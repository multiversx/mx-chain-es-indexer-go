package factory

import (
	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	processIndexer "github.com/ElrondNetwork/elastic-indexer-go/process"
	"github.com/ElrondNetwork/elastic-indexer-go/process/accounts"
	blockProc "github.com/ElrondNetwork/elastic-indexer-go/process/block"
	"github.com/ElrondNetwork/elastic-indexer-go/process/logsevents"
	"github.com/ElrondNetwork/elastic-indexer-go/process/miniblocks"
	"github.com/ElrondNetwork/elastic-indexer-go/process/statistics"
	"github.com/ElrondNetwork/elastic-indexer-go/process/templatesAndPolicies"
	"github.com/ElrondNetwork/elastic-indexer-go/process/transactions"
	"github.com/ElrondNetwork/elastic-indexer-go/process/validators"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

// ArgElasticProcessorFactory is struct that is used to store all components that are needed to create an elastic processor factory
type ArgElasticProcessorFactory struct {
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	DBClient                 processIndexer.DatabaseClientHandler
	AccountsDB               indexer.AccountsAdapter
	ShardCoordinator         indexer.ShardCoordinator
	TransactionFeeCalculator indexer.FeesProcessorHandler
	EnabledIndexes           []string
	Denomination             int
	IsInImportDBMode         bool
	UseKibana                bool
}

// CreateElasticProcessor will create a new instance of ElasticProcessor
func CreateElasticProcessor(arguments ArgElasticProcessorFactory) (indexer.ElasticProcessor, error) {
	templatesAndPoliciesReader := templatesAndPolicies.CreateTemplatesAndPoliciesReader(arguments.UseKibana)
	indexTemplates, indexPolicies, err := templatesAndPoliciesReader.GetElasticTemplatesAndPolicies()
	if err != nil {
		return nil, err
	}

	enabledIndexesMap := make(map[string]struct{})
	for _, index := range arguments.EnabledIndexes {
		enabledIndexesMap[index] = struct{}{}
	}
	if len(enabledIndexesMap) == 0 {
		return nil, indexer.ErrEmptyEnabledIndexes
	}

	balanceConverter, err := converters.NewBalanceConverter(arguments.Denomination)
	if err != nil {
		return nil, err
	}

	accountsProc, err := accounts.NewAccountsProcessor(
		arguments.Marshalizer,
		arguments.AddressPubkeyConverter,
		arguments.AccountsDB,
		balanceConverter,
	)
	if err != nil {
		return nil, err
	}

	blockProcHandler, err := blockProc.NewBlockProcessor(arguments.Hasher, arguments.Marshalizer)
	if err != nil {
		return nil, err
	}

	miniblocksProc, err := miniblocks.NewMiniblocksProcessor(arguments.ShardCoordinator.SelfId(), arguments.Hasher, arguments.Marshalizer)
	if err != nil {
		return nil, err
	}
	validatorsProc, err := validators.NewValidatorsProcessor(arguments.ValidatorPubkeyConverter)
	if err != nil {
		return nil, err
	}

	generalInfoProc := statistics.NewStatisticsProcessor()

	argsTxsProc := &transactions.ArgsTransactionProcessor{
		AddressPubkeyConverter: arguments.AddressPubkeyConverter,
		TxFeeCalculator:        arguments.TransactionFeeCalculator,
		ShardCoordinator:       arguments.ShardCoordinator,
		Hasher:                 arguments.Hasher,
		Marshalizer:            arguments.Marshalizer,
		IsInImportMode:         arguments.IsInImportDBMode,
	}
	txsProc, err := transactions.NewTransactionsProcessor(argsTxsProc)
	if err != nil {
		return nil, err
	}

	argsLogsAndEventsProc := &logsevents.ArgsLogsAndEventsProcessor{
		ShardCoordinator: arguments.ShardCoordinator,
		PubKeyConverter:  arguments.AddressPubkeyConverter,
		Marshalizer:      arguments.Marshalizer,
		BalanceConverter: balanceConverter,
		Hasher:           arguments.Hasher,
		TxFeeCalculator:  arguments.TransactionFeeCalculator,
	}
	logsAndEventsProc, err := logsevents.NewLogsAndEventsProcessor(argsLogsAndEventsProc)
	if err != nil {
		return nil, err
	}

	args := &processIndexer.ArgElasticProcessor{
		TransactionsProc:  txsProc,
		AccountsProc:      accountsProc,
		BlockProc:         blockProcHandler,
		MiniblocksProc:    miniblocksProc,
		ValidatorsProc:    validatorsProc,
		StatisticsProc:    generalInfoProc,
		LogsAndEventsProc: logsAndEventsProc,
		DBClient:          arguments.DBClient,
		EnabledIndexes:    enabledIndexesMap,
		UseKibana:         arguments.UseKibana,
		IndexTemplates:    indexTemplates,
		IndexPolicies:     indexPolicies,
		SelfShardID:       arguments.ShardCoordinator.SelfId(),
	}

	return processIndexer.NewElasticProcessor(args)
}
