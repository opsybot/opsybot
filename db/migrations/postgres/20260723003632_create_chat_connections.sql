-- +goose Up
CREATE TABLE chat_connections (
    id                 uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id       uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    provider           text NOT NULL,
    external_id        text NOT NULL DEFAULT '',
    external_name      text NOT NULL DEFAULT '',
    bot_user_id        text NOT NULL DEFAULT '',
    bot_token_enc      bytea,
    app_id             text NOT NULL DEFAULT '',
    tenant_id          text NOT NULL DEFAULT '',
    service_url        text NOT NULL DEFAULT '',
    hook_token_hash    text,
    hook_secret_enc    bytea,
    scopes             text[] NOT NULL DEFAULT ARRAY[]::text[],
    enabled            boolean NOT NULL DEFAULT true,
    health             text NOT NULL DEFAULT 'failing',
    health_note        text NOT NULL DEFAULT '',
    health_checked_at  timestamptz,
    naming_pattern     text NOT NULL DEFAULT 'inc-{number}',
    announce_channel   text NOT NULL DEFAULT '',
    archive_on_resolve boolean NOT NULL DEFAULT true,
    connected_by       uuid REFERENCES users(id) ON DELETE SET NULL,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT chat_connections_provider CHECK (provider IN ('slack', 'teams', 'discord', 'telegram')),
    CONSTRAINT chat_connections_health CHECK (health IN ('healthy', 'failing'))
);
CREATE UNIQUE INDEX chat_connections_ws_provider_uq ON chat_connections (workspace_id, provider);
CREATE UNIQUE INDEX chat_connections_hook_uq ON chat_connections (hook_token_hash) WHERE hook_token_hash IS NOT NULL;

CREATE TABLE chat_identities (
    id               uuid PRIMARY KEY DEFAULT uuidv7(),
    connection_id    uuid NOT NULL REFERENCES chat_connections(id) ON DELETE CASCADE,
    user_id          uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_user_id text NOT NULL,
    provider_handle  text NOT NULL DEFAULT '',
    dm_channel_id    text NOT NULL DEFAULT '',
    resolved_by      text NOT NULL DEFAULT 'manual',
    verified_at      timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX chat_identities_conn_user_uq ON chat_identities (connection_id, user_id);
CREATE UNIQUE INDEX chat_identities_conn_pid_uq ON chat_identities (connection_id, provider_user_id);

INSERT INTO casbin_rule (p_type, v0, v1, v2, v3)
SELECT 'p', r.role, w.id::text, r.object, r.action
FROM workspaces w
CROSS JOIN (VALUES
    ('admin', 'chat', 'read'), ('admin', 'chat', 'write'),
    ('member', 'chat', 'read')
) AS r(role, object, action)
WHERE NOT EXISTS (
    SELECT 1 FROM casbin_rule c
    WHERE c.p_type = 'p' AND c.v0 = r.role AND c.v1 = w.id::text AND c.v2 = r.object AND c.v3 = r.action
);

-- +goose Down
DELETE FROM casbin_rule WHERE p_type = 'p' AND v2 = 'chat';
DROP TABLE chat_identities;
DROP TABLE chat_connections;
