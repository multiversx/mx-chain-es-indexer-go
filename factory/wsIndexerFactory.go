package factory

import (
	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	factoryHost "github.com/multiversx/mx-chain-communication-go/websocket/factory"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	factoryHasher "github.com/multiversx/mx-chain-core-go/hashing/factory"
	"github.com/multiversx/mx-chain-core-go/marshal"
	factoryMarshaller "github.com/multiversx/mx-chain-core-go/marshal/factory"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/core"
	esFactory "github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/factory"
	"github.com/multiversx/mx-chain-es-indexer-go/process/factory"
	"github.com/multiversx/mx-chain-es-indexer-go/process/wsindexer"
)

var log = logger.GetOrCreate("elasticindexer")

// CreateWsIndexer will create a new instance of wsindexer.WSClient
func CreateWsIndexer(cfg config.Config, clusterCfg config.ClusterConfig, statusMetrics core.StatusMetricsHandler, version string) (wsindexer.WSClient, error) {
	wsMarshaller, err := factoryMarshaller.NewMarshalizer(clusterCfg.Config.WebSocket.DataMarshallerType)
	if err != nil {
		return nil, err
	}

	dataIndexer, err := createDataIndexer(cfg, clusterCfg, wsMarshaller, statusMetrics, version)
	if err != nil {
		return nil, err
	}

	args := wsindexer.ArgsIndexer{
		Marshaller:    wsMarshaller,
		DataIndexer:   dataIndexer,
		StatusMetrics: statusMetrics,
	}
	indexer, err := wsindexer.NewIndexer(args)
	if err != nil {
		return nil, err
	}

	host, err := createWsHost(clusterCfg, wsMarshaller)
	if err != nil {
		return nil, err
	}

	err = host.SetPayloadHandler(indexer)
	if err != nil {
		return nil, err
	}

	return host, nil
}

func createDataIndexer(
	cfg config.Config,
	clusterCfg config.ClusterConfig,
	wsMarshaller marshal.Marshalizer,
	statusMetrics core.StatusMetricsHandler,
	version string,
) (wsindexer.DataIndexer, error) {
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

	mainChainElastic := esFactory.ElasticConfig{
		Enabled:  clusterCfg.Config.MainChainCluster.Enabled,
		Url:      clusterCfg.Config.MainChainCluster.URL,
		UserName: clusterCfg.Config.MainChainCluster.UserName,
		Password: clusterCfg.Config.MainChainCluster.Password,
	}

	return factory.NewIndexer(factory.ArgsIndexerFactory{
		Sovereign:                cfg.Sovereign,
		MainChainElastic:         mainChainElastic,
		UseKibana:                clusterCfg.Config.ElasticCluster.UseKibana,
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
		StatusMetrics:            statusMetrics,
		Version:                  version,
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

func createWsHost(clusterCfg config.ClusterConfig, wsMarshaller marshal.Marshalizer) (factoryHost.FullDuplexHost, error) {
	return factoryHost.CreateWebSocketHost(factoryHost.ArgsWebSocketHost{
		WebSocketConfig: data.WebSocketConfig{
			URL:                     clusterCfg.Config.WebSocket.URL,
			WithAcknowledge:         clusterCfg.Config.WebSocket.WithAcknowledge,
			Mode:                    clusterCfg.Config.WebSocket.Mode,
			RetryDurationInSec:      int(clusterCfg.Config.WebSocket.RetryDurationInSec),
			AcknowledgeTimeoutInSec: int(clusterCfg.Config.WebSocket.AckTimeoutInSec),
			BlockingAckOnError:      clusterCfg.Config.WebSocket.BlockingAckOnError,
		},
		Marshaller: wsMarshaller,
		Log:        log,
	})
}
