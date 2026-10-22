package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
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

			if err := c.handler.Handle(
				ctx,
				record.Key,
				record.Value,
			); err != nil {
				handleErr = err
			}
		})

		if handleErr != nil {
			c.client.AllowRebalance()
			return fmt.Errorf("handle kafka message: %w", handleErr)
		}

		if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
			c.client.AllowRebalance()
			return fmt.Errorf("commit kafka offsets: %w", err)
		}

		c.client.AllowRebalance()
	}
}

func (c *Consumer) Close() {
	c.client.Close()
}
