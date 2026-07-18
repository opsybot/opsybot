-- +goose Up
ALTER TABLE users ADD COLUMN totp_last_step bigint;

-- +goose Down
ALTER TABLE users DROP COLUMN totp_last_step;
