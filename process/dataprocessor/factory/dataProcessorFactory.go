package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataprocessor"
	"github.com/multiversx/mx-chain-es-indexer-go/process/outputs"
	"github.com/multiversx/mx-chain-es-indexer-go/process/outputs/elastic"
	"github.com/multiversx/mx-chain-es-indexer-go/process/outputs/elastic/templatesAndPolicies"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/accounts"
	blockProc "github.com/multiversx/mx-chain-es-indexer-go/process/processors/block"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/converters"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/logsevents"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/miniblocks"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/operations"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/statistics"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/transactions"
	"github.com/multiversx/mx-chain-es-indexer-go/process/processors/validators"
)

const elasticOutputName = "elasticsearch"

// ArgDataProcessorFactory is struct that is used to store all components that are needed to create a data processor
type ArgDataProcessorFactory struct {
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	DBClient                 elastic.DatabaseClientHandler
	EnabledIndexes           []string
	IndicesWithPolicy        []string
	Version                  string
	Denomination             int
	BulkRequestMaxSize       int
	NumWritesInParallel      int
	ImportDB                 bool
	EnableEpochsConfig       config.EnableEpochsConfig
	UseTemplatesFromFiles    bool
	ConfigPath               string
	// ExtraOutputs are the outputs that receive the prepared data after it was saved in Elasticsearch
	ExtraOutputs []outputs.OutputWithPolicy
}

// the same processors prepare the data for the data processor and serialize it for the Elasticsearch output
type transactionsProcessor interface {
	dataprocessor.TransactionsHandler
	elastic.DBTransactionsHandler
}

type logsAndEventsProcessor interface {
	dataprocessor.LogsAndEventsHandler
	elastic.DBLogsAndEventsHandler
}

type operationsProcessor interface {
	dataprocessor.OperationsHandler
	elastic.OperationsHandler
}

type processors struct {
	transactionsProc  transactionsProcessor
	accountsProc      elastic.DBAccountHandler
	blockProc         elastic.DBBlockHandler
	miniblocksProc    elastic.DBMiniblocksHandler
	validatorsProc    elastic.DBValidatorsHandler
	statisticsProc    elastic.DBStatisticsHandler
	logsAndEventsProc logsAndEventsProcessor
	operationsProc    operationsProcessor
}

// CreateDataProcessor will create a new instance of DataProcessor.
// The Elasticsearch output is always the first output, followed by the provided extra outputs
func CreateDataProcessor(arguments ArgDataProcessorFactory) (dataindexer.DataProcessor, error) {
	procs, err := createProcessors(arguments)
	if err != nil {
		return nil, err
	}

	elasticOutput, err := createElasticOutput(arguments, procs)
	if err != nil {
		return nil, err
	}

	allOutputs := make([]outputs.OutputWithPolicy, 0, len(arguments.ExtraOutputs)+1)
	allOutputs = append(allOutputs, outputs.OutputWithPolicy{
		Name:            elasticOutputName,
		Output:          elasticOutput,
		ContinueOnError: false,
	})
	allOutputs = append(allOutputs, arguments.ExtraOutputs...)

	return dataprocessor.NewDataProcessor(dataprocessor.ArgDataProcessor{
		TransactionsProc:  procs.transactionsProc,
		LogsAndEventsProc: procs.logsAndEventsProc,
		OperationsProc:    procs.operationsProc,
		Outputs:           allOutputs,
	})
}

// CreateElasticOutput will create a new instance of the Elasticsearch output
func CreateElasticOutput(arguments ArgDataProcessorFactory) (elastic.OutputHandler, error) {
	procs, err := createProcessors(arguments)
	if err != nil {
		return nil, err
	}

	return createElasticOutput(arguments, procs)
}

func createElasticOutput(arguments ArgDataProcessorFactory, procs *processors) (elastic.OutputHandler, error) {
	templatesAndPoliciesReader := templatesAndPolicies.NewTemplatesAndPolicyReader(
		arguments.UseTemplatesFromFiles,
		arguments.ConfigPath,
		arguments.EnabledIndexes,
		arguments.IndicesWithPolicy,
	)

	enabledIndexesMap := make(map[string]struct{})
	for _, index := range arguments.EnabledIndexes {
		enabledIndexesMap[index] = struct{}{}
	}

	args := &elastic.ArgElasticOutput{
		BulkRequestMaxSize:  arguments.BulkRequestMaxSize,
		TransactionsProc:    procs.transactionsProc,
		AccountsProc:        procs.accountsProc,
		BlockProc:           procs.blockProc,
		MiniblocksProc:      procs.miniblocksProc,
		ValidatorsProc:      procs.validatorsProc,
		StatisticsProc:      procs.statisticsProc,
		LogsAndEventsProc:   procs.logsAndEventsProc,
		DBClient:            arguments.DBClient,
		EnabledIndexes:      enabledIndexesMap,
		OperationsProc:      procs.operationsProc,
		Version:             arguments.Version,
		MappingsHandler:     templatesAndPoliciesReader,
		NumWritesInParallel: arguments.NumWritesInParallel,
	}

	return elastic.NewElasticOutput(args)
}

func createProcessors(arguments ArgDataProcessorFactory) (*processors, error) {
	if len(arguments.EnabledIndexes) == 0 {
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

	blockProcHandler, err := blockProc.NewBlockProcessor(arguments.Hasher, arguments.Marshalizer, arguments.ValidatorPubkeyConverter)
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

	argsTxsProc := &transactions.ArgsTransactionProcessor{
		AddressPubkeyConverter: arguments.AddressPubkeyConverter,
		Hasher:                 arguments.Hasher,
		Marshalizer:            arguments.Marshalizer,
		BalanceConverter:       balanceConverter,
		EnableEpochsConfig:     arguments.EnableEpochsConfig,
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

	return &processors{
		transactionsProc:  txsProc,
		accountsProc:      accountsProc,
		blockProc:         blockProcHandler,
		miniblocksProc:    miniblocksProc,
		validatorsProc:    validatorsProc,
		statisticsProc:    statistics.NewStatisticsProcessor(),
		logsAndEventsProc: logsAndEventsProc,
		operationsProc:    operationsProc,
	}, nil
}
