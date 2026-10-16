package middleware

import (
	"runtime/debug"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	startTime = time.Now()

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_server_requests_seconds",
			Help: "Histogram of HTTP request durations in seconds",
			ConstLabels: prometheus.Labels{
				"runtime": "go",
			},
			Buckets: []float64{
				0.005,
				0.01,
				0.025,
				0.05,
				0.075,
				0.1,
				0.25,
				0.5,
				0.75,
				1.0,
				2.5,
				5.0,
				10.0,
			},
		},
		[]string{"method", "status", "uri"},
	)
)

func InitMetrics(pool *pgxpool.Pool) {
	prometheus.MustRegister(RequestDuration)

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "hikaricp_connections_active",
			Help: "Active DB connections",
			ConstLabels: prometheus.Labels{
				"runtime": "go",
			},
		},
		func() float64 {
			if pool != nil {
				return float64(pool.Stat().AcquiredConns())
			}
			return 0
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "hikaricp_connections_pending",
			Help: "Pending DB connections",
			ConstLabels: prometheus.Labels{
				"runtime": "go",
			},
		},
		func() float64 {
			if pool != nil {
				return float64(pool.Stat().ConstructingConns())
			}
			return 0
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "process_uptime_seconds",
			Help: "Process uptime in seconds",
			ConstLabels: prometheus.Labels{
				"runtime": "go",
			},
		},
		func() float64 {
			return time.Since(startTime).Seconds()
		},
	))

	var gcStats debug.GCStats
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "jvm_gc_pause_seconds_sum",
			Help: "Total GC pause time in seconds for Go runtime",
			ConstLabels: prometheus.Labels{
				"runtime": "go",
			},
		},
		func() float64 {
			debug.ReadGCStats(&gcStats)
			return gcStats.PauseTotal.Seconds()
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "jvm_gc_pause_seconds_count",
			Help: "Total GC pause count for Go runtime",
			ConstLabels: prometheus.Labels{
				"runtime": "go",
			},
		},
		func() float64 {
			debug.ReadGCStats(&gcStats)
			return float64(gcStats.NumGC)
		},
	))
}

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()

		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		uri := c.FullPath()
		if uri == "" {
			uri = "UNKNOWN"
		}

		RequestDuration.
			WithLabelValues(method, status, uri).
			Observe(duration)
	}
}
