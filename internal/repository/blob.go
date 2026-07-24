package repository

//go:generate go tool mockgen -source=blob.go -destination=./blob/blob_mock.go -package=blob

import (
	"context"
	"io"
)

type Blob interface {
	Enabled(ctx context.Context) bool
	Put(ctx context.Context, key string, content io.Reader, size int64, contentType string) error
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Remove(ctx context.Context, key string) error
}
