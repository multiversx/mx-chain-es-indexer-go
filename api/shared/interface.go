package shared

import (
	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-es-indexer-go/config"
)

// GroupHandler defines the actions needed to be performed by a gin API group
type GroupHandler interface {
	RegisterRoutes(
		ws *gin.RouterGroup,
		apiConfig config.ApiRoutesConfig,
	)
	IsInterfaceNil() bool
}

// FacadeHandler defines all the methods that a facade should implement
type FacadeHandler interface {
	GetMetrics() map[string]interface{}
	GetMetricsForPrometheus() string
	IsInterfaceNil() bool
}

// HttpServerCloser defines the basic actions of starting and closing that a web server should be able to do
type HttpServerCloser interface {
	Start()
	Close() error
	IsInterfaceNil() bool
}
