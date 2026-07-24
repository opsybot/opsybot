package blob

import (
	"context"
	"errors"
	"io"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/objectstore"
	"github.com/opsybot/opsybot/internal/repository"
)

type repo struct {
	client objectstore.Client
}

func New(client objectstore.Client) repository.Blob {
	return &repo{client: client}
}

func (r *repo) Enabled(ctx context.Context) bool {
	_ = ctx
	return r.client.Enabled()
}

func (r *repo) Put(ctx context.Context, key string, content io.Reader, size int64, contentType string) error {
	return unavailable(r.client.Put(ctx, key, content, size, contentType))
}

func (r *repo) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	body, err := r.client.Get(ctx, key)
	if err != nil {
		return nil, unavailable(err)
	}
	return body, nil
}

func (r *repo) Remove(ctx context.Context, key string) error {
	return unavailable(r.client.Remove(ctx, key))
}

func unavailable(err error) error {
	if errors.Is(err, objectstore.ErrDisabled) {
		return entity.ErrAttachmentStorageUnavailable
	}
	return err
}
