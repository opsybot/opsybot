package transactor

import (
	"context"

	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

type transactor struct {
	db postgres.Client
}

func New(db postgres.Client) repository.Transactor {
	return &transactor{db: db}
}

func (t *transactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return t.db.WithTx(ctx, fn)
}
