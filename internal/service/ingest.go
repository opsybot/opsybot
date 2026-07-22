package service

//go:generate go tool mockgen -source=ingest.go -destination=./ingest/ingest_mock.go -package=ingest

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Ingest interface {
	Webhook(ctx context.Context, req entity.IngestRequest) ([]entity.IngestResult, error)
}
