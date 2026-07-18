package recovery_code

import (
	"context"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	dbpostgres "github.com/opsybot/opsybot/internal/db/postgres"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.RecoveryCode {
	return &repo{db: db}
}

func (r *repo) Replace(ctx context.Context, userID string, codeHashes []string) error {
	exec := r.db.Querier(ctx)
	if _, err := dbpostgres.UserRecoveryCodes(qm.Where("user_id = ?", userID)).DeleteAll(ctx, exec); err != nil {
		return fmt.Errorf("clear recovery codes: %w", err)
	}
	for _, hash := range codeHashes {
		m := &dbpostgres.UserRecoveryCode{UserID: userID, CodeHash: hash}
		if err := m.Insert(ctx, exec, boil.Whitelist("user_id", "code_hash")); err != nil {
			return fmt.Errorf("insert recovery code: %w", err)
		}
	}
	return nil
}

func (r *repo) ListUnusedHashes(ctx context.Context, userID string) ([]string, error) {
	rows, err := dbpostgres.UserRecoveryCodes(
		qm.Select("code_hash"),
		qm.Where("user_id = ?", userID),
		qm.Where("used_at IS NULL"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list recovery codes: %w", err)
	}
	out := make([]string, 0, len(rows))
	for _, m := range rows {
		out = append(out, m.CodeHash)
	}
	return out, nil
}

func (r *repo) MarkUsed(ctx context.Context, userID, codeHash string) (bool, error) {
	n, err := dbpostgres.UserRecoveryCodes(
		qm.Where("user_id = ? AND code_hash = ?", userID, codeHash),
		qm.Where("used_at IS NULL"),
	).UpdateAll(ctx, r.db.Querier(ctx), dbpostgres.M{"used_at": time.Now()})
	if err != nil {
		return false, fmt.Errorf("mark recovery code used: %w", err)
	}
	return n > 0, nil
}

func (r *repo) CountUnused(ctx context.Context, userID string) (int, error) {
	n, err := dbpostgres.UserRecoveryCodes(
		qm.Where("user_id = ?", userID),
		qm.Where("used_at IS NULL"),
	).Count(ctx, r.db.Querier(ctx))
	if err != nil {
		return 0, fmt.Errorf("count recovery codes: %w", err)
	}
	return int(n), nil
}
