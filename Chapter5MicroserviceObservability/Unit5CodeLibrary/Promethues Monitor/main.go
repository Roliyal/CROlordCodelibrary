package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "demo_http_requests_total",
			Help: "Total HTTP requests.",
		},
		[]string{"path", "method", "code"},
	)

	httpLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "demo_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	buildInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "demo_build_info",
			Help: "Build info (value is always 1).",
		},
		[]string{"version"},
	)
)

func main() {
	rand.Seed(time.Now().UnixNano())

	prometheus.MustRegister(httpRequests, httpLatency, buildInfo)
	buildInfo.WithLabelValues("v1.0.0").Set(1)

	mux := http.NewServeMux()

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// 模拟一点延迟
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

		httpRequests.WithLabelValues("/hello", r.Method, "200").Inc()
		httpLatency.WithLabelValues("/hello", r.Method).Observe(time.Since(start).Seconds())
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello\n"))
	})

	// metrics 暴露点
	mux.Handle("/metrics", promhttp.Handler())

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
