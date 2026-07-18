-- +goose Up
CREATE TABLE teams (
    id           uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    slug         text NOT NULL,
    name         text NOT NULL,
    archived_at  timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT teams_slug_format CHECK (slug ~ '^[a-z][a-z0-9-]{0,39}$'),
    CONSTRAINT teams_name_nonempty CHECK (btrim(name) <> '')
);
CREATE UNIQUE INDEX teams_ws_slug_uq ON teams (workspace_id, slug);
CREATE UNIQUE INDEX teams_id_ws_uq ON teams (id, workspace_id);

-- +goose Down
DROP TABLE teams;
