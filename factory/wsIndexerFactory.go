package factory

import (
	"errors"
	"time"

	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	factoryHost "github.com/multiversx/mx-chain-communication-go/websocket/factory"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	factoryHasher "github.com/multiversx/mx-chain-core-go/hashing/factory"
	"github.com/multiversx/mx-chain-core-go/marshal"
	factoryMarshaller "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/core"
	"github.com/multiversx/mx-chain-es-indexer-go/process/factory"
	"github.com/multiversx/mx-chain-es-indexer-go/process/outputs"
	"github.com/multiversx/mx-chain-es-indexer-go/process/outputs/rabbitmq"
	"github.com/multiversx/mx-chain-es-indexer-go/process/wsindexer"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("elasticindexer")

const (
	rabbitOutputName            = "rabbitmq"
	rabbitConnectionName        = "mx-chain-es-indexer"
	defaultRabbitPublishTimeout = 10 * time.Second
)

// wsClientWithIndexer closes the indexer (and all its outputs) when the web socket host is closed.
// The web socket host closes the indexer only if it is started in the server mode
type wsClientWithIndexer struct {
	wsindexer.WSClient
	indexer interface {
		Close() error
	}
}

// Close closes the web socket host and the indexer
func (w *wsClientWithIndexer) Close() error {
	errHost := w.WSClient.Close()
	errIndexer := w.indexer.Close()

	return errors.Join(errHost, errIndexer)
}

// CreateWsIndexer will create a new instance of wsindexer.WSClient
func CreateWsIndexer(
	cfg config.Config,
	clusterCfg config.ClusterConfig,
	epochsCfg config.EnableEpochsConfig,
	statusMetrics core.StatusMetricsHandler,
	version string,
	configPathStr string,
) (wsindexer.WSClient, error) {
	wsMarshaller, err := factoryMarshaller.NewMarshalizer(clusterCfg.Config.WebSocket.DataMarshallerType)
	if err != nil {
		return nil, err
	}

	dataIndexer, err := createDataIndexer(cfg, clusterCfg, epochsCfg, wsMarshaller, statusMetrics, version, configPathStr)
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

	return &wsClientWithIndexer{
		WSClient: host,
		indexer:  indexer,
	}, nil
}

func createDataIndexer(
	cfg config.Config,
	clusterCfg config.ClusterConfig,
	enableEpochsCfg config.EnableEpochsConfig,
	wsMarshaller marshal.Marshalizer,
	statusMetrics core.StatusMetricsHandler,
	version string,
	configPathStr string,
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

	indicesWithPolicies := make([]string, 0)
	if cfg.Config.Policies.Enabled {
		indicesWithPolicies = prepareIndices(cfg.Config.Policies.IndicesWithPolicy, clusterCfg.Config.DisabledIndices)
	}

	extraOutputs, err := createExtraOutputs(clusterCfg, statusMetrics)
	if err != nil {
		return nil, err
	}

	return factory.NewIndexer(factory.ArgsIndexerFactory{
		Denomination:             cfg.Config.Economics.Denomination,
		BulkRequestMaxSize:       clusterCfg.Config.ElasticCluster.BulkRequestMaxSizeInBytes,
		Url:                      clusterCfg.Config.ElasticCluster.URL,
		UserName:                 clusterCfg.Config.ElasticCluster.UserName,
		Password:                 clusterCfg.Config.ElasticCluster.Password,
		EnabledIndexes:           prepareIndices(cfg.Config.AvailableIndices, clusterCfg.Config.DisabledIndices),
		IndicesWithPolicy:        indicesWithPolicies,
		Marshalizer:              marshaller,
		Hasher:                   hasher,
		AddressPubkeyConverter:   addressPubkeyConverter,
		ValidatorPubkeyConverter: validatorPubkeyConverter,
		HeaderMarshaller:         wsMarshaller,
		StatusMetrics:            statusMetrics,
		Version:                  version,
		EnableEpochsConfig:       enableEpochsCfg,
		NumWritesInParallel:      clusterCfg.Config.ElasticCluster.NumWritesInParallel,
		UseTemplatesFromFiles:    cfg.Config.UseTemplatesFromFiles,
		ConfigPath:               configPathStr,
		ExtraOutputs:             extraOutputs,
	})
}

// createExtraOutputs creates the outputs that receive the prepared data after it was saved in Elasticsearch
func createExtraOutputs(clusterCfg config.ClusterConfig, statusMetrics core.StatusMetricsHandler) ([]outputs.OutputWithPolicy, error) {
	rabbitCfg := clusterCfg.Config.RabbitMQ
	if !rabbitCfg.Enabled {
		return nil, nil
	}

	if !rabbitCfg.ContinueOnError && !clusterCfg.Config.WebSocket.BlockingAckOnError {
		log.Warn("RabbitMQ: a block cannot be sent again by the node because blocking-ack-on-error is false, " +
			"if a message cannot be published it will be lost")
	}

	publishTimeout := time.Duration(rabbitCfg.PublishTimeoutInSec) * time.Second
	if publishTimeout == 0 {
		publishTimeout = defaultRabbitPublishTimeout
	}

	rabbitClient, err := rabbitmq.NewClient(rabbitmq.ArgsClient{
		URL: rabbitCfg.URL,
		Exchanges: []rabbitmq.ExchangeConfig{
			{Name: rabbitCfg.BlocksExchange.Name, Type: rabbitCfg.BlocksExchange.Type},
			{Name: rabbitCfg.RevertsExchange.Name, Type: rabbitCfg.RevertsExchange.Type},
		},
		ConnectionName: rabbitConnectionName,
		// if the errors are only logged, the indexing should not wait for the RabbitMQ server
		FailFast:                 rabbitCfg.ContinueOnError,
		RequireInitialConnection: !rabbitCfg.ContinueOnError,
	})
	if err != nil {
		return nil, err
	}

	rabbitOutput, err := rabbitmq.NewOutput(rabbitmq.ArgsOutput{
		Client:          rabbitClient,
		BlocksExchange:  rabbitCfg.BlocksExchange.Name,
		RevertsExchange: rabbitCfg.RevertsExchange.Name,
		PublishTimeout:  publishTimeout,
		StatusMetrics:   statusMetrics,
	})
	if err != nil {
		_ = rabbitClient.Close()
		return nil, err
	}

	log.Info("RabbitMQ output enabled",
		"blocks exchange", rabbitCfg.BlocksExchange.Name,
		"reverts exchange", rabbitCfg.RevertsExchange.Name,
		"continue on error", rabbitCfg.ContinueOnError,
		"publish timeout", publishTimeout,
	)

	return []outputs.OutputWithPolicy{
		{
			Name:            rabbitOutputName,
			Output:          rabbitOutput,
			ContinueOnError: rabbitCfg.ContinueOnError,
		},
	}, nil
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
