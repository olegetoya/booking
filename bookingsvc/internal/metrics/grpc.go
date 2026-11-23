package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var grpcClientRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "booking_grpc_client_requests_total",
		Help: "Total outgoing gRPC calls.",
	},
	[]string{"method", "code"},
)

var grpcClientDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "booking_grpc_client_duration_seconds",
		Help:    "Duration of outgoing gRPC calls.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method"},
)

func UnaryClientInterceptor(
	ctx context.Context,
	method string,
	req, reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	start := time.Now()

	err := invoker(ctx, method, req, reply, cc, opts...)

	code := status.Code(err).String()

	grpcClientRequests.WithLabelValues(
		method, code,
	).Inc()

	grpcClientDuration.WithLabelValues(
		method,
	).Observe(time.Since(start).Seconds())

	return err
}
