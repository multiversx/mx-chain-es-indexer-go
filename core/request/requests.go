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
	OperationsCount   uint64        `json:"operations_count"`
	ErrorsCount       uint64        `json:"errors_count"`
	TotalIndexingTime time.Duration `json:"total_time"`
	TotalData         uint64        `json:"total_data"`
}

// ExtendTopicWithShardID will concatenate topic with shardID
func ExtendTopicWithShardID(topic string, shardID uint32) string {
	return topic + separator + fmt.Sprintf("%d", shardID)
}

// SplitTopicAndShardID will extract shard id from the provided topic
func SplitTopicAndShardID(topicWithShardID string) (string, uint32) {
	split := strings.Split(topicWithShardID, separator)
	if len(split) < 2 {
		return topicWithShardID, 0
	}

	shardIDIndex := len(split) - 1
	shardIDStr := split[shardIDIndex]
	shardID, err := strconv.ParseUint(shardIDStr, 10, 32)
	if err != nil {
		return topicWithShardID, 0
	}

	return strings.Join(split[:shardIDIndex], separator), uint32(shardID)
}
