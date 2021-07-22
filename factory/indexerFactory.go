package factory

import (
	"fmt"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/elastic/go-elasticsearch/v7"
)

// ArgsIndexerFactory holds all dependencies required by the data indexer factory in order to create
// new instances
type ArgsIndexerFactory struct {
	Enabled                  bool
	IndexerCacheSize         int
	ShardCoordinator         sharding.Coordinator
	Url                      string
	UserName                 string
	Password                 string
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	UseKibana                bool
	EnabledIndexes           []string
	Denomination             int
	AccountsDB               state.AccountsAdapter
	TransactionFeeCalculator process.TransactionFeeCalculator
	IsInImportDBMode         bool
}

// NewIndexer will create a new instance of Indexer
func NewIndexer(args *ArgsIndexerFactory) (process.Indexer, error) {
	err := checkDataIndexerParams(args)
	if err != nil {
		return nil, err
	}

	if !args.Enabled {
		return indexer.NewNilIndexer(), nil
	}

	elasticProcessor, err := createElasticProcessor(args)
	if err != nil {
		return nil, err
	}

	dispatcher, err := indexer.NewDataDispatcher(args.IndexerCacheSize)
	if err != nil {
		return nil, err
	}

	dispatcher.StartIndexData()

	arguments := indexer.ArgDataIndexer{
		Marshalizer:      args.Marshalizer,
		UseKibana:        args.UseKibana,
		ShardCoordinator: args.ShardCoordinator,
		ElasticProcessor: elasticProcessor,
		DataDispatcher:   dispatcher,
	}

	return indexer.NewDataIndexer(arguments)
}

func createDatabaseClient(url, userName, password string) (indexer.DatabaseClientHandler, error) {
	return indexer.NewElasticClient(elasticsearch.Config{
		Addresses: []string{url},
		Username:  userName,
		Password:  password,
	})
}

func createElasticProcessor(args *ArgsIndexerFactory) (indexer.ElasticProcessor, error) {
	databaseClient, err := createDatabaseClient(args.Url, args.UserName, args.Password)
	if err != nil {
		return nil, err
	}

	indexTemplates, indexPolicies, err := indexer.GetElasticTemplatesAndPolicies(args.UseKibana)
	if err != nil {
		return nil, err
	}

	enabledIndexesMap := make(map[string]struct{})
	for _, index := range args.EnabledIndexes {
		enabledIndexesMap[index] = struct{}{}
	}
	if len(enabledIndexesMap) == 0 {
		return nil, indexer.ErrEmptyEnabledIndexes
	}

	esIndexerArgs := indexer.ArgElasticProcessor{
		IndexTemplates:           indexTemplates,
		IndexPolicies:            indexPolicies,
		Marshalizer:              args.Marshalizer,
		Hasher:                   args.Hasher,
		AddressPubkeyConverter:   args.AddressPubkeyConverter,
		ValidatorPubkeyConverter: args.ValidatorPubkeyConverter,
		UseKibana:                args.UseKibana,
		DBClient:                 databaseClient,
		EnabledIndexes:           enabledIndexesMap,
		AccountsDB:               args.AccountsDB,
		Denomination:             args.Denomination,
		TransactionFeeCalculator: args.TransactionFeeCalculator,
		IsInImportDBMode:         args.IsInImportDBMode,
		ShardCoordinator:         args.ShardCoordinator,
	}

	return indexer.NewElasticProcessor(esIndexerArgs)
}

func checkDataIndexerParams(arguments *ArgsIndexerFactory) error {
	if arguments.IndexerCacheSize < 0 {
		return indexer.ErrNegativeCacheSize
	}
	if check.IfNil(arguments.AddressPubkeyConverter) {
		return fmt.Errorf("%w when setting AddressPubkeyConverter in indexer", indexer.ErrNilPubkeyConverter)
	}
	if check.IfNil(arguments.ValidatorPubkeyConverter) {
		return fmt.Errorf("%w when setting ValidatorPubkeyConverter in indexer", indexer.ErrNilPubkeyConverter)
	}
	if arguments.Url == "" {
		return core.ErrNilUrl
	}
	if check.IfNil(arguments.Marshalizer) {
		return core.ErrNilMarshalizer
	}
	if check.IfNil(arguments.Hasher) {
		return core.ErrNilHasher
	}
	if check.IfNil(arguments.TransactionFeeCalculator) {
		return core.ErrNilTransactionFeeCalculator
	}

	return nil
}
