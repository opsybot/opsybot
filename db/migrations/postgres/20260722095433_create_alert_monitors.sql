-- +goose Up
CREATE TABLE alert_monitors (
    id               uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id     uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    source_id        uuid NOT NULL,
    interval_seconds integer NOT NULL,
    grace_seconds    integer NOT NULL DEFAULT 300,
    policy_ref       text NOT NULL DEFAULT '',
    severity         text NOT NULL DEFAULT 'high',
    last_check_in_at timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT alert_monitors_interval_range CHECK (interval_seconds BETWEEN 60 AND 2592000),
    CONSTRAINT alert_monitors_grace_range CHECK (grace_seconds BETWEEN 0 AND 86400),
    CONSTRAINT alert_monitors_severity_valid CHECK (severity IN ('critical', 'high', 'warning')),
    FOREIGN KEY (source_id, workspace_id) REFERENCES alert_sources(id, workspace_id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX alert_monitors_source_uq ON alert_monitors (source_id);
CREATE INDEX alert_monitors_ws_idx ON alert_monitors (workspace_id);

-- +goose Down
DROP TABLE alert_monitors;
