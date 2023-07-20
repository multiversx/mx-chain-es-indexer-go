package metrics

import (
	"bytes"
	"fmt"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"google.golang.org/protobuf/proto"
)

func counterMetric(metricName string, shardID uint32, count uint64) string {
	metricFamily := &dto.MetricFamily{
		Name: proto.String(metricName),
		Type: dto.MetricType_COUNTER.Enum(),
		Metric: []*dto.Metric{
			{
				Label: []*dto.LabelPair{
					{
						Name:  proto.String("shardID"),
						Value: proto.String(fmt.Sprintf("%d", shardID)),
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
