package recovery_code

import (
	"context"
	"fmt"

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
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`DELETE FROM user_recovery_codes WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("clear recovery codes: %w", err)
	}
	for _, hash := range codeHashes {
		if _, err := r.db.Querier(ctx).ExecContext(ctx,
			`INSERT INTO user_recovery_codes (user_id, code_hash) VALUES ($1, $2)`, userID, hash); err != nil {
			return fmt.Errorf("insert recovery code: %w", err)
		}
	}
	return nil
}

func (r *repo) ListUnusedHashes(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.Querier(ctx).QueryContext(ctx,
		`SELECT code_hash FROM user_recovery_codes WHERE user_id = $1 AND used_at IS NULL`, userID)
	if err != nil {
		return nil, fmt.Errorf("list recovery codes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []string
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return nil, fmt.Errorf("scan recovery code: %w", err)
		}
		out = append(out, hash)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recovery codes: %w", err)
	}
	return out, nil
}

func (r *repo) MarkUsed(ctx context.Context, userID, codeHash string) (bool, error) {
	res, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE user_recovery_codes SET used_at = now() WHERE user_id = $1 AND code_hash = $2 AND used_at IS NULL`,
		userID, codeHash)
	if err != nil {
		return false, fmt.Errorf("mark recovery code used: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("mark recovery code rows: %w", err)
	}
	return n > 0, nil
}

func (r *repo) CountUnused(ctx context.Context, userID string) (int, error) {
	var count int
	if err := r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT count(*) FROM user_recovery_codes WHERE user_id = $1 AND used_at IS NULL`, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count recovery codes: %w", err)
	}
	return count, nil
}
