package request

import "time"

const (
	ContextKey  = "key"
	RemoveTopic = "req_remove"
	GetTopic    = "req_get"
	BulkTopic   = "req_bulk"
	UpdateTopic = "req_update"
)

// MetricsResponse defines the response for status metrics endpoint
type MetricsResponse struct {
	OperationsCount   uint64        `json:"operations_count"`
	ErrorsCount       uint64        `json:"errors_count"`
	TotalIndexingTime time.Duration `json:"total_time"`
	TotalData         uint64        `json:"total_data"`
}
