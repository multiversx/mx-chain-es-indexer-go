package prometheus

type prometheusDisabled struct{}

func (pc *prometheusDisabled) RegisterMetric(_ string, _ string, _ float64) {
	return
}
