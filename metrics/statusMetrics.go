package metrics

import (
	"strings"
	"sync"
)

//const (
//	numWSConnections  = "num_ws_connections"
//	numIndexingErrors = "num_indexing_errors"
//)

type statusMetrics struct {
	metrics             map[string]interface{}
	mutEndpointsMetrics sync.RWMutex
}

// NewStatusMetrics will return an instance of the statusMetrics
func NewStatusMetrics() *statusMetrics {
	return &statusMetrics{
		metrics: make(map[string]interface{}),
	}
}

// GetMetrics returns the metrics map
func (sm *statusMetrics) GetMetrics() map[string]interface{} {
	sm.mutEndpointsMetrics.RLock()
	defer sm.mutEndpointsMetrics.RUnlock()

	return sm.getAll()
}

// GetMetricsForPrometheus returns the metrics in a prometheus format
func (sm *statusMetrics) GetMetricsForPrometheus() string {
	sm.mutEndpointsMetrics.RLock()
	defer sm.mutEndpointsMetrics.RUnlock()

	//metricsMap := sm.getAll()

	stringBuilder := strings.Builder{}

	// TODO populate with metrics

	return stringBuilder.String()
}

func (sm *statusMetrics) getAll() map[string]interface{} {
	newMap := make(map[string]interface{})
	for key, value := range sm.metrics {
		newMap[key] = value
	}

	return newMap
}

// IsInterfaceNil returns true if there is no value under the interface
func (sm *statusMetrics) IsInterfaceNil() bool {
	return sm == nil
}
