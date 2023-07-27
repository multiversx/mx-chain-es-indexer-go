package metrics

import (
	"bytes"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"google.golang.org/protobuf/proto"
)

func counterMetric(metricName, operation string, shardIDStr string, count uint64) string {
	metricFamily := &dto.MetricFamily{
		Name: proto.String(metricName),
		Type: dto.MetricType_COUNTER.Enum(),
		Metric: []*dto.Metric{
			{
				Label: []*dto.LabelPair{
					{
						Name:  proto.String("operation"),
						Value: proto.String(operation),
					},
					{
						Name:  proto.String("shardID"),
						Value: proto.String(shardIDStr),
					},
				},
				Counter: &dto.Counter{
					Value: proto.Float64(float64(count)),
				},
			},
		},
	}

	return promMetricAsString(metricFamily)
}

func promMetricAsString(metric *dto.MetricFamily) string {
	out := bytes.NewBuffer(make([]byte, 0))
	_, err := expfmt.MetricFamilyToText(out, metric)
	if err != nil {
		return ""
	}

	return out.String() + "\n"
}
