package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/olegetoya/booking/bookingsvc/internal/metrics"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
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
) (resultErr error) {
	const op = "kafka.Produce"

	ctx, span := otel.Tracer("bookingsvc/kafka").Start(ctx, "publish "+p.topic,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(attribute.String("messaging.system", "kafka"), attribute.String("messaging.destination.name", p.topic), attribute.String("messaging.operation.type", "send")),
	)
	defer func() {
		if resultErr != nil {
			span.RecordError(errors.New("kafka publication failed"))
			span.SetStatus(codes.Error, "kafka publication failed")
		}
		span.End()
	}()

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

	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	for key, value := range carrier {
		record.Headers = append(record.Headers, kgo.RecordHeader{Key: key, Value: []byte(value)})
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
