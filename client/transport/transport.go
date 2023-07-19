package transport

import (
	"fmt"
	"net/http"
	"time"

	"github.com/multiversx/mx-chain-es-indexer-go/core"
	"github.com/multiversx/mx-chain-es-indexer-go/core/request"
	"github.com/multiversx/mx-chain-es-indexer-go/metrics"
)

type metricsTransport struct {
	statusMetrics core.StatusMetricsHandler
	transport     http.RoundTripper
}

func NewMetricsTransport(statusMetrics core.StatusMetricsHandler) (*metricsTransport, error) {
	return &metricsTransport{
		statusMetrics: statusMetrics,
		transport:     http.DefaultTransport,
	}, nil
}

func (m *metricsTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	startTime := time.Now()
	size := req.ContentLength
	resp, err := m.transport.RoundTrip(req)
	duration := time.Since(startTime)

	valueFromCtx := req.Context().Value(request.ContextKey)
	if valueFromCtx == nil {
		return resp, err
	}
	topic := fmt.Sprintf("%s", valueFromCtx)

	m.statusMetrics.AddIndexingData(metrics.ArgsAddIndexingData{
		GotError:   err != nil,
		MessageLen: uint64(size),
		Topic:      topic,
		Duration:   duration,
	})

	return resp, err
}
