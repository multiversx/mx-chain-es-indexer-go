package factory

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/outport/grpcadapter"
	factoryMarshaller "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/core"
	"github.com/multiversx/mx-chain-es-indexer-go/process/grpcAdapter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type IndexerCloseHandler interface {
	Close() error
}

func CreateGRPCIndexer(
	cfg config.Config,
	clusterCfg config.ClusterConfig,
	epochsCfg config.EnableEpochsConfig,
	statusMetrics core.StatusMetricsHandler,
	version string,
	configPathStr string,
) (IndexerCloseHandler, error) {
	headerMarshaller, err := factoryMarshaller.NewMarshalizer(clusterCfg.Config.WebSocket.DataMarshallerType)
	if err != nil {
		return nil, err
	}

	dataIndexer, err := createDataIndexer(cfg, clusterCfg, epochsCfg, headerMarshaller, statusMetrics, version, configPathStr)
	if err != nil {
		return nil, err
	}

	adapter, err := grpcAdapter.NewGRPCAdapter(dataIndexer)

	outportGRPCServer, err := grpcadapter.NewOutportGRPCServerWithAdapter(
		clusterCfg.Config.GRPCConfig.URL,
		adapter,
		grpc.UnaryInterceptor(requestLoggingInterceptor),
	)
	if err != nil {
		return nil, err
	}

	go func() {
		log.Info("starting indexer GRPC server", "url", clusterCfg.Config.GRPCConfig.URL)
		errS := outportGRPCServer.Start()
		log.LogIfError(errS, "outportGRPCServer closed with error")
	}()

	return outportGRPCServer, nil
}

func requestLoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()

	p, _ := peer.FromContext(ctx)

	resp, err := handler(ctx, req)

	log.Trace("grpc request",
		"method", info.FullMethod,
		"peer", p.Addr.String(),
		"duration", time.Since(start),
		"error", err,
	)

	return resp, err
}
