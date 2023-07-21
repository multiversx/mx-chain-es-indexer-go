package transport

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/multiversx/mx-chain-es-indexer-go/core"
	"github.com/multiversx/mx-chain-es-indexer-go/core/request"
	"github.com/multiversx/mx-chain-es-indexer-go/metrics"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func TestNewMetricsTransport(t *testing.T) {
	t.Parallel()

	transportHandler, err := NewMetricsTransport(nil)
	require.Nil(t, transportHandler)
	require.Equal(t, core.ErrNilMetricsHandler, err)

	metricsHandler := metrics.NewStatusMetrics()
	transportHandler, err = NewMetricsTransport(metricsHandler)
	require.Nil(t, err)
	require.NotNil(t, transportHandler)
}

func TestMetricsTransport_RoundTrip(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.NewStatusMetrics()
	transportHandler, _ := NewMetricsTransport(metricsHandler)

	transportHandler.transport = &mock.TransportMock{
		Response: &http.Response{
			StatusCode: http.StatusOK,
		},
		Err: nil,
	}

	testTopic := "test"
	contextWithValue := context.WithValue(context.Background(), request.ContextKey, testTopic)
	req, _ := http.NewRequestWithContext(contextWithValue, http.MethodGet, "dummy", bytes.NewBuffer([]byte("test")))

	_, _ = transportHandler.RoundTrip(req)

	metricsMap := metricsHandler.GetMetrics()
	require.Equal(t, uint64(1), metricsMap[testTopic].OperationsCount)
	require.Equal(t, uint64(0), metricsMap[testTopic].ErrorsCount)
	require.Equal(t, uint64(4), metricsMap[testTopic].TotalData)
}

func TestMetricsTransport_RoundTripNoValueInContextShouldNotAddMetrics(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.NewStatusMetrics()
	transportHandler, _ := NewMetricsTransport(metricsHandler)

	transportHandler.transport = &mock.TransportMock{
		Response: &http.Response{
			StatusCode: http.StatusOK,
		},
		Err: nil,
	}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "dummy", bytes.NewBuffer([]byte("test")))
	_, _ = transportHandler.RoundTrip(req)

	metricsMap := metricsHandler.GetMetrics()
	require.Len(t, metricsMap, 0)
}
