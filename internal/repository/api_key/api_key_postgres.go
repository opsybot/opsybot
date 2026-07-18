package api_key

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

const keyColumns = `id, workspace_id, kind, owner_user_id, created_by, name, token_hint,
	array_to_string(scopes, ','), last_used_at, revoked_at, created_at`

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.APIKey {
	return &repo{db: db}
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanKey(row rowScanner) (entity.APIKey, error) {
	var (
		k         entity.APIKey
		kind      string
		owner     sql.NullString
		createdBy sql.NullString
		scopesCSV string
		lastUsed  sql.NullTime
		revoked   sql.NullTime
	)
	if err := row.Scan(&k.ID, &k.WorkspaceID, &kind, &owner, &createdBy, &k.Name, &k.TokenHint,
		&scopesCSV, &lastUsed, &revoked, &k.CreatedAt); err != nil {
		return entity.APIKey{}, err
	}
	k.Kind = entity.KeyKind(kind)
	k.OwnerUserID = owner.String
	k.CreatedBy = createdBy.String
	k.Scopes = parseScopes(scopesCSV)
	k.LastUsedAt = lastUsed.Time
	k.RevokedAt = revoked.Time
	return k, nil
}

func parseScopes(csv string) []entity.Scope {
	if csv == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	out := make([]entity.Scope, 0, len(parts))
	for _, p := range parts {
		out = append(out, entity.Scope(p))
	}
	return out
}

func joinScopes(scopes []entity.Scope) string {
	parts := make([]string, 0, len(scopes))
	for _, s := range scopes {
		parts = append(parts, string(s))
	}
	return strings.Join(parts, ",")
}

func (r *repo) Create(ctx context.Context, rec entity.APIKeyRecord) (entity.APIKey, error) {
	owner := sql.NullString{String: rec.OwnerUserID, Valid: rec.OwnerUserID != ""}
	createdBy := sql.NullString{String: rec.CreatedBy, Valid: rec.CreatedBy != ""}
	k, err := scanKey(r.db.Querier(ctx).QueryRowContext(ctx,
		`INSERT INTO api_keys (workspace_id, kind, owner_user_id, created_by, name, token_hash, token_hint, scopes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, string_to_array($8, ','))
		 RETURNING `+keyColumns,
		rec.WorkspaceID, string(rec.Kind), owner, createdBy, rec.Name, rec.TokenHash, rec.TokenHint, joinScopes(rec.Scopes)))
	if err != nil {
		return entity.APIKey{}, fmt.Errorf("create api key: %w", err)
	}
	return k, nil
}

func (r *repo) ListByOwner(ctx context.Context, workspaceID, ownerUserID string) ([]entity.APIKey, error) {
	return r.list(ctx,
		`SELECT `+keyColumns+` FROM api_keys
		 WHERE workspace_id = $1 AND owner_user_id = $2 AND revoked_at IS NULL
		 ORDER BY created_at DESC`, workspaceID, ownerUserID)
}

func (r *repo) ListWorkspaceKeys(ctx context.Context, workspaceID string) ([]entity.APIKey, error) {
	return r.list(ctx,
		`SELECT `+keyColumns+` FROM api_keys
		 WHERE workspace_id = $1 AND kind = 'workspace' AND revoked_at IS NULL
		 ORDER BY created_at DESC`, workspaceID)
}

func (r *repo) list(ctx context.Context, query string, args ...any) ([]entity.APIKey, error) {
	rows, err := r.db.Querier(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]entity.APIKey, 0)
	for rows.Next() {
		k, err := scanKey(rows)
		if err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		out = append(out, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate api keys: %w", err)
	}
	return out, nil
}

func (r *repo) GetByID(ctx context.Context, workspaceID, id string) (entity.APIKey, error) {
	k, err := scanKey(r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT `+keyColumns+` FROM api_keys WHERE workspace_id = $1 AND id = $2`, workspaceID, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.APIKey{}, entity.ErrAPIKeyNotFound
		}
		return entity.APIKey{}, fmt.Errorf("get api key: %w", err)
	}
	return k, nil
}

func (r *repo) GetByTokenHash(ctx context.Context, tokenHash string) (entity.APIKey, error) {
	k, err := scanKey(r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT `+keyColumns+` FROM api_keys WHERE token_hash = $1`, tokenHash))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.APIKey{}, entity.ErrAPIKeyNotFound
		}
		return entity.APIKey{}, fmt.Errorf("get api key by hash: %w", err)
	}
	return k, nil
}

func (r *repo) Revoke(ctx context.Context, workspaceID, id string) error {
	res, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE api_keys SET revoked_at = now() WHERE workspace_id = $1 AND id = $2 AND revoked_at IS NULL`,
		workspaceID, id)
	if err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("revoke api key rows: %w", err)
	}
	if n == 0 {
		return entity.ErrAPIKeyNotFound
	}
	return nil
}

func (r *repo) TouchLastUsed(ctx context.Context, id string, at time.Time) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE api_keys SET last_used_at = $2 WHERE id = $1`, id, at); err != nil {
		return fmt.Errorf("touch api key: %w", err)
	}
	return nil
}
