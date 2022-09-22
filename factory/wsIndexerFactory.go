package factory

import (
	"github.com/ElrondNetwork/elastic-indexer-go/config"
	"github.com/ElrondNetwork/elastic-indexer-go/process/factory"
	"github.com/ElrondNetwork/elastic-indexer-go/process/wsclient"
	"github.com/ElrondNetwork/elastic-indexer-go/process/wsindexer"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	factoryHasher "github.com/ElrondNetwork/elrond-go-core/hashing/factory"
	factoryMarshaller "github.com/ElrondNetwork/elrond-go-core/marshal/factory"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("factory")

func CreateWsIndexer(cfg *config.Config) (wsindexer.WSClient, error) {
	dataIndexer, err := createDataIndexer(cfg)
	if err != nil {
		return nil, err
	}

	wsMarshaller, err := factoryMarshaller.NewMarshalizer(cfg.Config.WebSocket.DataMarshallerType)
	if err != nil {
		return nil, err
	}

	indexer, err := wsindexer.NewIndexer(wsMarshaller, dataIndexer)

	return wsclient.NewWebSocketClient(cfg.Config.WebSocket.ServerURL, indexer.GetFunctionsMap())
}

func createDataIndexer(cfg *config.Config) (wsindexer.DataIndexer, error) {
	marshaller, err := factoryMarshaller.NewMarshalizer(cfg.Config.Marshaller.Type)
	if err != nil {
		return nil, err
	}
	hasher, err := factoryHasher.NewHasher(cfg.Config.Hasher.Type)
	if err != nil {
		return nil, err
	}
	addressPubkeyConverter, err := pubkeyConverter.NewBech32PubkeyConverter(cfg.Config.AddressConverter.Length, log)
	if err != nil {
		return nil, err
	}
	validatorPubkeyConverter, err := pubkeyConverter.NewHexPubkeyConverter(cfg.Config.ValidatorKeysConverter.Length)
	if err != nil {
		return nil, err
	}

	return factory.NewIndexer(&factory.ArgsIndexerFactory{
		UseKibana: cfg.Config.ElasticCluster.UseKibana,
		// Todo check if this is needed
		IndexerCacheSize:         1,
		Denomination:             cfg.Config.Economics.Denomination,
		BulkRequestMaxSize:       cfg.Config.ElasticCluster.BulkRequestMaxSizeInBytes,
		Url:                      cfg.Config.ElasticCluster.URL,
		UserName:                 cfg.Config.ElasticCluster.UserName,
		Password:                 cfg.Config.ElasticCluster.Password,
		EnabledIndexes:           cfg.Config.EnabledIndices,
		Marshalizer:              marshaller,
		Hasher:                   hasher,
		AddressPubkeyConverter:   addressPubkeyConverter,
		ValidatorPubkeyConverter: validatorPubkeyConverter,
	})
}
