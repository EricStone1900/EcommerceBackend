package port

import "context"

// AssistantPort defines the interface for AI assistant operations.
// Implementations can be mock (local) or real AI microservice (remote).
// Method signatures include context.Context and return error to reflect
// potential network failures and timeouts.
type AssistantPort interface {
	// GenerateProductDescription generates an AI-powered description for a product.
	// This may be a remote call that could timeout or fail.
	GenerateProductDescription(ctx context.Context, productID uint) (string, error)
}
