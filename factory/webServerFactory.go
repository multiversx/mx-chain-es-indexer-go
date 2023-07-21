package factory

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-es-indexer-go/api/gin"
	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/core"
	"github.com/multiversx/mx-chain-es-indexer-go/facade"
)

// CreateWebServer will create a new instance of core.WebServerHandler
func CreateWebServer(apiConfig config.ApiRoutesConfig, statusMetricsHandler core.StatusMetricsHandler) (core.WebServerHandler, error) {
	if check.IfNil(statusMetricsHandler) {
		return nil, core.ErrNilMetricsHandler
	}

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
