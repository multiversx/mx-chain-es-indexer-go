package prometheus

type prometheusDisabled struct{}

// RegisterMetric will do nothing
func (pc *prometheusDisabled) RegisterMetric(_ string, _ string, _ float64) {
	return
}
