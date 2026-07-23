-- +goose Up
CREATE TABLE alert_action_tokens (
    id           uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    alert_id     uuid NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel_id   uuid REFERENCES user_channels(id) ON DELETE SET NULL,
    action       text NOT NULL,
    token_hash   text NOT NULL,
    expires_at   timestamptz NOT NULL,
    used_at      timestamptz,
    used_ip      inet,
    created_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT alert_action_tokens_action CHECK (action IN ('ack', 'resolve'))
);
CREATE UNIQUE INDEX alert_action_tokens_hash_uq ON alert_action_tokens (token_hash);
CREATE INDEX alert_action_tokens_alert_idx ON alert_action_tokens (alert_id) WHERE used_at IS NULL;

-- +goose Down
DROP TABLE alert_action_tokens;
