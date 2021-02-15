package metrics

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/unit"
)

// Reporter allows reporting of metrics
type Reporter interface {
	ReportSentPackets(ctx context.Context, node string, value int64)
	ReportReceivedPackets(ctx context.Context, node string, value int64)
	ReportDupPackets(ctx context.Context, node string, value int64)
	ReportLostPackets(ctx context.Context, node string, value int64)
	ReportRTT(ctx context.Context, node string, value float64)
}

// Server exposes metrics and allows to report them
type Server struct {
	address         string
	prom            *prometheus.Exporter
	sentPackets     metric.Int64Counter
	receivedPackets metric.Int64Counter
	lostPackets     metric.Int64Counter
	receivedDups    metric.Int64Counter
	rtt             metric.Float64ValueRecorder
}

// NewServer returns a new metric server
func NewServer(address string) (*Server, error) {
	srv := &Server{address: address}
	err := srv.init()
	if err != nil {
		return nil, err
	}

	return srv, nil
}

// Listen listens for incoming metric scrape requests
func (srv *Server) Listen() error {
	http.HandleFunc("/metrics", srv.prom.ServeHTTP)

	logrus.Infof("Listen for metrics requests on %s", srv.address)
	return http.ListenAndServe(srv.address, nil)
}

func (srv *Server) init() error {
	exporter, err := prometheus.InstallNewPipeline(prometheus.Config{
		DefaultHistogramBoundaries: []float64{0.1, 0.2, 0.5, 1.0, 2.5, 5.0, 10.0, 20.0},
	})
	if err != nil {
		return errors.Wrap(err, "failed to initialize prometheus exporter")
	}
	srv.prom = exporter

	const ns = "node_cni_probe"
	meter := otel.Meter(ns)
	srv.sentPackets = metric.Must(meter).NewInt64Counter(ns+"/sent_packet_count", metric.WithDescription("Echo packets sent since start"))
	srv.receivedPackets = metric.Must(meter).NewInt64Counter(ns+"/received_packet_count", metric.WithDescription("Echo responses received since start"))
	srv.lostPackets = metric.Must(meter).NewInt64Counter(ns+"/lost_packet_count", metric.WithDescription("Lost echo packets requests since start"))
	srv.receivedDups = metric.Must(meter).NewInt64Counter(ns+"/received_duplicate_packet_count", metric.WithDescription("Duplicate responses received since start"))
	srv.rtt = metric.Must(meter).NewFloat64ValueRecorder(ns+"/rtt", metric.WithDescription("Roundtrip time"), metric.WithUnit(unit.Milliseconds))

	return nil
}

// ReportSentPackets increases the sent packet count counter
func (srv *Server) ReportSentPackets(ctx context.Context, node string, value int64) {
	srv.reportCounter(ctx, srv.sentPackets, node, value)
}

// ReportReceivedPackets increases the received packet count counter
func (srv *Server) ReportReceivedPackets(ctx context.Context, node string, value int64) {
	srv.reportCounter(ctx, srv.receivedPackets, node, value)
}

// ReportDupPackets increases the duplicate packet count counter
func (srv *Server) ReportDupPackets(ctx context.Context, node string, value int64) {
	srv.reportCounter(ctx, srv.receivedDups, node, value)
}

// ReportLostPackets increases the lost packet count counter
func (srv *Server) ReportLostPackets(ctx context.Context, node string, value int64) {
	srv.reportCounter(ctx, srv.lostPackets, node, value)
}

// ReportRTT reports a round trip time for populating the histogram
func (srv *Server) ReportRTT(ctx context.Context, node string, value float64) {
	srv.rtt.Record(ctx, value, label.String("destination", node))
}

func (srv *Server) reportCounter(ctx context.Context, counter metric.Int64Counter, node string, value int64) {
	counter.Add(ctx, value, label.String("destination", node))
}
