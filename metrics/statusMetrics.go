package metrics

import (
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-es-indexer-go/core/request"
)

type statusMetrics struct {
	metrics map[string]*request.MetricsResponse
	mut     sync.RWMutex
}

// NewStatusMetrics will return an instance of the statusMetrics
func NewStatusMetrics() *statusMetrics {
	return &statusMetrics{
		metrics: make(map[string]*request.MetricsResponse),
	}
}

// AddIndexingData will add the indexing data for the give topic
func (sm *statusMetrics) AddIndexingData(args ArgsAddIndexingData) {
	sm.mut.Lock()
	defer sm.mut.Unlock()

	topic := args.Topic
	_, found := sm.metrics[topic]
	if !found {
		sm.metrics[topic] = &request.MetricsResponse{}
	}

	sm.metrics[topic].OperationsCount++
	sm.metrics[topic].TotalIndexingTime += args.Duration
	sm.metrics[topic].TotalData += args.MessageLen

	if args.GotError {
		sm.metrics[topic].ErrorsCount++
	}
}

// GetMetrics returns the metrics map
func (sm *statusMetrics) GetMetrics() map[string]*request.MetricsResponse {
	sm.mut.RLock()
	defer sm.mut.RUnlock()

	return sm.getAllUnprotected()
}

// GetMetricsForPrometheus returns the metrics in a prometheus format
func (sm *statusMetrics) GetMetricsForPrometheus() string {
	sm.mut.RLock()
	metrics := sm.getAllUnprotected()
	sm.mut.RUnlock()

	stringBuilder := strings.Builder{}

	for topicWithShardID, metricsData := range metrics {
		topic, shardID := request.SplitTopicAndShardID(topicWithShardID)
		stringBuilder.WriteString(counterMetric(topic, shardID, metricsData.TotalData))
		stringBuilder.WriteString(counterMetric(topic, shardID, metricsData.ErrorsCount))
		stringBuilder.WriteString(counterMetric(topic, shardID, metricsData.OperationsCount))
		stringBuilder.WriteString(counterMetric(topic, shardID, uint64(metricsData.TotalIndexingTime.Milliseconds())))
	}

	promMetricsOutput := stringBuilder.String()

	return promMetricsOutput
}

func (sm *statusMetrics) getAllUnprotected() map[string]*request.MetricsResponse {
	newMap := make(map[string]*request.MetricsResponse)
	for key, value := range sm.metrics {
		newMap[key] = value
	}

	return newMap
}

// IsInterfaceNil returns true if there is no value under the interface
func (sm *statusMetrics) IsInterfaceNil() bool {
	return sm == nil
}
