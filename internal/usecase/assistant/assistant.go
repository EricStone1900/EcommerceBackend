package assistant

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// mockAssistant implements port.AssistantPort.
// It returns a fixed description after a simulated delay, mimicking a remote AI call.
type mockAssistant struct {
	logger *zap.Logger
}

// NewMockAssistant creates a new mock assistant for development/testing.
//
// FUTURE: Replace this implementation with a real AI microservice call.
// The port.AssistantPort interface does NOT need to change — the remote call
// will be an HTTP or gRPC invocation that respects context deadline/timeout.
func NewMockAssistant(logger *zap.Logger) *mockAssistant {
	return &mockAssistant{logger: logger}
}

// GenerateProductDescription returns a mock product description.
// It simulates a 50ms latency to validate context timeout handling.
func (m *mockAssistant) GenerateProductDescription(ctx context.Context, productID uint) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(50 * time.Millisecond):
	}

	m.logger.Info("mock assistant: generating description",
		zap.Uint("product_id", productID),
	)

	return fmt.Sprintf("AI generated description for product %d. This is a high-quality item with excellent features and great value for money.", productID), nil
}
