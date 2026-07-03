package assistant

import (
	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/port"
)

// newTestAssistant creates an assistant usecase with a logger for testing.
func newTestAssistant() *mockAssistant {
	logger, _ := zap.NewDevelopment()
	return &mockAssistant{logger: logger}
}

// ensureMockImplementsInterface verifies mockAssistant satisfies port.AssistantPort at compile time.
var _ port.AssistantPort = (*mockAssistant)(nil)
