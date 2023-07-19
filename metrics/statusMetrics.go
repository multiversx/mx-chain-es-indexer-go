package metrics

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-es-indexer-go/core"
)

type statusMetrics struct {
	metrics map[string]*core.IndexTopicMetricsResponse
	mut     sync.RWMutex
}

// NewStatusMetrics will return an instance of the statusMetrics
func NewStatusMetrics() *statusMetrics {
	return &statusMetrics{
		metrics: make(map[string]*core.IndexTopicMetricsResponse),
	}
}

// AddIndexingData will add the indexing data for the give topic
func (sm *statusMetrics) AddIndexingData(topic string, shardID uint32, duration time.Duration, gotError bool) {
	sm.mut.Lock()
	defer sm.mut.Unlock()

	key := fmt.Sprintf("%s_%d", topic, shardID)

	_, found := sm.metrics[key]
	if !found {
		sm.metrics[key] = &core.IndexTopicMetricsResponse{}
	}

	sm.metrics[key].NumIndexingOperations++
	sm.metrics[key].TotalIndexingTime += duration
	sm.metrics[key].LastIndexingTime = duration

	if gotError == true {
		sm.metrics[key].NumTotalErrors++
	}
}

// GetMetrics returns the metrics map
func (sm *statusMetrics) GetMetrics() map[string]*core.IndexTopicMetricsResponse {
	sm.mut.RLock()
	defer sm.mut.RUnlock()

	return sm.getAllUnprotected()
}

// GetMetricsForPrometheus returns the metrics in a prometheus format
func (sm *statusMetrics) GetMetricsForPrometheus() string {
	sm.mut.RLock()
	defer sm.mut.RUnlock()

	//metricsMap := sm.getAll()

	stringBuilder := strings.Builder{}

	// TODO populate with metrics

	return stringBuilder.String()
}

func (sm *statusMetrics) getAllUnprotected() map[string]*core.IndexTopicMetricsResponse {
	newMap := make(map[string]*core.IndexTopicMetricsResponse)
	for key, value := range sm.metrics {
		newMap[key] = value
	}

	return newMap
}

// IsInterfaceNil returns true if there is no value under the interface
func (sm *statusMetrics) IsInterfaceNil() bool {
	return sm == nil
}
