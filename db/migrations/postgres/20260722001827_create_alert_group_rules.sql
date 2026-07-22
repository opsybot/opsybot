-- +goose Up
CREATE TABLE alert_group_rules (
    id             uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id   uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    fields         text[] NOT NULL,
    window_seconds integer NOT NULL DEFAULT 300,
    position       integer NOT NULL DEFAULT 0,
    created_at     timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT alert_group_rules_fields_nonempty CHECK (cardinality(fields) BETWEEN 1 AND 5),
    CONSTRAINT alert_group_rules_window_range CHECK (window_seconds BETWEEN 60 AND 86400)
);
CREATE INDEX alert_group_rules_ws_position_idx ON alert_group_rules (workspace_id, position);

-- +goose Down
DROP TABLE alert_group_rules;
