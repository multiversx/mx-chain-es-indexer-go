package request

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// StringKeyType defines the type for the context key
type StringKeyType string

const (
	noShardID = "#"
	// ContextKey the key for the value that will be added in the context
	ContextKey StringKeyType = "key"
	separator  string        = "_"
	// RemoveTopic is the identifier for the remove requests metrics
	RemoveTopic string = "req_remove"
	// GetTopic is the identifier for the get requests metrics
	GetTopic string = "req_get"
	// BulkTopic is the identifier for the bulk requests metrics
	BulkTopic string = "req_bulk"
	// UpdateTopic is the identifier for the update requests metrics
	UpdateTopic string = "req_update"
	// ScrollTopic is the identifier for the scroll requests metrics
	ScrollTopic string = "req_scroll"
)

// MetricsResponse defines the response for status metrics endpoint
type MetricsResponse struct {
	TotalData         uint64         `json:"total_data"`
	OperationsCount   uint64         `json:"operations_count"`
	TotalErrorsCount  uint64         `json:"total_errors_count"`
	ErrorsCount       map[int]uint64 `json:"errors_count,omitempty"`
	TotalIndexingTime time.Duration  `json:"total_time"`
}

// ExtendTopicWithShardID will concatenate topic with shardID
func ExtendTopicWithShardID(topic string, shardID uint32) string {
	return topic + separator + fmt.Sprintf("%d", shardID)
}

// SplitTopicAndShardID will extract shard id from the provided topic
func SplitTopicAndShardID(topicWithShardID string) (string, string) {
	split := strings.Split(topicWithShardID, separator)
	if len(split) < 2 {
		return topicWithShardID, noShardID
	}

	shardIDIndex := len(split) - 1
	shardIDStr := split[shardIDIndex]
	_, err := strconv.ParseUint(shardIDStr, 10, 32)
	if err != nil {
		return topicWithShardID, noShardID
	}

	return strings.Join(split[:shardIDIndex], separator), shardIDStr
}
