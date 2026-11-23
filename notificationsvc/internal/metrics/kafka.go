package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var messagesTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "notification_kafka_messages_total",
		Help: "Kafka messages handled by outcome.",
	},
	[]string{"topic", "result"},
)

var handlerDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "notification_kafka_handler_duration_seconds",
		Help:    "Duration of Kafka message handling.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"topic"},
)

var fetchErrors = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "notification_kafka_fetch_errors_total",
		Help: "Kafka fetch errors.",
	},
	[]string{"topic"},
)

var commitsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "notification_kafka_commits_total",
		Help: "Kafka offset commit attempts.",
	},
	[]string{"result"},
)

func ObserveMessage(topic, result string, duration time.Duration) {
	messagesTotal.WithLabelValues(topic, result).Inc()
	handlerDuration.WithLabelValues(topic).Observe(duration.Seconds())
}

func FetchError(topic string) {
	fetchErrors.WithLabelValues(topic).Inc()
}

func Commit(result string) {
	commitsTotal.WithLabelValues(result).Inc()
}
