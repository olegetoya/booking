package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

func AccessLog(logger *slog.Logger, next http.Handler) http.Handler {
	accessLogger := logger.With("component", "access_log")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(lw, r)

		accessLogger.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", lw.statusCode,
			"duration", time.Since(start),
		)
	})
}
