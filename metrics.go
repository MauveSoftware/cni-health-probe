package main

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/unit"
)

type metrics struct {
	address         string
	prom            *prometheus.Exporter
	sentPackets     metric.Int64Counter
	receivedPackets metric.Int64Counter
	lostPackets     metric.Int64Counter
	receivedDups    metric.Int64Counter
	rtt             metric.Float64ValueRecorder
}

func newMetrics(address string) (*metrics, error) {
	m := &metrics{address: address}
	err := m.init()
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (e *metrics) listen() error {
	http.HandleFunc("/metrics", e.prom.ServeHTTP)

	logrus.Infof("Listen for metrics requests on %s", e.address)
	return http.ListenAndServe(e.address, nil)
}

func (e *metrics) init() error {
	exporter, err := prometheus.InstallNewPipeline(prometheus.Config{
		DefaultHistogramBoundaries: []float64{0.1, 0.2, 0.5, 1.0, 2.5, 5.0, 10.0, 20.0},
	})
	if err != nil {
		return errors.Wrap(err, "failed to initialize prometheus exporter")
	}
	e.prom = exporter

	meter := metric.Must(otel.Meter("node_cni_probe"))
	e.sentPackets = meter.NewInt64Counter("sent_packet_count", metric.WithDescription("Echo packets sent since start"))
	e.receivedPackets = meter.NewInt64Counter("received_packet_count", metric.WithDescription("Echo responses received since start"))
	e.lostPackets = meter.NewInt64Counter("lost_packet_count", metric.WithDescription("Lost echo packets requests since start"))
	e.receivedDups = meter.NewInt64Counter("received_duplicate_packet_count", metric.WithDescription("Duplicate responses received since start"))
	e.rtt = meter.NewFloat64ValueRecorder("rtt", metric.WithDescription("Roundtrip time"), metric.WithUnit(unit.Milliseconds))

	return nil
}
