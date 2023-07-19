package core

import "time"

// IndexTopicMetricsResponse defines the response for status metrics endpoint
type IndexTopicMetricsResponse struct {
	NumIndexingOperations uint64        `json:"num_indexing_operations"`
	NumTotalErrors        uint64        `json:"num_total_errors"`
	LastIndexingTime      time.Duration `json:"last_indexing_time"`
	TotalIndexingTime     time.Duration `json:"total_indexing_time"`
}
