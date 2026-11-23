package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olegetoya/booking/bookingsvc/internal/metrics"
	"github.com/twmb/franz-go/pkg/kgo"
	"time"
)

type Producer struct {
	client *kgo.Client
	topic  string
}

func NewProducer(brokers []string, topic string) (*Producer, error) {

	const op = "kafka.NewProducer"

	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.RecordDeliveryTimeout(5*time.Second),
		kgo.RecordRetries(3),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: create kafka client: %w", op, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping kafka: %w", err)
	}

	return &Producer{
		client: client,
		topic:  topic,
	}, nil
}

func (p *Producer) Produce(
	ctx context.Context,
	key string,
	value any,
) error {
	const op = "kafka.Produce"

	start := time.Now()
	result := "error"

	defer func() {
		metrics.ObserveKafkaProduce(
			p.topic,
			result,
			time.Since(start),
		)
	}()

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("%s: marshal kafka message: %w", op, err)
	}

	record := &kgo.Record{
		Topic: p.topic,
		Key:   []byte(key),
		Value: data,
	}

	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("%s: produce kafka message: %w", op, err)
	}

	result = "success"
	return nil
}

func (p *Producer) Close() {
	p.client.Close()
}
