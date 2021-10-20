package prometheus

// PrometheusHandler defines what a prometheus handler should do
type PrometheusHandler interface {
	RegisterMetric(id string, op string, value float64)
}

// CreatePrometheusHandler will create a new instance of PrometheusHandler
func CreatePrometheusHandler(enabled bool, prometheusAPIInterface string) (PrometheusHandler, error) {
	if enabled {
		return newPrometheusHandler(prometheusAPIInterface)
	}

	return &prometheusDisabled{}, nil
}
