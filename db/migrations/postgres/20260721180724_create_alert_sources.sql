-- +goose Up
CREATE TABLE alert_sources (
    id                         uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id               uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    slug                       text NOT NULL,
    name                       text NOT NULL,
    format                     text NOT NULL,
    ingest_token               text NOT NULL,
    signing_secret             text NOT NULL DEFAULT '',
    signing_secret_previous    text NOT NULL DEFAULT '',
    secret_rotated_at          timestamptz,
    require_signature          boolean NOT NULL DEFAULT false,
    default_severity           text NOT NULL DEFAULT 'warning',
    auto_resolve_after_seconds integer NOT NULL DEFAULT 0,
    last_event_at              timestamptz,
    failure_count              integer NOT NULL DEFAULT 0,
    paused_at                  timestamptz,
    created_at                 timestamptz NOT NULL DEFAULT now(),
    updated_at                 timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT alert_sources_slug_format CHECK (slug ~ '^[a-z][a-z0-9-]{0,39}$'),
    CONSTRAINT alert_sources_name_nonempty CHECK (btrim(name) <> ''),
    CONSTRAINT alert_sources_format_valid CHECK (format IN ('alertmanager', 'grafana', 'kuma', 'heartbeat', 'generic')),
    CONSTRAINT alert_sources_severity_valid CHECK (default_severity IN ('critical', 'high', 'warning')),
    CONSTRAINT alert_sources_auto_resolve_range CHECK (auto_resolve_after_seconds BETWEEN 0 AND 2592000),
    CONSTRAINT alert_sources_failure_count_positive CHECK (failure_count >= 0)
);
CREATE UNIQUE INDEX alert_sources_ws_slug_uq ON alert_sources (workspace_id, slug);
CREATE UNIQUE INDEX alert_sources_id_ws_uq ON alert_sources (id, workspace_id);
CREATE UNIQUE INDEX alert_sources_token_uq ON alert_sources (ingest_token);

-- +goose Down
DROP TABLE alert_sources;
