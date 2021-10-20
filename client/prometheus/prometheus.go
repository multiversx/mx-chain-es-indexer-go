package prometheus

import (
	"fmt"
	"net/http"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var log = logger.GetOrCreate("indexer/client/prometheus")

type prometheusClient struct {
	histogramVec *prometheus.HistogramVec
}

func newPrometheusHandler(prometheusAPIInterface string) (*prometheusClient, error) {
	gaugeVec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "metrics",
			Subsystem: "elastic_indexer",
			Name:      "ops",
			Help:      "Metrics",
			Buckets:   []float64{0},
		},
		[]string{
			"id",
			"operation",
		},
	)

	err := prometheus.Register(gaugeVec)
	if err != nil {
		return nil, err
	}
	http.Handle("/indexer-metrics", promhttp.Handler())

	err = startServer(prometheusAPIInterface)
	if err != nil {
		return nil, err
	}

	return &prometheusClient{
		histogramVec: gaugeVec,
	}, nil
}

func startServer(prometheusAPIInterface string) error {
	c := make(chan error)
	go func(c chan error) {
		errStart := http.ListenAndServe(fmt.Sprintf("%s", prometheusAPIInterface), nil)
		if errStart != nil {
			log.Warn("newPrometheusHandler cannot start prometheus http server", "error", errStart.Error())
			c <- errStart
		}
	}(c)

	for {
		select {
		case <-time.After(time.Second):
			return nil

		case err := <-c:
			return err
		}
	}
}

func (pc *prometheusClient) RegisterMetric(id string, op string, value float64) {
	pc.histogramVec.WithLabelValues(id, op).Observe(value)
}
