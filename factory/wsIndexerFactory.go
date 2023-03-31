package factory

import (
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	factoryHasher "github.com/multiversx/mx-chain-core-go/hashing/factory"
	"github.com/multiversx/mx-chain-core-go/marshal"
	factoryMarshaller "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-core-go/websocketOutportDriver/client"
	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/process/factory"
	"github.com/multiversx/mx-chain-es-indexer-go/process/wsindexer"
)

const (
	indexerCacheSize = 1
)

// CreateWsIndexer will create a new instance of wsindexer.WSClient
func CreateWsIndexer(cfg config.Config, clusterCfg config.ClusterConfig) (wsindexer.WSClient, error) {
	wsMarshaller, err := factoryMarshaller.NewMarshalizer(clusterCfg.Config.WebSocket.DataMarshallerType)
	if err != nil {
		return nil, err
	}

	dataIndexer, err := createDataIndexer(cfg, clusterCfg, wsMarshaller)
	if err != nil {
		return nil, err
	}

	indexer, err := wsindexer.NewIndexer(wsMarshaller, dataIndexer)
	if err != nil {
		return nil, err
	}

	return client.CreateWsClient(client.ArgsCreateWsClient{
		Url:                clusterCfg.Config.WebSocket.ServerURL,
		RetryDurationInSec: clusterCfg.Config.WebSocket.RetryDurationInSec,
		BlockingAckOnError: clusterCfg.Config.WebSocket.BlockingAckOnError,
		PayloadProcessor:   indexer,
	})
}

func createDataIndexer(cfg config.Config, clusterCfg config.ClusterConfig, wsMarshaller marshal.Marshalizer) (wsindexer.DataIndexer, error) {
	marshaller, err := factoryMarshaller.NewMarshalizer(cfg.Config.Marshaller.Type)
	if err != nil {
		return nil, err
	}
	hasher, err := factoryHasher.NewHasher(cfg.Config.Hasher.Type)
	if err != nil {
		return nil, err
	}
	addressPubkeyConverter, err := pubkeyConverter.NewBech32PubkeyConverter(cfg.Config.AddressConverter.Length, cfg.Config.AddressConverter.Prefix)
	if err != nil {
		return nil, err
	}
	validatorPubkeyConverter, err := pubkeyConverter.NewHexPubkeyConverter(cfg.Config.ValidatorKeysConverter.Length)
	if err != nil {
		return nil, err
	}

	return factory.NewIndexer(factory.ArgsIndexerFactory{
		UseKibana:                clusterCfg.Config.ElasticCluster.UseKibana,
		IndexerCacheSize:         indexerCacheSize,
		Denomination:             cfg.Config.Economics.Denomination,
		BulkRequestMaxSize:       clusterCfg.Config.ElasticCluster.BulkRequestMaxSizeInBytes,
		Url:                      clusterCfg.Config.ElasticCluster.URL,
		UserName:                 clusterCfg.Config.ElasticCluster.UserName,
		Password:                 clusterCfg.Config.ElasticCluster.Password,
		EnabledIndexes:           prepareIndices(cfg.Config.AvailableIndices, clusterCfg.Config.DisabledIndices),
		Marshalizer:              marshaller,
		Hasher:                   hasher,
		AddressPubkeyConverter:   addressPubkeyConverter,
		ValidatorPubkeyConverter: validatorPubkeyConverter,
		HeaderMarshaller:         wsMarshaller,
	})
}

func prepareIndices(availableIndices, disabledIndices []string) []string {
	indices := make([]string, 0)

	mapDisabledIndices := make(map[string]struct{})
	for _, index := range disabledIndices {
		mapDisabledIndices[index] = struct{}{}
	}

	for _, availableIndex := range availableIndices {
		_, shouldSkip := mapDisabledIndices[availableIndex]
		if shouldSkip {
			continue
		}
		indices = append(indices, availableIndex)
	}

	return indices
}
