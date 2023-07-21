package request

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type StringKey string

const (
	separator = "_"

	ContextKey  StringKey = "key"
	RemoveTopic           = "req_remove"
	GetTopic              = "req_get"
	BulkTopic             = "req_bulk"
	UpdateTopic           = "req_update"
	ScrollTopic           = "req_scroll"
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

	shardIDStr := split[len(split)-1]
	shardID, err := strconv.ParseUint(shardIDStr, 10, 32)
	if err != nil {
		return topicWithShardID, 0
	}

	return strings.Join(split[:len(split)-1], separator), uint32(shardID)
}
