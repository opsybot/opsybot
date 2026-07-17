-- +goose Up
CREATE TABLE users (
    id              uuid PRIMARY KEY DEFAULT uuidv7(),
    email           text NOT NULL,
    name            text NOT NULL DEFAULT '',
    password_hash   text,
    timezone        text NOT NULL DEFAULT 'UTC',
    totp_secret_enc bytea,
    totp_enabled_at timestamptz,
    last_active_at  timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_email_format CHECK (email ~ '^[^@[:space:]]+@[^@[:space:]]+\.[^@[:space:]]+$'),
    CONSTRAINT users_totp_pair CHECK (totp_enabled_at IS NULL OR totp_secret_enc IS NOT NULL)
);
CREATE UNIQUE INDEX users_email_lower_uq ON users (lower(email));

-- +goose Down
DROP TABLE users;
