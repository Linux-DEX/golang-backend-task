package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Simple metrics for basic monitoring
var (
	JobsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "edi_jobs_total",
			Help: "Total number of EDI jobs processed",
		},
		[]string{"status"},
	)
)

func RecordJobCreated() {
	JobsTotal.WithLabelValues("created").Inc()
}

func RecordJobCompleted() {
	JobsTotal.WithLabelValues("completed").Inc()
}

func RecordJobFailed() {
	JobsTotal.WithLabelValues("failed").Inc()
}
