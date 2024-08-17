package background

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/craigpastro/pgmq-go"
)

type QueueProducer[T any] struct {
	q         *pgmq.PGMQ
	queueName string
}

func NewQueueProducer[T any](ctx context.Context, queueName string, connString string) *QueueProducer[T] {
	q, err := pgmq.New(ctx, connString)
	if err != nil {
		panic(err)
	}
	if err := q.CreateQueue(ctx, queueName); err != nil {
		panic(err)
	}
	return &QueueProducer[T]{
		q:         q,
		queueName: queueName,
	}
}

func (q *QueueProducer[T]) Send(ctx context.Context, message T) error {
	bs, err := json.Marshal(message)
	if err != nil {
		return err
	}
	_, err = q.q.Send(ctx, q.queueName, bs)
	return err
}

func (q *QueueProducer[T]) SendWithDelay(ctx context.Context, message T, delay time.Duration) error {
	bs, err := json.Marshal(message)
	if err != nil {
		return err
	}
	_, err = q.q.SendWithDelay(ctx, q.queueName, bs, int(delay.Seconds()))
	return err
}

type QueueConsumer[T any] struct {
	q         *pgmq.PGMQ
	queueName string
	vt        int64
	handler   func(context.Context, T) error
	stopChan  chan struct{}
	logger    *slog.Logger
}

// NewQueueConsumer creates a new read loop on the queue, with a visibility timeout in seconds and a handler function
func NewQueueConsumer[T any](ctx context.Context, queueName string, connString string, vt int64, handler func(context.Context, T) error, logger *slog.Logger) *QueueConsumer[T] {
	q, err := pgmq.New(ctx, connString)
	if err != nil {
		panic(err)
	}
	return &QueueConsumer[T]{
		q:         q,
		queueName: queueName,
		vt:        vt,
		handler:   handler,
		stopChan:  make(chan struct{}),
		logger:    logger,
	}
}

func (c *QueueConsumer[T]) Start(ctx context.Context) {
	go func() {
		timer := time.NewTimer(0)
		defer timer.Stop()
		for {
			select {
			case <-c.stopChan:
				return
			case <-timer.C:
				msg, err := c.q.Read(ctx, c.queueName, c.vt)
				if err != nil {
					if err.Error() == "pgmq: no rows in result set" {
						timer.Reset(5 * time.Second)
						continue
					}
					c.logger.Error("error reading message", slog.Any("error", err))
					continue
				}
				// Reset timer for immediate next read
				timer.Reset(0)

				var decodedMsg T
				if err := json.Unmarshal(msg.Message, &decodedMsg); err != nil {
					c.logger.Error("error decoding message", slog.Any("error", err))
					continue
				}

				// after the visibilty timeout, other consumers will be able to read the message,
				// so give the handler a deadline
				ctx, cancel := context.WithTimeout(ctx, time.Duration(c.vt)*time.Second)
				done := make(chan error, 1)
				go func() {
					done <- c.handler(ctx, decodedMsg)
				}()
				select {
				case <-ctx.Done():
					c.logger.Error("visibility timeout deadline exceeded before handler returned")
					cancel()
					continue
				case err := <-done:
					if err != nil {
						c.logger.Error("error handling message", slog.Any("error", err))
						cancel()
						continue
					}
				}
				cancel()
				_, err = c.q.Archive(context.Background(), c.queueName, msg.MsgID)
				if err != nil {
					c.logger.Error("error deleting message", slog.Any("error", err))
				}
			}
		}
	}()
}

func (c *QueueConsumer[T]) Stop() {
	close(c.stopChan)
}
