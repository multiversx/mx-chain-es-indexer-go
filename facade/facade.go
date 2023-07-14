package facade

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-es-indexer-go/core"
)

type metricsFacade struct {
	statusMetrics core.StatusMetricsHandler
}

func NewMetricsFacade(statusMetrics core.StatusMetricsHandler) (*metricsFacade, error) {
	if check.IfNil(statusMetrics) {
		return nil, core.ErrNilMetricsHandler
	}

	return &metricsFacade{
		statusMetrics: statusMetrics,
	}, nil
}

// GetMetrics will return metrics in json format
func (mf *metricsFacade) GetMetrics() map[string]interface{} {
	return mf.statusMetrics.GetMetrics()
}

// GetMetricsForPrometheus will return metrics in prometheus format
func (mf *metricsFacade) GetMetricsForPrometheus() string {
	return mf.statusMetrics.GetMetricsForPrometheus()
}

// IsInterfaceNil returns true if there is no value under the interface
func (mf *metricsFacade) IsInterfaceNil() bool {
	return mf == nil
}
