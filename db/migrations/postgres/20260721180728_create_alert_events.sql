-- +goose Up
CREATE TABLE alert_events (
    id       uuid PRIMARY KEY DEFAULT uuidv7(),
    alert_id uuid NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    at       timestamptz NOT NULL DEFAULT now(),
    kind     text NOT NULL,
    text     text NOT NULL DEFAULT '',
    result   text NOT NULL DEFAULT '',
    CONSTRAINT alert_events_kind_valid CHECK (
        kind IN ('received', 'deduped', 'grouped', 'routed', 'suppressed', 'escalation', 'push', 'sms', 'timeout', 'chat', 'acked', 'resolved')
    )
);
CREATE INDEX alert_events_alert_at_idx ON alert_events (alert_id, at);

-- +goose Down
DROP TABLE alert_events;
