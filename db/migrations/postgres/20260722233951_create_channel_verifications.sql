-- +goose Up
CREATE TABLE channel_verifications (
    id         uuid PRIMARY KEY DEFAULT uuidv7(),
    channel_id uuid NOT NULL REFERENCES user_channels(id) ON DELETE CASCADE,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    method     text NOT NULL,
    token_hash text NOT NULL,
    code_hash  text NOT NULL DEFAULT '',
    nonce      text NOT NULL DEFAULT '',
    expires_at timestamptz NOT NULL,
    used_at    timestamptz,
    attempts   integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT channel_verifications_method_valid CHECK (method IN ('email', 'ntfy', 'webhook', 'telegram', 'chat')),
    CONSTRAINT channel_verifications_attempts CHECK (attempts >= 0)
);
CREATE UNIQUE INDEX channel_verifications_token_uq ON channel_verifications (token_hash);
CREATE INDEX channel_verifications_open_idx ON channel_verifications (channel_id) WHERE used_at IS NULL;

-- +goose Down
DROP TABLE channel_verifications;
