package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var requestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "booking_http_requests_total",
		Help: "Total HTTP requests handled by bookingsvc.",
	},
	[]string{"method", "route", "status"},
)

var requestDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "booking_http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "route"},
)

func HTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Не учитываем запросы самого Prometheus.
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		status := ww.Status()
		if status == 0 {
			status = http.StatusOK
		}

		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unmatched"
		}

		requestsTotal.WithLabelValues(
			r.Method,
			route,
			strconv.Itoa(status),
		).Inc()

		requestDuration.WithLabelValues(
			r.Method,
			route,
		).Observe(time.Since(start).Seconds())
	})
}
