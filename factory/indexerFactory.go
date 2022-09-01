package factory

import (
	"fmt"
	"math"
	"net/http"
	"time"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/client"
	"github.com/ElrondNetwork/elastic-indexer-go/client/logging"
	"github.com/ElrondNetwork/elastic-indexer-go/process/factory"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/elastic/go-elasticsearch/v7"
)

var log = logger.GetOrCreate("indexer/factory")

// ArgsIndexerFactory holds all dependencies required by the data indexer factory in order to create
// new instances
type ArgsIndexerFactory struct {
	Enabled   bool
	UseKibana bool
	// move this in ArgSaveBlock
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
		ElasticProcessor: elasticProcessor,
		DataDispatcher:   dispatcher,
	}

	return indexer.NewDataIndexer(arguments)
}

func retryBackOff(attempt int) time.Duration {
	d := time.Duration(math.Exp2(float64(attempt))) * time.Second
	log.Debug("elastic: retry backoff", "attempt", attempt, "sleep duration", d)

	return d
}

func createElasticProcessor(args *ArgsIndexerFactory) (indexer.ElasticProcessor, error) {
	databaseClient, err := client.NewElasticClient(elasticsearch.Config{
		Addresses:     []string{args.Url},
		Username:      args.UserName,
		Password:      args.Password,
		Logger:        &logging.CustomLogger{},
		RetryOnStatus: []int{http.StatusConflict},
		RetryBackoff:  retryBackOff,
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
		Denomination:             args.Denomination,
		IsInImportDBMode:         args.IsInImportDBMode,
		ShardCoordinator:         args.ShardCoordinator,
		EnabledIndexes:           args.EnabledIndexes,
		BulkRequestMaxSize:       args.BulkRequestMaxSize,
	}

	return factory.CreateElasticProcessor(argsElasticProcFac)
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
	if check.IfNil(arguments.ShardCoordinator) {
		return indexer.ErrNilShardCoordinator
	}

	return nil
}
