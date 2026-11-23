package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var requestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "hotel_http_requests_total",
		Help: "Total HTTP requests handled by hotelsvc.",
	},
	[]string{"method", "route", "status"},
)

var requestDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "hotel_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "route"},
)

type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func HTTP(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			mux.ServeHTTP(w, r)
			return
		}

		_, route := mux.Handler(r)
		if route == "" {
			route = "unmatched"
		}

		recorder := &statusWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		start := time.Now()
		mux.ServeHTTP(recorder, r)
		duration := time.Since(start).Seconds()

		requestsTotal.WithLabelValues(
			r.Method,
			route,
			strconv.Itoa(recorder.status),
		).Inc()

		requestDuration.WithLabelValues(
			r.Method,
			route,
		).Observe(duration)
	})
}
