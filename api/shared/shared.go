package shared

import "github.com/gin-gonic/gin"

// MiddlewarePosition is the type that specifies the position of a middleware relative to the base endpoint handler
type MiddlewarePosition bool

const (
	// Before indicates that the middleware should be used before the base endpoint handler
	Before MiddlewarePosition = true

	// After indicates that the middleware should be used after the base endpoint handler
	After MiddlewarePosition = false
)

// AdditionalMiddleware holds the data needed for adding a middleware to an API endpoint
type AdditionalMiddleware struct {
	Middleware gin.HandlerFunc
	Position   MiddlewarePosition
}

// EndpointHandlerData holds the items needed for creating a new gin HTTP endpoint
type EndpointHandlerData struct {
	Path                  string
	Method                string
	Handler               gin.HandlerFunc
	AdditionalMiddlewares []AdditionalMiddleware
}

// GenericAPIResponse defines the structure of all responses on API endpoints
type GenericAPIResponse struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
	Code  string      `json:"code"`
}
