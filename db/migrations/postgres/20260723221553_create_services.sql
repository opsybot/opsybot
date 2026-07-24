-- +goose Up
CREATE TABLE services (
    id           uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    slug         text NOT NULL,
    name         text NOT NULL,
    team_id      uuid REFERENCES teams(id) ON DELETE SET NULL,
    description  text NOT NULL DEFAULT '',
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT services_slug_format CHECK (slug ~ '^[a-z][a-z0-9-]{0,39}$'),
    CONSTRAINT services_name_nonempty CHECK (btrim(name) <> '')
);
CREATE UNIQUE INDEX services_ws_slug_uq ON services (workspace_id, slug);
CREATE UNIQUE INDEX services_id_ws_uq ON services (id, workspace_id);

-- +goose Down
DROP TABLE services;
