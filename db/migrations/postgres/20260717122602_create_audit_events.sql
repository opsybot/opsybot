-- +goose Up
CREATE TABLE audit_events (
    id               uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id     uuid REFERENCES workspaces(id) ON DELETE CASCADE,
    at               timestamptz NOT NULL DEFAULT now(),
    actor_user_id    uuid REFERENCES users(id) ON DELETE SET NULL,
    actor_api_key_id uuid,
    actor_label      text NOT NULL DEFAULT '',
    action           text NOT NULL,
    target           text NOT NULL DEFAULT '',
    ip               inet,
    meta             jsonb,
    CONSTRAINT audit_events_action_format CHECK (action ~ '^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$')
);
CREATE INDEX audit_events_ws_at_idx ON audit_events (workspace_id, at DESC, id DESC);
CREATE INDEX audit_events_ws_action_idx ON audit_events (workspace_id, action text_pattern_ops);
CREATE INDEX audit_events_actor_at_idx ON audit_events (actor_user_id, at DESC);

-- +goose Down
DROP TABLE audit_events;
