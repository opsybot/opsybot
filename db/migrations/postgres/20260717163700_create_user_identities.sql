-- +goose Up
CREATE TABLE user_identities (
    id            uuid PRIMARY KEY DEFAULT uuidv7(),
    user_id       uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    connection_id uuid NOT NULL REFERENCES sso_connections(id) ON DELETE CASCADE,
    subject       text NOT NULL,
    email         text NOT NULL DEFAULT '',
    created_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_identities_conn_subject_uq UNIQUE (connection_id, subject)
);
CREATE INDEX user_identities_user_idx ON user_identities (user_id);

-- +goose Down
DROP TABLE user_identities;
