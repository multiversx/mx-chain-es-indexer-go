package factory

import (
	"github.com/multiversx/mx-chain-es-indexer-go/api/gin"
	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/core"
	"github.com/multiversx/mx-chain-es-indexer-go/facade"
)

func CreateWebServer(apiConfig config.ApiRoutesConfig, statusMetricsHandler core.StatusMetricsHandler) (core.WebServerHandler, error) {
	metricsFacade, err := facade.NewMetricsFacade(statusMetricsHandler)
	if err != nil {
		return nil, err
	}

	args := gin.ArgsWebServer{
		Facade:    metricsFacade,
		ApiConfig: apiConfig,
	}
	return gin.NewWebServer(args)
}
