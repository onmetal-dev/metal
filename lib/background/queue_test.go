package background

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"log/slog"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testConnString = "postgres://postgres:postgres@localhost:5433/metal_test?sslmode=disable"

type TestMessage struct {
	ID   int
	Name string
}

func TestQueueProducerAndConsumer(t *testing.T) {
	ctx := context.Background()

	t.Run("Basic send and receive", func(t *testing.T) {
		queueName := "test_queue_basic"
		producer := NewQueueProducer[TestMessage](ctx, queueName, testConnString)

		msg := TestMessage{ID: 1, Name: "Test"}
		err := producer.Send(ctx, msg)
		require.NoError(t, err)

		var receivedMsg TestMessage
		var mu sync.Mutex

		consumer := NewQueueConsumer(ctx, queueName, testConnString, 30, func(ctx context.Context, m TestMessage) error {
			mu.Lock()
			receivedMsg = m
			mu.Unlock()
			return nil
		}, slog.Default())

		consumer.Start(ctx)
		defer consumer.Stop()

		time.Sleep(1 * time.Second)
		mu.Lock()
		assert.Equal(t, msg, receivedMsg)
		mu.Unlock()
	})

	t.Run("Send with delay", func(t *testing.T) {
		queueName := "test_queue_delay"
		producer := NewQueueProducer[TestMessage](ctx, queueName, testConnString)

		msg := TestMessage{ID: 2, Name: "Delayed"}
		delay := 5 * time.Second
		err := producer.SendWithDelay(ctx, msg, delay)
		require.NoError(t, err)

		var receivedMsg TestMessage
		var receivedTime time.Time
		var mu sync.Mutex

		startTime := time.Now()
		consumer := NewQueueConsumer(ctx, queueName, testConnString, 30, func(ctx context.Context, m TestMessage) error {
			mu.Lock()
			receivedMsg = m
			receivedTime = time.Now()
			mu.Unlock()
			return nil
		}, slog.Default())

		consumer.Start(ctx)
		defer consumer.Stop()

		time.Sleep(6 * time.Second)
		mu.Lock()
		assert.Equal(t, msg, receivedMsg)
		assert.GreaterOrEqual(t, receivedTime.Sub(startTime), delay-500*time.Millisecond)
		assert.LessOrEqual(t, receivedTime.Sub(startTime), delay+500*time.Millisecond)
		mu.Unlock()
	})

	t.Run("Multiple messages with different types", func(t *testing.T) {
		queueName := "test_queue_multiple"
		producer := NewQueueProducer[interface{}](ctx, queueName, testConnString)

		msg1 := TestMessage{ID: 3, Name: "First"}
		msg2 := 42
		msg3 := "Hello, World!"

		require.NoError(t, producer.Send(ctx, msg1))
		require.NoError(t, producer.Send(ctx, msg2))
		require.NoError(t, producer.Send(ctx, msg3))

		var mu sync.Mutex
		receivedMsgs := make([]interface{}, 0, 3)
		consumer := NewQueueConsumer(ctx, queueName, testConnString, 30, func(ctx context.Context, m interface{}) error {
			mu.Lock()
			receivedMsgs = append(receivedMsgs, m)
			mu.Unlock()
			return nil
		}, slog.Default())

		consumer.Start(ctx)
		defer consumer.Stop()

		time.Sleep(500 * time.Millisecond) // Give some time for the consumer to process all messages

		mu.Lock()
		assert.Len(t, receivedMsgs, 3)
		for _, msg := range []interface{}{msg1, msg2, msg3} {
			msgJson, _ := json.Marshal(msg)
			assert.True(t, lo.ContainsBy(receivedMsgs, func(m interface{}) bool {
				mJson, _ := json.Marshal(m)
				return string(mJson) == string(msgJson)
			}))
		}
		mu.Unlock()
	})

	t.Run("Multiple producers and consumers", func(t *testing.T) {
		queueName := "test_queue_multi_prod_cons"
		producer1 := NewQueueProducer[TestMessage](ctx, queueName, testConnString)
		producer2 := NewQueueProducer[TestMessage](ctx, queueName, testConnString)

		numMessages := 50
		sentMessages := make(map[int]bool)

		// Send messages from both producers
		for i := 0; i < numMessages; i++ {
			msg := TestMessage{ID: i, Name: fmt.Sprintf("Message %d", i)}
			sentMessages[i] = true
			if i%2 == 0 {
				require.NoError(t, producer1.Send(ctx, msg))
			} else {
				require.NoError(t, producer2.Send(ctx, msg))
			}
		}

		receivedMessages := make(map[int]bool)
		var mu sync.Mutex

		consumerFunc := func(ctx context.Context, m TestMessage) error {
			mu.Lock()
			defer mu.Unlock()
			receivedMessages[m.ID] = true
			return nil
		}

		consumer1 := NewQueueConsumer(ctx, queueName, testConnString, 30, consumerFunc, slog.Default())
		consumer2 := NewQueueConsumer(ctx, queueName, testConnString, 30, consumerFunc, slog.Default())

		consumer1.Start(ctx)
		consumer2.Start(ctx)
		defer consumer1.Stop()
		defer consumer2.Stop()

		// Wait for messages to be processed
		time.Sleep(2 * time.Second)

		mu.Lock()
		assert.Equal(t, len(sentMessages), len(receivedMessages), "Number of received messages should match sent messages")
		for id := range sentMessages {
			assert.True(t, receivedMessages[id], fmt.Sprintf("Message with ID %d should have been received", id))
		}
		mu.Unlock()
	})

	t.Run("Visibility timeout and message redelivery", func(t *testing.T) {
		queueName := "test_queue_visibility_timeout"
		producer := NewQueueProducer[TestMessage](ctx, queueName, testConnString)

		msg := TestMessage{ID: 100, Name: "Visibility Test"}
		err := producer.Send(ctx, msg)
		require.NoError(t, err)

		visibilityTimeout := 2 // 2 seconds
		var wg sync.WaitGroup
		wg.Add(2)

		var firstConsumerProcessed atomic.Bool
		var secondConsumerProcessed atomic.Bool

		// First consumer that hangs
		consumer1 := NewQueueConsumer(ctx, queueName, testConnString, int64(visibilityTimeout), func(ctx context.Context, m TestMessage) error {
			defer wg.Done()
			firstConsumerProcessed.Store(true)
			// Simulate hanging by sleeping for longer than the visibility timeout
			time.Sleep(time.Duration(visibilityTimeout*3) * time.Second)
			return nil
		}, slog.Default())

		// Second consumer that should receive the message after the visibility timeout
		consumer2 := NewQueueConsumer(ctx, queueName, testConnString, int64(visibilityTimeout), func(ctx context.Context, m TestMessage) error {
			defer wg.Done()
			secondConsumerProcessed.Store(true)
			assert.Equal(t, msg, m)
			return nil
		}, slog.Default())

		consumer1.Start(ctx)
		consumer2.Start(ctx)
		defer consumer1.Stop()
		defer consumer2.Stop()

		// Wait for both consumers to finish processing
		wg.Wait()

		assert.True(t, firstConsumerProcessed.Load(), "First consumer should have processed the message")
		assert.True(t, secondConsumerProcessed.Load(), "Second consumer should have processed the message after visibility timeout")
	})
}
