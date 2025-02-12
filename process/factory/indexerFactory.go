package factory

import (
	"fmt"
	"net/http"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-es-indexer-go/client"
	"github.com/multiversx/mx-chain-es-indexer-go/client/logging"
	"github.com/multiversx/mx-chain-es-indexer-go/client/transport"
	indexerCore "github.com/multiversx/mx-chain-es-indexer-go/core"
	"github.com/multiversx/mx-chain-es-indexer-go/factory/runType"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/factory"
)

var log = logger.GetOrCreate("indexer/factory")

// ArgsIndexerFactory holds all dependencies required by the data indexer factory in order to create
// new instances
type ArgsIndexerFactory struct {
	Enabled                  bool
	UseKibana                bool
	ImportDB                 bool
	Sovereign                bool
	ESDTPrefix               string
	MainChainElastic         factory.ElasticConfig
	Denomination             int
	BulkRequestMaxSize       int
	Url                      string
	UserName                 string
	Password                 string
	TemplatesPath            string
	Version                  string
	EnabledIndexes           []string
	HeaderMarshaller         marshal.Marshalizer
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	StatusMetrics            indexerCore.StatusMetricsHandler
	RunTypeComponents        runType.RunTypeComponentsHandler
}

// NewIndexer will create a new instance of Indexer
func NewIndexer(args ArgsIndexerFactory) (dataindexer.Indexer, error) {
	err := checkDataIndexerParams(args)
	if err != nil {
		return nil, err
	}

	if args.Sovereign {
		args.RunTypeComponents, err = createManagedRunTypeComponents(runType.NewSovereignRunTypeComponentsFactory(args.MainChainElastic, args.ESDTPrefix))
	} else {
		args.RunTypeComponents, err = createManagedRunTypeComponents(runType.NewRunTypeComponentsFactory())
	}
	if err != nil {
		return nil, err
	}

	elasticProcessor, err := createElasticProcessor(args)
	if err != nil {
		return nil, err
	}

	blockContainer, err := createBlockCreatorsContainer()
	if err != nil {
		return nil, err
	}

	arguments := dataindexer.ArgDataIndexer{
		HeaderMarshaller: args.HeaderMarshaller,
		ElasticProcessor: elasticProcessor,
		BlockContainer:   blockContainer,
	}

	return dataindexer.NewDataIndexer(arguments)
}

func createManagedRunTypeComponents(factory runType.RunTypeComponentsCreator) (runType.RunTypeComponentsHandler, error) {
	managedRunTypeComponents, err := runType.NewManagedRunTypeComponents(factory)
	if err != nil {
		return nil, err
	}

	err = managedRunTypeComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedRunTypeComponents, nil
}

func createElasticProcessor(args ArgsIndexerFactory) (dataindexer.ElasticProcessor, error) {
	databaseClient, err := createElasticClient(args)
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
		EnabledIndexes:           args.EnabledIndexes,
		BulkRequestMaxSize:       args.BulkRequestMaxSize,
		ImportDB:                 args.ImportDB,
		Version:                  args.Version,
		TxHashExtractor:          args.RunTypeComponents.TxHashExtractorCreator(),
		RewardTxData:             args.RunTypeComponents.RewardTxDataCreator(),
		IndexTokensHandler:       args.RunTypeComponents.IndexTokensHandlerCreator(),
	}

	return factory.CreateElasticProcessor(argsElasticProcFac)
}

func createElasticClient(args ArgsIndexerFactory) (elasticproc.DatabaseClientHandler, error) {
	argsEsClient := elasticsearch.Config{
		Addresses:     []string{args.Url},
		Username:      args.UserName,
		Password:      args.Password,
		Logger:        &logging.CustomLogger{},
		RetryOnStatus: []int{http.StatusConflict},
		RetryBackoff:  client.RetryBackOff,
	}

	if check.IfNil(args.StatusMetrics) {
		return client.NewElasticClient(argsEsClient)
	}

	transportMetrics, err := transport.NewMetricsTransport(args.StatusMetrics)
	if err != nil {
		return nil, err
	}
	argsEsClient.Transport = transportMetrics

	return client.NewElasticClient(argsEsClient)
}

func checkDataIndexerParams(arguments ArgsIndexerFactory) error {
	if check.IfNil(arguments.AddressPubkeyConverter) {
		return fmt.Errorf("%w when setting AddressPubkeyConverter in indexer", dataindexer.ErrNilPubkeyConverter)
	}
	if check.IfNil(arguments.ValidatorPubkeyConverter) {
		return fmt.Errorf("%w when setting ValidatorPubkeyConverter in indexer", dataindexer.ErrNilPubkeyConverter)
	}
	if arguments.Url == "" {
		return dataindexer.ErrNilUrl
	}
	if check.IfNil(arguments.Marshalizer) {
		return dataindexer.ErrNilMarshalizer
	}
	if check.IfNil(arguments.Hasher) {
		return dataindexer.ErrNilHasher
	}
	if check.IfNil(arguments.HeaderMarshaller) {
		return fmt.Errorf("%w: header marshaller", dataindexer.ErrNilMarshalizer)
	}

	return nil
}

func createBlockCreatorsContainer() (dataindexer.BlockContainerHandler, error) {
	container := block.NewEmptyBlockCreatorsContainer()
	err := container.Add(core.ShardHeaderV1, block.NewEmptyHeaderCreator())
	if err != nil {
		return nil, err
	}
	err = container.Add(core.ShardHeaderV2, block.NewEmptyHeaderV2Creator())
	if err != nil {
		return nil, err
	}
	err = container.Add(core.MetaHeader, block.NewEmptyMetaBlockCreator())
	if err != nil {
		return nil, err
	}
	err = container.Add(core.SovereignChainHeader, block.NewEmptySovereignHeaderCreator())
	if err != nil {
		return nil, err
	}

	return container, nil
}
