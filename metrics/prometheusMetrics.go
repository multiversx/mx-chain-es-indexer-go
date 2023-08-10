package metrics

import (
	"bytes"
	"strconv"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"google.golang.org/protobuf/proto"
)

const (
	operationName = "operation"
	shardIDName   = "shardID"
	errorCodeName = "errorCode"
)

func counterMetric(metricName, operation string, shardIDStr string, count uint64) string {
	metricFamily := &dto.MetricFamily{
		Name: proto.String(metricName),
		Type: dto.MetricType_COUNTER.Enum(),
		Metric: []*dto.Metric{
			{
				Label: []*dto.LabelPair{
					{
						Name:  proto.String(operationName),
						Value: proto.String(operation),
					},
					{
						Name:  proto.String(shardIDName),
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

func errorsMetric(metricName, operation string, shardIDStr string, errorsCount map[int]uint64) string {
	metricFamily := &dto.MetricFamily{
		Name:   proto.String(metricName),
		Type:   dto.MetricType_GAUGE.Enum(),
		Metric: make([]*dto.Metric, 0, len(errorsCount)),
	}

	for code, count := range errorsCount {
		m := &dto.Metric{
			Label: []*dto.LabelPair{
				{
					Name:  proto.String(operationName),
					Value: proto.String(operation),
				},
				{
					Name:  proto.String(shardIDName),
					Value: proto.String(shardIDStr),
				},
				{
					Name:  proto.String(errorCodeName),
					Value: proto.String(strconv.Itoa(code)),
				},
			},
			Gauge: &dto.Gauge{
				Value: proto.Float64(float64(count)),
			},
		}

		metricFamily.Metric = append(metricFamily.Metric, m)
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
