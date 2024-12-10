package middleware

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "loan_service_http_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path", "method", "status"})

	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "loan_service_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"path", "method", "status"})

	activeRequests = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "loan_service_active_requests",
		Help: "Number of active requests.",
	})
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		activeRequests.Inc()

		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r)

		activeRequests.Dec()
		duration := time.Since(start).Seconds()

		status := wrapped.status
		if status == 0 {
			status = 200
		}

		statusStr := http.StatusText(status)
		httpDuration.WithLabelValues(r.URL.Path, r.Method, statusStr).Observe(duration)
		requestsTotal.WithLabelValues(r.URL.Path, r.Method, statusStr).Inc()
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
