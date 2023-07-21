package metrics

import (
	"testing"

	"github.com/multiversx/mx-chain-es-indexer-go/core/request"
	"github.com/stretchr/testify/require"
)

func TestStatusMetrics_AddIndexingDataAndGetMetrics(t *testing.T) {
	t.Parallel()

	statusMetricsHandler := NewStatusMetrics()
	require.False(t, statusMetricsHandler.IsInterfaceNil())

	topic1 := "test1_0"
	statusMetricsHandler.AddIndexingData(ArgsAddIndexingData{
		GotError:   true,
		MessageLen: 100,
		Topic:      topic1,
		Duration:   12,
	})
	statusMetricsHandler.AddIndexingData(ArgsAddIndexingData{
		GotError:   false,
		MessageLen: 222,
		Topic:      topic1,
		Duration:   15,
	})

	topic2 := "test2_2"
	statusMetricsHandler.AddIndexingData(ArgsAddIndexingData{
		GotError:   true,
		MessageLen: 100,
		Topic:      topic2,
		Duration:   12,
	})

	metrics := statusMetricsHandler.GetMetrics()
	require.Equal(t, &request.MetricsResponse{
		OperationsCount:   2,
		ErrorsCount:       1,
		TotalIndexingTime: 27,
		TotalData:         322,
	}, metrics[topic1])
	require.Equal(t, &request.MetricsResponse{
		OperationsCount:   1,
		ErrorsCount:       1,
		TotalIndexingTime: 12,
		TotalData:         100,
	}, metrics[topic2])

	prometheusMetrics := statusMetricsHandler.GetMetricsForPrometheus()
	require.Equal(t, "# TYPE test1 counter\ntest1{shardID=\"0\"} 322\n\n# TYPE test1 counter\ntest1{shardID=\"0\"} 1\n\n# TYPE test1 counter\ntest1{shardID=\"0\"} 2\n\n# TYPE test1 counter\ntest1{shardID=\"0\"} 0\n\n# TYPE test2 counter\ntest2{shardID=\"2\"} 100\n\n# TYPE test2 counter\ntest2{shardID=\"2\"} 1\n\n# TYPE test2 counter\ntest2{shardID=\"2\"} 1\n\n# TYPE test2 counter\ntest2{shardID=\"2\"} 0\n\n", prometheusMetrics)
}
