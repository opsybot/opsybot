-- +goose Up
CREATE TABLE team_members (
    team_id      uuid NOT NULL,
    workspace_id uuid NOT NULL,
    user_id      uuid NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (team_id, user_id),
    FOREIGN KEY (team_id, workspace_id) REFERENCES teams(id, workspace_id) ON DELETE CASCADE,
    FOREIGN KEY (workspace_id, user_id) REFERENCES workspace_members(workspace_id, user_id) ON DELETE CASCADE
);
CREATE INDEX team_members_ws_user_idx ON team_members (workspace_id, user_id);

-- +goose Down
DROP TABLE team_members;
