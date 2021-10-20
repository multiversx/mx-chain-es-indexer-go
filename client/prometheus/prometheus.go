package prometheus

import (
	"net/http"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var log = logger.GetOrCreate("indexer/client/prometheus")

const prometheusEndpoint = "/indexer-metrics"

type prometheusClient struct {
	histogramVec *prometheus.HistogramVec
	server       *http.Server
}

func newPrometheusHandler(prometheusAPIInterface string) (*prometheusClient, error) {
	histogramVec := prometheus.NewHistogramVec(
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

	err := prometheus.Register(histogramVec)
	if err != nil {
		return nil, err
	}
	http.Handle(prometheusEndpoint, promhttp.Handler())

	promClient := &prometheusClient{
		histogramVec: histogramVec,
		server:       &http.Server{Addr: prometheusAPIInterface},
	}

	err = promClient.startServer()
	if err != nil {
		return nil, err
	}

	return promClient, nil
}

func (pc *prometheusClient) startServer() error {
	c := make(chan error)
	go func(c chan error) {
		errStart := pc.server.ListenAndServe()
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

// RegisterMetric will register the provided values in a histogram vec
func (pc *prometheusClient) RegisterMetric(id string, op string, value float64) {
	pc.histogramVec.WithLabelValues(id, op).Observe(value)
}

// Close will close the http server
func (pc *prometheusClient) Close() error {
	return pc.server.Close()
}
