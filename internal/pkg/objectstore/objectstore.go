package objectstore

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/opsybot/opsybot/internal/config"
)

var ErrDisabled = errors.New("objectstore disabled: no s3.endpoint configured")

type Client struct {
	*minio.Client
	bucket string
}

func New(cfg config.S3) (Client, func(), error) {
	noop := func() {}
	if cfg.Endpoint == "" || cfg.Bucket == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		return Client{}, noop, nil
	}

	api, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return Client{}, noop, fmt.Errorf("objectstore client: %w", err)
	}

	return Client{Client: api, bucket: cfg.Bucket}, noop, nil
}

func (c Client) Enabled() bool {
	return c.Client != nil
}

func (c Client) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	if !c.Enabled() {
		return ErrDisabled
	}
	_, err := c.PutObject(ctx, c.bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("objectstore put: %w", err)
	}
	return nil
}

func (c Client) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if !c.Enabled() {
		return nil, ErrDisabled
	}
	obj, err := c.GetObject(ctx, c.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("objectstore get: %w", err)
	}
	if _, err := obj.Stat(); err != nil {
		obj.Close()
		return nil, fmt.Errorf("objectstore stat: %w", err)
	}
	return obj, nil
}

func (c Client) Remove(ctx context.Context, key string) error {
	if !c.Enabled() {
		return ErrDisabled
	}
	if err := c.RemoveObject(ctx, c.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("objectstore remove: %w", err)
	}
	return nil
}
