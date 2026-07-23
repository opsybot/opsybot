-- +goose Up
ALTER TABLE user_channels ADD COLUMN secret_enc bytea;
ALTER TABLE user_channels ADD COLUMN label text NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE user_channels DROP COLUMN label;
ALTER TABLE user_channels DROP COLUMN secret_enc;
