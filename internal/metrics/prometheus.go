package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Регистрация метрик
var (
	TotalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Общее количество HTTP-запросов",
		},
		[]string{"method", "route", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Время выполнения HTTP-запросов",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
)

// InitMetrics инициализирует метрики
func InitMetrics() {
	prometheus.MustRegister(TotalRequests)
	prometheus.MustRegister(RequestDuration)
}

// Middleware для сбора метрик
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Обертка для записи статуса ответа
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)

		// Обновление метрик
		duration := time.Since(start).Seconds()
		route := r.URL.Path
		method := r.Method
		status := ww.statusCode

		TotalRequests.WithLabelValues(method, route, http.StatusText(status)).Inc()
		RequestDuration.WithLabelValues(method, route).Observe(duration)
	})
}

// responseWriter используется для захвата статуса ответа
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Экспорт метрик
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
