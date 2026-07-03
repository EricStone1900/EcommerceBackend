package event

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestEventBus_PublishSubscribe(t *testing.T) {
	logger := newTestLogger(t)
	bus := NewEventBus(logger)
	defer bus.Close()

	var (
		mu       sync.Mutex
		received []string
	)

	bus.Subscribe("test.event", func(ctx context.Context, payload any) error {
		mu.Lock()
		received = append(received, payload.(string))
		mu.Unlock()
		return nil
	})

	err := bus.Publish(context.Background(), "test.event", "hello")
	assert.NoError(t, err)

	// Wait for async handler
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, []string{"hello"}, received)
	mu.Unlock()
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	logger := newTestLogger(t)
	bus := NewEventBus(logger)
	defer bus.Close()

	var (
		mu sync.Mutex
		c1 int
		c2 int
	)

	bus.Subscribe("test.event", func(ctx context.Context, payload any) error {
		mu.Lock()
		c1++
		mu.Unlock()
		return nil
	})

	bus.Subscribe("test.event", func(ctx context.Context, payload any) error {
		mu.Lock()
		c2++
		mu.Unlock()
		return nil
	})

	_ = bus.Publish(context.Background(), "test.event", "msg")
	_ = bus.Publish(context.Background(), "test.event", "msg")

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 2, c1)
	assert.Equal(t, 2, c2)
	mu.Unlock()
}

func TestEventBus_NoSubscriber(t *testing.T) {
	logger := newTestLogger(t)
	bus := NewEventBus(logger)
	defer bus.Close()

	// Publishing to an event with no subscribers should not panic or error
	err := bus.Publish(context.Background(), "nonexistent", "data")
	assert.NoError(t, err)
}

func newTestLogger(t *testing.T) *zap.Logger {
	t.Helper()
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{"stdout"}
	logger, err := cfg.Build()
	if err != nil {
		t.Fatalf("failed to create test logger: %v", err)
	}
	return logger
}
