package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var kafkaProduceTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "booking_kafka_produce_total",
		Help: "Kafka publication attempts.",
	},
	[]string{"topic", "result"},
)

var kafkaProduceDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "booking_kafka_produce_duration_seconds",
		Help:    "Duration of Kafka publication attempts.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"topic"},
)

func ObserveKafkaProduce(topic, result string, duration time.Duration) {
	kafkaProduceTotal.WithLabelValues(topic, result).Inc()
	kafkaProduceDuration.WithLabelValues(topic).
		Observe(duration.Seconds())
}
