package factory

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/client"
	"github.com/ElrondNetwork/elastic-indexer-go/client/logging"
	postgres "github.com/ElrondNetwork/elastic-indexer-go/client/postgresql"
	"github.com/ElrondNetwork/elastic-indexer-go/process/factory"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/elastic/go-elasticsearch/v7"
)

// ArgsIndexerFactory holds all dependencies required by the data indexer factory in order to create
// new instances
type ArgsIndexerFactory struct {
	Enabled                  bool
	UseKibana                bool
	IsInImportDBMode         bool
	IndexerCacheSize         int
	Denomination             int
	BulkRequestMaxSize       int
	Url                      string
	UserName                 string
	Password                 string
	TemplatesPath            string
	EnabledIndexes           []string
	ShardCoordinator         indexer.ShardCoordinator
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	AccountsDB               indexer.AccountsAdapter
	TransactionFeeCalculator indexer.FeesProcessorHandler
	PostgresURL              string
	PostgresDBName           string
	UsePostgres              bool
}

// NewIndexer will create a new instance of Indexer
func NewIndexer(args *ArgsIndexerFactory) (indexer.Indexer, error) {
	err := checkDataIndexerParams(args)
	if err != nil {
		return nil, err
	}

	if !args.Enabled {
		return indexer.NewNilIndexer(), nil
	}

	var processor indexer.ElasticProcessor
	if args.UsePostgres {
		processor, err = createPostgresProcessor(args)
		if err != nil {
			return nil, err
		}
	} else {
		processor, err = createElasticProcessor(args)
		if err != nil {
			return nil, err
		}
	}

	dispatcher, err := indexer.NewDataDispatcher(args.IndexerCacheSize)
	if err != nil {
		return nil, err
	}

	dispatcher.StartIndexData()

	arguments := indexer.ArgDataIndexer{
		Marshalizer:      args.Marshalizer,
		ShardCoordinator: args.ShardCoordinator,
		ElasticProcessor: processor,
		DataDispatcher:   dispatcher,
	}

	return indexer.NewDataIndexer(arguments)
}

func createElasticProcessor(args *ArgsIndexerFactory) (indexer.ElasticProcessor, error) {
	databaseClient, err := client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{args.Url},
		Username:  args.UserName,
		Password:  args.Password,
		Logger:    &logging.CustomLogger{},
	})
	if err != nil {
		return nil, err
	}

	argsElasticProcFac := factory.ArgElasticProcessorFactory{
		Marshalizer:              args.Marshalizer,
		Hasher:                   args.Hasher,
		AddressPubkeyConverter:   args.AddressPubkeyConverter,
		ValidatorPubkeyConverter: args.ValidatorPubkeyConverter,
		UseKibana:                args.UseKibana,
		DBClient:                 databaseClient,
		AccountsDB:               args.AccountsDB,
		Denomination:             args.Denomination,
		TransactionFeeCalculator: args.TransactionFeeCalculator,
		IsInImportDBMode:         args.IsInImportDBMode,
		ShardCoordinator:         args.ShardCoordinator,
		EnabledIndexes:           args.EnabledIndexes,
		BulkRequestMaxSize:       args.BulkRequestMaxSize,
	}

	return factory.CreateElasticProcessor(argsElasticProcFac)
}

func createPostgresProcessor(args *ArgsIndexerFactory) (indexer.ElasticProcessor, error) {
	host, port, err := getHostAndPortFromURL(args.PostgresURL)
	if err != nil {
		return nil, err
	}

	postgresArgs := &postgres.ArgsPostgresClient{
		Hostname: host,
		Port:     port,
		Username: args.UserName,
		Password: args.Password,
		DBName:   args.PostgresDBName,
	}
	databaseClient, err := postgres.NewPostgresClient(postgresArgs)
	if err != nil {
		return nil, err
	}

	argsElasticProcFac := factory.ArgPostgresProcessorFactory{
		Marshalizer:              args.Marshalizer,
		Hasher:                   args.Hasher,
		AddressPubkeyConverter:   args.AddressPubkeyConverter,
		ValidatorPubkeyConverter: args.ValidatorPubkeyConverter,
		UseKibana:                args.UseKibana,
		DBClient:                 databaseClient,
		AccountsDB:               args.AccountsDB,
		Denomination:             args.Denomination,
		TransactionFeeCalculator: args.TransactionFeeCalculator,
		IsInImportDBMode:         args.IsInImportDBMode,
		ShardCoordinator:         args.ShardCoordinator,
		EnabledIndexes:           args.EnabledIndexes,
	}

	return factory.CreatePostgresProcessor(argsElasticProcFac)
}

func getHostAndPortFromURL(urlStr string) (string, int, error) {
	u, err := url.Parse(urlStr)
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return "", 0, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}

	return host, port, nil
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
		return indexer.ErrNilUrl
	}
	if check.IfNil(arguments.Marshalizer) {
		return indexer.ErrNilMarshalizer
	}
	if check.IfNil(arguments.Hasher) {
		return indexer.ErrNilHasher
	}
	if check.IfNil(arguments.TransactionFeeCalculator) {
		return indexer.ErrNilTransactionFeeCalculator
	}
	if check.IfNil(arguments.AccountsDB) {
		return indexer.ErrNilAccountsDB
	}
	if check.IfNil(arguments.ShardCoordinator) {
		return indexer.ErrNilShardCoordinator
	}

	return nil
}
