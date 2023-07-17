package core

// StatusMetricsHandler defines the behavior of a component that handles status metrics
type StatusMetricsHandler interface {
	GetMetrics() map[string]interface{}
	GetMetricsForPrometheus() string
	IsInterfaceNil() bool
}

// WebServerHandler defines the behavior of a component that handles the web server
type WebServerHandler interface {
	StartHttpServer() error
	Close() error
}
