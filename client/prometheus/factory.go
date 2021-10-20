package prometheus

// PrometheusHandler -
type PrometheusHandler interface {
	RegisterMetric(id string, op string, value float64)
}

func CreatePrometheusHandler(enabled bool, prometheusAPIInterface string) (PrometheusHandler, error) {
	if enabled {
		return newPrometheusHandler(prometheusAPIInterface)
	}

	return &prometheusDisabled{}, nil
}
