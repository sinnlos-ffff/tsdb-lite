package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	IngestTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tsdb_ingest_total",
		Help: "Total number of points ingested",
	})

	IngestLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "tsdb_ingest_latency_seconds",
		Help:    "Latency of point ingestion",
		Buckets: prometheus.DefBuckets,
	})

	CompactedChunksTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tsdb_chunks_compacted_total",
		Help: "Total compacted chunks",
	})
)

func InitMetrics() {
	prometheus.MustRegister(IngestTotal, IngestLatency, CompactedChunksTotal)
}
