-- +goose Up
CREATE TABLE user_recovery_codes (
    id         uuid PRIMARY KEY DEFAULT uuidv7(),
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash  text NOT NULL,
    used_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX user_recovery_codes_user_hash_uq ON user_recovery_codes (user_id, code_hash);

-- +goose Down
DROP TABLE user_recovery_codes;
