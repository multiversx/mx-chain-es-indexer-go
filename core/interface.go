package core

import "time"

// StatusMetricsHandler defines the behavior of a component that handles status metrics
type StatusMetricsHandler interface {
	AddIndexingData(topic string, shardID uint32, duration time.Duration, gotError bool)
	GetMetrics() map[string]*IndexTopicMetricsResponse
	GetMetricsForPrometheus() string
	IsInterfaceNil() bool
}

// WebServerHandler defines the behavior of a component that handles the web server
type WebServerHandler interface {
	StartHttpServer() error
	Close() error
}
