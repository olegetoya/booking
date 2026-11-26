package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/olegetoya/booking/notificationsvc/internal/handler"
	"github.com/olegetoya/booking/notificationsvc/internal/metrics"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type MessageHandler interface {
	Handle(ctx context.Context, key, value []byte) error
}

type Consumer struct {
	client  *kgo.Client
	handler MessageHandler
	log     *slog.Logger
}

func NewConsumer(
	brokers []string,
	topic string,
	groupID string,
	handler MessageHandler,
	log *slog.Logger,
) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumeTopics(topic),
		kgo.ConsumerGroup(groupID),
		kgo.DisableAutoCommit(),
		kgo.BlockRebalanceOnPoll(),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka client: %w", err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping kafka: %w", err)
	}

	return &Consumer{
		client:  client,
		handler: handler,
		log:     log,
	}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return nil
		}

		fetches := c.client.PollFetches(ctx)

		if fetches.IsClientClosed() {
			return nil
		}

		if ctx.Err() != nil {
			c.client.AllowRebalance()
			return nil
		}

		var fetchErr error
		fetches.EachError(func(topic string, partition int32, err error) {
			c.log.Error("kafka fetch error",
				slog.String("topic", topic),
				slog.Int("partition", int(partition)),
				slog.Any("error", err),
			)
			fetchErr = errors.Join(fetchErr, err)
			metrics.FetchError(topic)
		})

		if fetchErr != nil {
			c.client.AllowRebalance()
			return fmt.Errorf("fetch messages: %w", fetchErr)
		}

		var handleErr error

		fetches.EachRecord(func(record *kgo.Record) {
			if handleErr != nil {
				return
			}

			handleErr = c.processRecord(ctx, record)
		})

		if handleErr != nil {
			c.client.AllowRebalance()
			return fmt.Errorf("handle kafka message: %w", handleErr)
		}

		if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
			metrics.Commit("error")
			c.client.AllowRebalance()
			return fmt.Errorf("commit kafka offsets: %w", err)
		}

		metrics.Commit("success")

		c.client.AllowRebalance()
	}
}

func (c *Consumer) Close() {
	c.client.Close()
}

func (c *Consumer) processRecord(ctx context.Context, record *kgo.Record) error {
	carrier := propagation.MapCarrier{}
	for _, header := range record.Headers {
		if header.Key == "traceparent" || header.Key == "tracestate" {
			carrier.Set(header.Key, string(header.Value))
		}
	}
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	ctx, span := otel.Tracer("notificationsvc/kafka").Start(ctx, "process "+record.Topic,
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination.name", record.Topic),
			attribute.String("messaging.operation.type", "process"),
			attribute.Int64("messaging.kafka.offset", record.Offset),
			attribute.Int("messaging.destination.partition.id", int(record.Partition)),
		),
	)
	defer span.End()

	start := time.Now()

	err := c.handler.Handle(
		ctx,
		record.Key,
		record.Value,
	)

	result := "success"

	if errors.Is(err, handler.ErrInvalidEvent) {
		result = "invalid"
	} else if err != nil {
		result = "error"
	}

	span.SetAttributes(attribute.String("messaging.process.result", result))
	metrics.ObserveMessage(
		record.Topic,
		result,
		time.Since(start),
	)

	if err == nil {
		c.log.InfoContext(ctx, "processed kafka event", slog.String("topic", record.Topic), slog.Int("partition", int(record.Partition)), slog.Int64("offset", record.Offset))
		return nil
	}

	if errors.Is(err, handler.ErrInvalidEvent) {
		c.log.WarnContext(ctx,
			"skipping invalid kafka event",
			slog.String("topic", record.Topic),
			slog.Int("partition", int(record.Partition)),
			slog.Int64("offset", record.Offset),
			slog.Any("error", err),
		)

		return nil
	}

	span.RecordError(errors.New("event processing failed"))
	span.SetStatus(codes.Error, "event processing failed")
	c.log.ErrorContext(ctx, "kafka event processing failed", slog.Any("error", err))
	return err
}
