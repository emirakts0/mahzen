package domain

import (
	"context"
	"io"
	"time"
)

// ObjectStorage defines operations for storing and retrieving large objects.
type ObjectStorage interface {
	Upload(ctx context.Context, key string, reader io.Reader, contentType string, size int64) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}
