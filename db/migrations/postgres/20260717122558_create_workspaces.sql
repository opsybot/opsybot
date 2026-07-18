-- +goose Up
CREATE TABLE workspaces (
    id          uuid PRIMARY KEY DEFAULT uuidv7(),
    slug        text NOT NULL,
    name        text NOT NULL,
    timezone    text NOT NULL DEFAULT 'UTC',
    environment text NOT NULL DEFAULT '',
    created_by  uuid REFERENCES users(id) ON DELETE RESTRICT,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT workspaces_slug_format CHECK (slug ~ '^[a-z][a-z0-9-]{0,39}$'),
    CONSTRAINT workspaces_name_nonempty CHECK (btrim(name) <> '')
);
CREATE UNIQUE INDEX workspaces_slug_uq ON workspaces (slug);

-- +goose Down
DROP TABLE workspaces;
