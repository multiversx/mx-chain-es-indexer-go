package core

// StatusMetricsHandler defines the behavior of a component that handles status metrics
type StatusMetricsHandler interface {
	GetMetrics() map[string]interface{}
	GetMetricsForPrometheus() string
	IsInterfaceNil() bool
}

type WebServerHandler interface {
	StartHttpServer() error
	Close() error
}
