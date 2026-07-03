package event

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// EventBus is an in-process event publisher/subscriber implementation.
//
// It currently dispatches events to handlers in separate goroutines, simulating
// the fire-and-forget semantics of a real message queue.
//
// FUTURE: Replace this implementation with NATS/RabbitMQ/etc. The port.EventPublisher
// interface does NOT need to change — subscribers will be independent remote microservice
// processes instead of in-process goroutines.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]func(ctx context.Context, payload any) error
	logger      *zap.Logger
	wg          sync.WaitGroup
	closed      bool
	closeCh     chan struct{}
}

// NewEventBus creates a new in-process event bus.
func NewEventBus(logger *zap.Logger) *EventBus {
	return &EventBus{
		subscribers: make(map[string][]func(ctx context.Context, payload any) error),
		logger:      logger,
		closeCh:     make(chan struct{}),
	}
}

// Publish dispatches an event to all registered subscribers asynchronously.
// It returns immediately — handlers run in separate goroutines.
// Handlers receive a detached context (WithoutCancel) so they can complete
// even after the original request context is cancelled.
// Errors in handlers are logged but not propagated to the caller,
// matching the fire-and-forget semantics of a real message queue.
func (b *EventBus) Publish(ctx context.Context, eventName string, payload any) error {
	b.mu.RLock()
	handlers, ok := b.subscribers[eventName]
	b.mu.RUnlock()

	if !ok || len(handlers) == 0 {
		return nil
	}

	// Detach from request context so handlers can complete asynchronously
	handlerCtx := context.WithoutCancel(ctx)

	for _, handler := range handlers {
		select {
		case <-b.closeCh:
			return nil
		default:
		}

		b.wg.Add(1)
		go func(h func(ctx context.Context, payload any) error) {
			defer b.wg.Done()
			if err := h(handlerCtx, payload); err != nil {
				b.logger.Error("event handler error",
					zap.String("event", eventName),
					zap.Error(err),
				)
			}
		}(handler)
	}

	return nil
}

// Subscribe registers a handler for the given event name.
// Handlers are called for every matching Publish call.
func (b *EventBus) Subscribe(eventName string, handler func(ctx context.Context, payload any) error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventName] = append(b.subscribers[eventName], handler)

	b.logger.Debug("event subscriber registered",
		zap.String("event", eventName),
	)
}

// Close gracefully shuts down the event bus.
// It waits for all in-flight handlers to complete.
func (b *EventBus) Close() {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}
	b.closed = true
	close(b.closeCh)
	b.mu.Unlock()

	b.wg.Wait()
	b.logger.Info("event bus closed")
}
