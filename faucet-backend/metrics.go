package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const defaultMetricsPullAddr = "0.0.0.0:7000"

var (
	// Labels to use for partitioning requests.
	requestLabels = []string{"endpoint", "status"}

	// Labels to use for partitioning request latencies.
	requestLatencyLabels = []string{"endpoint"}

	// Labels to use for partitioning balances.
	balanceLabels = []string{"network"}
)

type FaucetMetrics struct {
	// Counts of funding requests.
	Requests *prometheus.CounterVec

	// Latencies of requests.
	RequestLatencies *prometheus.SummaryVec

	// Current faucet balances.
	Balances *prometheus.GaugeVec
}

func NewDefaultFaucetMetrics() *FaucetMetrics {
	metrics := FaucetMetrics{
		Requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("faucet_requests"),
				Help: fmt.Sprintf("How many requests were received, partitioned by endpoint and status"),
			},
			requestLabels,
		),
		RequestLatencies: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: fmt.Sprintf("faucet_request_durations"),
				Help: fmt.Sprintf("How long requests take to process, partitioned by endpoint"),
			},
			requestLatencyLabels,
		),
		Balances: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("faucet_balances"),
				Help: fmt.Sprintf("Balances of faucet funds, partitioned by paratime"),
			},
			balanceLabels,
		),
	}
	prometheus.MustRegister(metrics.Requests)
	prometheus.MustRegister(metrics.RequestLatencies)
	prometheus.MustRegister(metrics.Balances)
	return &metrics
}

func (svc *Service) MetricsWorker() {
	svc.log.Printf("metrics: started")
	addr := svc.cfg.MetricsPullAddr
	if addr == "" {
		addr = defaultMetricsPullAddr
	}

	metricsServer := &http.Server{
		Addr:           addr,
		Handler:        promhttp.Handler(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	for {
		if err := metricsServer.ListenAndServe(); err != nil {
			svc.log.Printf("metrics: error serving request %v", err)
		}
	}
}
