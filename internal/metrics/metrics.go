package metrics

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	EventReceived = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "flox_events_received_total",
		Help: "Total number of events received",
	})

	EventFiltered = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "flox_events_filtered_total",
		Help: "Total number of events filtered",
	})

	OutputSuccess = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "flox_output_success_total",
		Help: "Total number of successful outputs",
	})

	OutputFailure = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "flox_output_failure_total",
		Help: "Total number of failed outputs",
	})
)

func InitMetricsServer() {
	prometheus.MustRegister(EventReceived, EventFiltered, OutputSuccess, OutputFailure)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Println("Prometheus metrics exposed at :2112/metrics")
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Fatalf("Error starting metrics server: %v", err)
		}
	}()
}
