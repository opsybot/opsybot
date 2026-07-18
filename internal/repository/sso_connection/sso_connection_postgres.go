package sso_connection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/aarondl/sqlboiler/v4/queries/qm"

	dbpostgres "github.com/opsybot/opsybot/internal/db/postgres"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/pkg/secretbox"
	"github.com/opsybot/opsybot/internal/repository"
)

const columns = `id, workspace_id, mode, issuer, client_id, (client_secret_enc IS NOT NULL),
	array_to_string(scopes, ','), saml_metadata_url, enabled, required, jit_provisioning,
	array_to_string(allowed_email_domains, ','), updated_at`

type repo struct {
	db  postgres.Client
	box secretbox.Client
}

func New(db postgres.Client, box secretbox.Client) repository.SSOConnection {
	return &repo{db: db, box: box}
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func scanConnection(row interface {
	Scan(dest ...any) error
}) (entity.SSOConnection, error) {
	var (
		c          entity.SSOConnection
		mode       string
		scopesCSV  string
		domainsCSV string
	)
	if err := row.Scan(&c.ID, &c.WorkspaceID, &mode, &c.Issuer, &c.ClientID, &c.HasClientSecret,
		&scopesCSV, &c.SAMLMetadataURL, &c.Enabled, &c.Required, &c.JITProvisioning, &domainsCSV, &c.UpdatedAt); err != nil {
		return entity.SSOConnection{}, err
	}
	c.Mode = entity.SSOMode(mode)
	c.Scopes = splitCSV(scopesCSV)
	c.AllowedEmailDomains = splitCSV(domainsCSV)
	return c, nil
}

func toEntity(m *dbpostgres.SsoConnection) entity.SSOConnection {
	return entity.SSOConnection{
		ID:                  m.ID,
		WorkspaceID:         m.WorkspaceID,
		Mode:                entity.SSOMode(m.Mode),
		Issuer:              m.Issuer,
		ClientID:            m.ClientID,
		HasClientSecret:     m.ClientSecretEnc.Valid,
		Scopes:              splitArray(m.Scopes),
		SAMLMetadataURL:     m.SamlMetadataURL,
		Enabled:             m.Enabled,
		Required:            m.Required,
		JITProvisioning:     m.JitProvisioning,
		AllowedEmailDomains: splitArray(m.AllowedEmailDomains),
		UpdatedAt:           m.UpdatedAt,
	}
}

func splitArray(a []string) []string {
	if len(a) == 0 {
		return nil
	}
	return a
}

func (r *repo) Get(ctx context.Context, workspaceID string) (entity.SSOConnection, error) {
	m, err := dbpostgres.SsoConnections(qm.Where("workspace_id = ?", workspaceID)).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.SSOConnection{}, entity.ErrSSONotConfigured
		}
		return entity.SSOConnection{}, fmt.Errorf("get sso connection: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) ClientSecret(ctx context.Context, workspaceID string) (string, error) {
	m, err := dbpostgres.SsoConnections(
		qm.Select("client_secret_enc"),
		qm.Where("workspace_id = ?", workspaceID),
	).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", entity.ErrSSONotConfigured
		}
		return "", fmt.Errorf("get sso client secret: %w", err)
	}
	if len(m.ClientSecretEnc.Bytes) == 0 {
		return "", nil
	}
	plain, err := r.box.Decrypt(m.ClientSecretEnc.Bytes)
	if err != nil {
		return "", fmt.Errorf("decrypt sso client secret: %w", err)
	}
	return string(plain), nil
}

func (r *repo) Save(ctx context.Context, workspaceID string, in entity.SSOConfigInput) (entity.SSOConnection, error) {
	var enc []byte
	writeSecret := in.ClearClientSecret
	if in.ClientSecret != "" {
		if !r.box.Enabled() {
			return entity.SSOConnection{}, entity.ErrSSOUnavailable
		}
		sealed, err := r.box.Encrypt([]byte(in.ClientSecret))
		if err != nil {
			return entity.SSOConnection{}, fmt.Errorf("encrypt sso client secret: %w", err)
		}
		enc = sealed
		writeSecret = true
	}

	c, err := scanConnection(r.db.Querier(ctx).QueryRowContext(ctx,
		`INSERT INTO sso_connections
		   (workspace_id, mode, issuer, client_id, client_secret_enc, scopes, saml_metadata_url,
		    enabled, required, jit_provisioning, allowed_email_domains, updated_at)
		 VALUES ($1, $2, $3, $4, $5,
		    COALESCE(string_to_array(NULLIF($6, ''), ','), ARRAY[]::text[]),
		    $7, $8, $9, $10,
		    COALESCE(string_to_array(NULLIF($11, ''), ','), ARRAY[]::text[]), now())
		 ON CONFLICT (workspace_id) DO UPDATE SET
		    mode = EXCLUDED.mode, issuer = EXCLUDED.issuer, client_id = EXCLUDED.client_id,
		    scopes = EXCLUDED.scopes, saml_metadata_url = EXCLUDED.saml_metadata_url,
		    enabled = EXCLUDED.enabled, required = EXCLUDED.required,
		    jit_provisioning = EXCLUDED.jit_provisioning,
		    allowed_email_domains = EXCLUDED.allowed_email_domains, updated_at = now(),
		    client_secret_enc = CASE WHEN $12 THEN EXCLUDED.client_secret_enc ELSE sso_connections.client_secret_enc END
		 RETURNING `+columns,
		workspaceID, string(in.Mode), in.Issuer, in.ClientID, enc,
		strings.Join(in.Scopes, ","), in.SAMLMetadataURL, in.Enabled, in.Required, in.JITProvisioning,
		strings.Join(in.AllowedEmailDomains, ","), writeSecret))
	if err != nil {
		return entity.SSOConnection{}, fmt.Errorf("save sso connection: %w", err)
	}
	return c, nil
}
