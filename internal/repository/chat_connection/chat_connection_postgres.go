package chat_connection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/types"

	"github.com/opsybot/opsybot/internal/config"
	dbpostgres "github.com/opsybot/opsybot/internal/db/postgres"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/pkg/secretbox"
	"github.com/opsybot/opsybot/internal/repository"
)

const columns = `id, workspace_id, provider, external_id, external_name, bot_user_id,
	(bot_token_enc IS NOT NULL) AS has_bot_token, app_id, tenant_id, scopes, enabled,
	health, health_note, health_checked_at, naming_pattern, announce_channel, archive_on_resolve,
	connected_by, created_at, updated_at`

type repo struct {
	db            postgres.Client
	box           secretbox.Client
	slackToken    string
	discToken     string
	telegramToken string
}

func New(db postgres.Client, box secretbox.Client, slack config.Slack, discord config.Discord, telegram config.Telegram) repository.ChatConnection {
	return &repo{db: db, box: box, slackToken: slack.BotToken, discToken: discord.BotToken, telegramToken: telegram.BotToken}
}

func (r *repo) envToken(provider entity.ChatProvider) string {
	switch provider {
	case entity.ChatProviderSlack:
		return r.slackToken
	case entity.ChatProviderDiscord:
		return r.discToken
	case entity.ChatProviderTelegram:
		return r.telegramToken
	default:
		return ""
	}
}

type row struct {
	ID               string            `boil:"id"`
	WorkspaceID      string            `boil:"workspace_id"`
	Provider         string            `boil:"provider"`
	ExternalID       string            `boil:"external_id"`
	ExternalName     string            `boil:"external_name"`
	BotUserID        string            `boil:"bot_user_id"`
	HasBotToken      bool              `boil:"has_bot_token"`
	AppID            string            `boil:"app_id"`
	TenantID         string            `boil:"tenant_id"`
	Scopes           types.StringArray `boil:"scopes"`
	Enabled          bool              `boil:"enabled"`
	Health           string            `boil:"health"`
	HealthNote       string            `boil:"health_note"`
	HealthCheckedAt  null.Time         `boil:"health_checked_at"`
	NamingPattern    string            `boil:"naming_pattern"`
	AnnounceChannel  string            `boil:"announce_channel"`
	ArchiveOnResolve bool              `boil:"archive_on_resolve"`
	ConnectedBy      null.String       `boil:"connected_by"`
	CreatedAt        time.Time         `boil:"created_at"`
	UpdatedAt        time.Time         `boil:"updated_at"`
}

func (r row) toEntity() entity.ChatConnection {
	return entity.ChatConnection{
		ID: r.ID, WorkspaceID: r.WorkspaceID, Provider: entity.ChatProvider(r.Provider),
		ExternalID: r.ExternalID, ExternalName: r.ExternalName, BotUserID: r.BotUserID,
		HasBotToken: r.HasBotToken, AppID: r.AppID, TenantID: r.TenantID, Scopes: []string(r.Scopes),
		Enabled: r.Enabled, Health: entity.ChatHealth(r.Health), HealthNote: r.HealthNote,
		HealthCheckedAt: r.HealthCheckedAt.Time, NamingPattern: r.NamingPattern,
		AnnounceChannel: r.AnnounceChannel, ArchiveOnResolve: r.ArchiveOnResolve,
		ConnectedBy: r.ConnectedBy.String, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func (r *repo) List(ctx context.Context, workspaceID string) ([]entity.ChatConnection, error) {
	var rows []row
	if err := dbpostgres.NewQuery(
		qm.Select(columns), qm.From("chat_connections"),
		qm.Where("workspace_id = ?", workspaceID), qm.OrderBy("provider"),
	).Bind(ctx, r.db.Querier(ctx), &rows); err != nil {
		return nil, fmt.Errorf("list chat connections: %w", err)
	}
	out := make([]entity.ChatConnection, 0, len(rows))
	for _, m := range rows {
		out = append(out, m.toEntity())
	}
	return out, nil
}

func (r *repo) Get(ctx context.Context, workspaceID string, provider entity.ChatProvider) (entity.ChatConnection, error) {
	var m row
	err := dbpostgres.NewQuery(
		qm.Select(columns), qm.From("chat_connections"),
		qm.Where("workspace_id = ? AND provider = ?", workspaceID, string(provider)),
	).Bind(ctx, r.db.Querier(ctx), &m)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ChatConnection{}, entity.ErrChatNotConnected
		}
		return entity.ChatConnection{}, fmt.Errorf("get chat connection: %w", err)
	}
	return m.toEntity(), nil
}

func (r *repo) BotToken(ctx context.Context, workspaceID string, provider entity.ChatProvider) (string, error) {
	m, err := dbpostgres.ChatConnections(
		qm.Select("bot_token_enc"),
		qm.Where("workspace_id = ? AND provider = ?", workspaceID, string(provider)),
	).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", entity.ErrChatNotConnected
		}
		return "", fmt.Errorf("get chat bot token: %w", err)
	}
	if len(m.BotTokenEnc.Bytes) == 0 {
		return r.envToken(provider), nil
	}
	if !r.box.Enabled() {
		return "", entity.ErrChatSecretUnavailable
	}
	plain, err := r.box.Decrypt(m.BotTokenEnc.Bytes)
	if err != nil {
		return "", fmt.Errorf("decrypt chat bot token: %w", err)
	}
	return string(plain), nil
}

func (r *repo) SecretsEnabled(ctx context.Context) bool {
	_ = ctx
	return r.box.Enabled()
}

func (r *repo) Save(ctx context.Context, workspaceID string, in entity.ChatConnectionInput) (entity.ChatConnection, error) {
	scopes := types.StringArray(in.Scopes)
	if scopes == nil {
		scopes = types.StringArray{}
	}
	m := &dbpostgres.ChatConnection{
		WorkspaceID: workspaceID, Provider: string(in.Provider),
		ExternalID: in.ExternalID, ExternalName: in.ExternalName, BotUserID: in.BotUserID,
		AppID: in.AppID, Scopes: scopes, Enabled: true,
		Health: string(entity.ChatHealthy), HealthNote: "connected", HealthCheckedAt: null.TimeFrom(time.Now().UTC()),
		UpdatedAt: time.Now().UTC(),
	}
	if in.ConnectedBy != "" {
		m.ConnectedBy = null.StringFrom(in.ConnectedBy)
	}
	if in.NamingPattern != "" {
		m.NamingPattern = in.NamingPattern
	}
	if in.AnnounceChannel != "" {
		m.AnnounceChannel = in.AnnounceChannel
	}
	m.ArchiveOnResolve = in.ArchiveOnResolve
	cols := []string{"workspace_id", "provider", "external_id", "external_name", "bot_user_id",
		"app_id", "scopes", "enabled", "health", "health_note", "health_checked_at",
		"announce_channel", "archive_on_resolve", "connected_by", "updated_at"}
	if in.NamingPattern != "" {
		cols = append(cols, "naming_pattern")
	}
	if in.BotToken != "" {
		if !r.box.Enabled() {
			return entity.ChatConnection{}, entity.ErrChatSecretUnavailable
		}
		sealed, err := r.box.Encrypt([]byte(in.BotToken))
		if err != nil {
			return entity.ChatConnection{}, fmt.Errorf("seal chat bot token: %w", err)
		}
		m.BotTokenEnc = null.BytesFrom(sealed)
		cols = append(cols, "bot_token_enc")
	}
	if err := m.Upsert(ctx, r.db.Querier(ctx), true,
		[]string{"workspace_id", "provider"},
		boil.Whitelist(updateCols(cols)...),
		boil.Whitelist(cols...)); err != nil {
		return entity.ChatConnection{}, fmt.Errorf("save chat connection: %w", err)
	}
	return r.Get(ctx, workspaceID, in.Provider)
}

func updateCols(cols []string) []string {
	out := make([]string, 0, len(cols))
	for _, c := range cols {
		if c == "workspace_id" || c == "provider" {
			continue
		}
		out = append(out, c)
	}
	return out
}

func (r *repo) SetHealth(ctx context.Context, workspaceID string, provider entity.ChatProvider, health entity.ChatHealth, note string, at time.Time) error {
	_, err := dbpostgres.ChatConnections(
		qm.Where("workspace_id = ? AND provider = ?", workspaceID, string(provider)),
	).UpdateAll(ctx, r.db.Querier(ctx), dbpostgres.M{
		"health": string(health), "health_note": note, "health_checked_at": at, "updated_at": at,
	})
	if err != nil {
		return fmt.Errorf("set chat health: %w", err)
	}
	return nil
}

func (r *repo) SetDefaults(ctx context.Context, workspaceID string, provider entity.ChatProvider, namingPattern, announceChannel string, archiveOnResolve bool) error {
	_, err := dbpostgres.ChatConnections(
		qm.Where("workspace_id = ? AND provider = ?", workspaceID, string(provider)),
	).UpdateAll(ctx, r.db.Querier(ctx), dbpostgres.M{
		"naming_pattern": namingPattern, "announce_channel": announceChannel,
		"archive_on_resolve": archiveOnResolve, "updated_at": time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("set chat defaults: %w", err)
	}
	return nil
}

func (r *repo) Delete(ctx context.Context, workspaceID string, provider entity.ChatProvider) error {
	n, err := dbpostgres.ChatConnections(
		qm.Where("workspace_id = ? AND provider = ?", workspaceID, string(provider)),
	).DeleteAll(ctx, r.db.Querier(ctx))
	if err != nil {
		return fmt.Errorf("delete chat connection: %w", err)
	}
	if n == 0 {
		return entity.ErrChatNotConnected
	}
	return nil
}
