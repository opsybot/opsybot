-- +goose Up
CREATE TABLE schedules (
    id           uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    team_id      uuid NOT NULL,
    slug         text NOT NULL,
    timezone     text NOT NULL,
    feed_token   text NOT NULL,
    paused_at    timestamptz,
    archived_at  timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT schedules_slug_format CHECK (slug ~ '^[a-z][a-z0-9-]{0,39}$'),
    CONSTRAINT schedules_timezone_nonempty CHECK (btrim(timezone) <> ''),
    FOREIGN KEY (team_id, workspace_id) REFERENCES teams(id, workspace_id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX schedules_ws_slug_uq ON schedules (workspace_id, slug);
CREATE UNIQUE INDEX schedules_id_ws_uq ON schedules (id, workspace_id);
CREATE UNIQUE INDEX schedules_feed_token_uq ON schedules (feed_token);
CREATE INDEX schedules_ws_team_idx ON schedules (workspace_id, team_id);

-- +goose Down
DROP TABLE schedules;
