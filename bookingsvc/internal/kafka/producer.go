package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olegetoya/booking/bookingsvc/internal/config"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	client *kgo.Client
	topic  string
}

func NewProducer(cfg *config.KafkaConfig) (*Producer, error) {

	const op = "kafka.NewProducer"

	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: create kafka client: %w", op, err)
	}

	return &Producer{
		client: client,
		topic:  cfg.Topic,
	}, nil
}

func (p *Producer) Produce(
	ctx context.Context,
	key string,
	value any,
) error {
	const op = "kafka.Produce"
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

	return nil
}

func (p *Producer) Close() {
	p.client.Close()
}
