package factory

import (
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	factoryHasher "github.com/multiversx/mx-chain-core-go/hashing/factory"
	"github.com/multiversx/mx-chain-core-go/marshal"
	factoryMarshaller "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-core-go/webSocket"
	"github.com/multiversx/mx-chain-core-go/webSocket/client"
	"github.com/multiversx/mx-chain-core-go/webSocket/server"
	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/process/factory"
	"github.com/multiversx/mx-chain-es-indexer-go/process/wsindexer"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const (
	indexerCacheSize = 1
)

var log = logger.GetOrCreate("elasticindexer")

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

	host, err := createWsHost(clusterCfg)
	if err != nil {
		return nil, err
	}

	err = host.SetPayloadHandler(indexer)
	if err != nil {
		return nil, err
	}

	return host, nil
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

func createWsHost(clusterCfg config.ClusterConfig) (webSocket.HostWebSocket, error) {
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	payloadConverter, err := webSocket.NewWebSocketPayloadConverter(uint64Converter)
	if err != nil {
		return nil, err
	}

	if clusterCfg.Config.WebSocket.IsServer {
		return server.NewWebSocketServer(server.ArgsWebSocketServer{
			RetryDurationInSeconds: int(clusterCfg.Config.WebSocket.RetryDurationInSec),
			BlockingAckOnError:     clusterCfg.Config.WebSocket.BlockingAckOnError,
			WithAcknowledge:        clusterCfg.Config.WebSocket.WithAcknowledge,
			URL:                    clusterCfg.Config.WebSocket.ServerURL,
			PayloadConverter:       payloadConverter,
			Log:                    log,
		})
	}

	return client.NewWebSocketClient(client.ArgsWebSocketClient{
		RetryDurationInSeconds: int(clusterCfg.Config.WebSocket.RetryDurationInSec),
		BlockingAckOnError:     clusterCfg.Config.WebSocket.BlockingAckOnError,
		WithAcknowledge:        clusterCfg.Config.WebSocket.WithAcknowledge,
		URL:                    clusterCfg.Config.WebSocket.ServerURL,
		PayloadConverter:       payloadConverter,
		Log:                    log,
	})
}
