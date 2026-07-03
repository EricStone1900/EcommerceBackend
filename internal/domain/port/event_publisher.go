package port

import "context"

// EventPublisher defines the interface for event publishing and subscribing.
// The current in-process implementation can be replaced by NATS/RabbitMQ/etc.
// without changing any use case code.
type EventPublisher interface {
	// Publish sends an event with the given name and payload to all subscribers.
	// Implementations should handle errors internally (logging) and not block the caller.
	Publish(ctx context.Context, eventName string, payload any) error

	// Subscribe registers a handler for a specific event name.
	// The handler receives the event context and payload.
	Subscribe(eventName string, handler func(ctx context.Context, payload any) error)
}
