package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var grpcServerRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "hotel_grpc_server_requests_total",
		Help: "Total incoming gRPC calls.",
	},
	[]string{"method", "code"},
)

var grpcServerDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "hotel_grpc_server_duration_seconds",
		Help:    "Duration of incoming gRPC calls.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method"},
)

func UnaryServerInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	code := status.Code(err).String()

	grpcServerRequests.WithLabelValues(
		info.FullMethod, code,
	).Inc()

	grpcServerDuration.WithLabelValues(
		info.FullMethod,
	).Observe(time.Since(start).Seconds())

	return resp, err
}
