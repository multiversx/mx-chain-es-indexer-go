package factory

import (
	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	factoryHost "github.com/multiversx/mx-chain-communication-go/websocket/factory"
	"github.com/multiversx/mx-chain-core-go/marshal"
	factoryMarshaller "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/process/wsindexer"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const (
	indexerCacheSize = 1
)

var log = logger.GetOrCreate("elasticindexer")

// CreateWsIndexer will create a new instance of wsindexer.WSClient
func CreateWsIndexer(cfg config.Config, clusterCfg config.ClusterConfig, importDB bool) (wsindexer.WSClient, error) {
	wsMarshaller, err := factoryMarshaller.NewMarshalizer(clusterCfg.Config.WebSocket.DataMarshallerType)
	if err != nil {
		return nil, err
	}

	host, err := createWsHost(clusterCfg, wsMarshaller)
	if err != nil {
		return nil, err
	}

	creator, err := newPayloadHandlerCreator(cfg, clusterCfg, importDB, wsMarshaller)
	if err != nil {
		return nil, err
	}

	err = host.SetPayloadHandlerCreator(creator)
	if err != nil {
		return nil, err
	}

	return host, nil
}

func createWsHost(clusterCfg config.ClusterConfig, wsMarshaller marshal.Marshalizer) (factoryHost.FullDuplexHost, error) {
	return factoryHost.CreateWebSocketHost(factoryHost.ArgsWebSocketHost{
		WebSocketConfig: data.WebSocketConfig{
			URL:                clusterCfg.Config.WebSocket.URL,
			WithAcknowledge:    clusterCfg.Config.WebSocket.WithAcknowledge,
			Mode:               clusterCfg.Config.WebSocket.Mode,
			RetryDurationInSec: int(clusterCfg.Config.WebSocket.RetryDurationInSec),
			BlockingAckOnError: clusterCfg.Config.WebSocket.BlockingAckOnError,
		},
		Marshaller: wsMarshaller,
		Log:        log,
	})
}
