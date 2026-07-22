-- +goose Up
CREATE TABLE alerts (
    id                       uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id             uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    source_id                uuid NOT NULL,
    parent_alert_id          uuid REFERENCES alerts(id) ON DELETE SET NULL,
    dedup_key                text NOT NULL,
    group_key                text NOT NULL DEFAULT '',
    title                    text NOT NULL,
    description              text NOT NULL DEFAULT '',
    severity                 text NOT NULL,
    status                   text NOT NULL DEFAULT 'open',
    source_label             text NOT NULL DEFAULT '',
    service_name             text NOT NULL DEFAULT '',
    labels                   jsonb NOT NULL DEFAULT '{}'::jsonb,
    count                    integer NOT NULL DEFAULT 1,
    started_at               timestamptz NOT NULL,
    last_seen_at             timestamptz NOT NULL,
    ended_at                 timestamptz,
    acked_at                 timestamptz,
    resolved_at              timestamptz,
    acked_by_user_id         uuid REFERENCES users(id) ON DELETE SET NULL,
    acked_by_label           text NOT NULL DEFAULT '',
    resolve_mode             text NOT NULL DEFAULT '',
    routed_policy_ref        text NOT NULL DEFAULT '',
    suppressed_by_silence_id uuid,
    suppressed_at            timestamptz,
    payload                  text NOT NULL DEFAULT '',
    created_at               timestamptz NOT NULL DEFAULT now(),
    updated_at               timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT alerts_status_valid CHECK (status IN ('open', 'acked', 'resolved')),
    CONSTRAINT alerts_severity_valid CHECK (severity IN ('critical', 'high', 'warning')),
    CONSTRAINT alerts_resolve_mode_valid CHECK (resolve_mode IN ('', 'source', 'manual', 'incident', 'timeout')),
    CONSTRAINT alerts_count_positive CHECK (count >= 1),
    CONSTRAINT alerts_title_nonempty CHECK (btrim(title) <> ''),
    CONSTRAINT alerts_window_ordered CHECK (last_seen_at >= started_at),
    CONSTRAINT alerts_resolved_has_time CHECK ((status = 'resolved') = (resolved_at IS NOT NULL)),
    FOREIGN KEY (source_id, workspace_id) REFERENCES alert_sources(id, workspace_id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX alerts_open_dedup_uq ON alerts (workspace_id, source_id, dedup_key) WHERE resolved_at IS NULL;
CREATE UNIQUE INDEX alerts_id_ws_uq ON alerts (id, workspace_id);
CREATE INDEX alerts_ws_last_seen_idx ON alerts (workspace_id, last_seen_at DESC, id DESC);
CREATE INDEX alerts_ws_status_idx ON alerts (workspace_id, status);
CREATE INDEX alerts_parent_idx ON alerts (parent_alert_id) WHERE parent_alert_id IS NOT NULL;
CREATE INDEX alerts_ws_group_idx ON alerts (workspace_id, group_key) WHERE group_key <> '';

-- +goose Down
DROP TABLE alerts;
