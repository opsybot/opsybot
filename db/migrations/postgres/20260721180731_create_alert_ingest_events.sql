-- +goose Up
CREATE TABLE alert_ingest_events (
    id           uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    source_id    uuid NOT NULL REFERENCES alert_sources(id) ON DELETE CASCADE,
    alert_id     uuid REFERENCES alerts(id) ON DELETE SET NULL,
    dedup_key    text NOT NULL DEFAULT '',
    outcome      text NOT NULL,
    at           timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT alert_ingest_events_outcome_valid CHECK (
        outcome IN ('created', 'updated', 'duplicate', 'resolved', 'stale', 'failed', 'flood_dropped')
    )
);
CREATE INDEX alert_ingest_events_source_at_idx ON alert_ingest_events (source_id, at DESC);

CREATE TABLE alert_ingest_failures (
    id           uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    source_id    uuid REFERENCES alert_sources(id) ON DELETE CASCADE,
    reason       text NOT NULL,
    detail       text NOT NULL DEFAULT '',
    payload      text NOT NULL DEFAULT '',
    at           timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX alert_ingest_failures_ws_at_idx ON alert_ingest_failures (workspace_id, at DESC);

-- +goose Down
DROP TABLE alert_ingest_failures;
DROP TABLE alert_ingest_events;
