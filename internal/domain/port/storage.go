package port

import "context"

// Storage defines the interface for file storage operations.
// Implementations can be local disk, S3, MinIO, etc.
// Method signatures include context.Context to reflect potential remote calls.
type Storage interface {
	// Upload stores the file data and returns a URL that can be used to access it.
	Upload(ctx context.Context, data []byte, filename string) (url string, err error)

	// Delete removes the file at the given URL from storage.
	Delete(ctx context.Context, url string) error
}
