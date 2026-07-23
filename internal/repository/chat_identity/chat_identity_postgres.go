package chat_identity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	dbpostgres "github.com/opsybot/opsybot/internal/db/postgres"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.ChatIdentity {
	return &repo{db: db}
}

func toEntity(m *dbpostgres.ChatIdentity) entity.ChatIdentity {
	return entity.ChatIdentity{
		ID: m.ID, ConnectionID: m.ConnectionID, UserID: m.UserID,
		ProviderUserID: m.ProviderUserID, ProviderHandle: m.ProviderHandle,
		DMChannelID: m.DMChannelID, ResolvedBy: m.ResolvedBy, Verified: m.VerifiedAt.Valid,
	}
}

func (r *repo) Upsert(ctx context.Context, in entity.ChatIdentity) (entity.ChatIdentity, error) {
	m := &dbpostgres.ChatIdentity{
		ConnectionID: in.ConnectionID, UserID: in.UserID, ProviderUserID: in.ProviderUserID,
		ProviderHandle: in.ProviderHandle, DMChannelID: in.DMChannelID, ResolvedBy: in.ResolvedBy,
		UpdatedAt: time.Now().UTC(),
	}
	if in.Verified {
		m.VerifiedAt = null.TimeFrom(time.Now().UTC())
	}
	if err := m.Upsert(ctx, r.db.Querier(ctx), true,
		[]string{"connection_id", "user_id"},
		boil.Whitelist("provider_user_id", "provider_handle", "dm_channel_id", "resolved_by", "verified_at", "updated_at"),
		boil.Whitelist("connection_id", "user_id", "provider_user_id", "provider_handle", "dm_channel_id", "resolved_by", "verified_at")); err != nil {
		return entity.ChatIdentity{}, fmt.Errorf("upsert chat identity: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) GetForUser(ctx context.Context, connectionID, userID string) (entity.ChatIdentity, error) {
	m, err := dbpostgres.ChatIdentities(
		qm.Where("connection_id = ? AND user_id = ?", connectionID, userID),
	).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ChatIdentity{}, entity.ErrChatNotConnected
		}
		return entity.ChatIdentity{}, fmt.Errorf("get chat identity: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) LinkedProviders(ctx context.Context, workspaceID, userID string) ([]entity.ChatProvider, error) {
	var rows []struct {
		Provider string `boil:"provider"`
	}
	const sql = `SELECT cc.provider FROM chat_identities ci
JOIN chat_connections cc ON cc.id = ci.connection_id
WHERE cc.workspace_id = $1 AND ci.user_id = $2 AND ci.verified_at IS NOT NULL`
	if err := queries.Raw(sql, workspaceID, userID).Bind(ctx, r.db.Querier(ctx), &rows); err != nil {
		return nil, fmt.Errorf("list linked providers: %w", err)
	}
	out := make([]entity.ChatProvider, 0, len(rows))
	for _, m := range rows {
		out = append(out, entity.ChatProvider(m.Provider))
	}
	return out, nil
}

func (r *repo) SetDMChannel(ctx context.Context, id, dmChannelID string) error {
	_, err := dbpostgres.ChatIdentities(qm.Where("id = ?", id)).
		UpdateAll(ctx, r.db.Querier(ctx), dbpostgres.M{"dm_channel_id": dmChannelID, "updated_at": time.Now().UTC()})
	if err != nil {
		return fmt.Errorf("set dm channel: %w", err)
	}
	return nil
}
