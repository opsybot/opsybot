-- +goose Up
CREATE TABLE alert_links (
    id       uuid PRIMARY KEY DEFAULT uuidv7(),
    alert_id uuid NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    kind     text NOT NULL,
    label    text NOT NULL,
    url      text NOT NULL DEFAULT '',
    position integer NOT NULL DEFAULT 0,
    CONSTRAINT alert_links_kind_valid CHECK (kind IN ('runbook', 'dashboard', 'source'))
);
CREATE INDEX alert_links_alert_idx ON alert_links (alert_id, position);

-- +goose Down
DROP TABLE alert_links;
