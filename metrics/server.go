package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/prometheus/otlptranslator"
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
	sentPackets     metric.Int64Counter
	receivedPackets metric.Int64Counter
	lostPackets     metric.Int64Counter
	receivedDups    metric.Int64Counter
	rtt             metric.Float64Histogram
}

// NewServer returns a new metric server
func NewServer(address string) (*Server, error) {
	srv := &Server{address: address}
	if err := srv.init(); err != nil {
		return nil, err
	}
	return srv, nil
}

// Listen listens for incoming metric scrape requests
func (srv *Server) Listen() error {
	http.Handle("/metrics", promhttp.Handler())
	logrus.Infof("Listen for metrics requests on %s", srv.address)
	return http.ListenAndServe(srv.address, nil)
}

func (srv *Server) init() error {
	exporter, err := prometheus.New(prometheus.WithTranslationStrategy(otlptranslator.UnderscoreEscapingWithSuffixes))
	if err != nil {
		return fmt.Errorf("failed to initialize prometheus exporter: %w", err)
	}

	const ns = "kube_cni_probe"
	rttView := sdkmetric.NewView(
		sdkmetric.Instrument{Name: ns + "/rtt"},
		sdkmetric.Stream{
			Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: []float64{0.1, 0.2, 0.5, 1.0, 2.5, 5.0, 10.0, 20.0},
			},
		},
	)
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithView(rttView),
	)
	otel.SetMeterProvider(provider)

	meter := otel.Meter(ns)

	if srv.sentPackets, err = meter.Int64Counter(ns+"/sent_packet_count",
		metric.WithDescription("Echo packets sent since start")); err != nil {
		return fmt.Errorf("failed to create sent_packet_count: %w", err)
	}
	if srv.receivedPackets, err = meter.Int64Counter(ns+"/received_packet_count",
		metric.WithDescription("Echo responses received since start")); err != nil {
		return fmt.Errorf("failed to create received_packet_count: %w", err)
	}
	if srv.lostPackets, err = meter.Int64Counter(ns+"/lost_packet_count",
		metric.WithDescription("Lost echo packets requests since start")); err != nil {
		return fmt.Errorf("failed to create lost_packet_count: %w", err)
	}
	if srv.receivedDups, err = meter.Int64Counter(ns+"/received_duplicate_packet_count",
		metric.WithDescription("Duplicate responses received since start")); err != nil {
		return fmt.Errorf("failed to create received_duplicate_packet_count: %w", err)
	}
	if srv.rtt, err = meter.Float64Histogram(ns+"/rtt",
		metric.WithDescription("Roundtrip time")); err != nil {
		return fmt.Errorf("failed to create rtt histogram: %w", err)
	}

	return nil
}

// ReportSentPackets increases the sent packet count counter
func (srv *Server) ReportSentPackets(ctx context.Context, node string, value int64) {
	srv.sentPackets.Add(ctx, value, metric.WithAttributes(attribute.String("destination", node)))
}

// ReportReceivedPackets increases the received packet count counter
func (srv *Server) ReportReceivedPackets(ctx context.Context, node string, value int64) {
	srv.receivedPackets.Add(ctx, value, metric.WithAttributes(attribute.String("destination", node)))
}

// ReportDupPackets increases the duplicate packet count counter
func (srv *Server) ReportDupPackets(ctx context.Context, node string, value int64) {
	srv.receivedDups.Add(ctx, value, metric.WithAttributes(attribute.String("destination", node)))
}

// ReportLostPackets increases the lost packet count counter
func (srv *Server) ReportLostPackets(ctx context.Context, node string, value int64) {
	srv.lostPackets.Add(ctx, value, metric.WithAttributes(attribute.String("destination", node)))
}

// ReportRTT reports a round trip time for populating the histogram
func (srv *Server) ReportRTT(ctx context.Context, node string, value float64) {
	srv.rtt.Record(ctx, value, metric.WithAttributes(attribute.String("destination", node)))
}
