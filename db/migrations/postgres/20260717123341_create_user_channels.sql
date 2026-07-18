-- +goose Up
CREATE TABLE user_channels (
    id          uuid PRIMARY KEY DEFAULT uuidv7(),
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        text NOT NULL,
    detail      text NOT NULL,
    verified_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_channels_type CHECK (type IN ('slack','teams','discord','telegram','ntfy','email','webhook')),
    CONSTRAINT user_channels_detail_nonempty CHECK (btrim(detail) <> '')
);
CREATE UNIQUE INDEX user_channels_user_type_detail_uq ON user_channels (user_id, type, detail);

-- +goose Down
DROP TABLE user_channels;
