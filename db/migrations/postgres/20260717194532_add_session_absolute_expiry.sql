-- +goose Up
ALTER TABLE sessions ADD COLUMN absolute_expires_at timestamptz;

-- +goose Down
ALTER TABLE sessions DROP COLUMN absolute_expires_at;
