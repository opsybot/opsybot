-- +goose Up
CREATE TABLE workspace_members (
    workspace_id   uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id        uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status         text NOT NULL DEFAULT 'invited',
    joined_at      timestamptz,
    deactivated_at timestamptz,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (workspace_id, user_id),
    CONSTRAINT workspace_members_status CHECK (status IN ('invited','active','deactivated')),
    CONSTRAINT workspace_members_deactivated_pair CHECK ((status = 'deactivated') = (deactivated_at IS NOT NULL)),
    CONSTRAINT workspace_members_joined_pair CHECK (status = 'invited' OR joined_at IS NOT NULL)
);
CREATE INDEX workspace_members_user_idx ON workspace_members (user_id);
CREATE INDEX workspace_members_ws_active_idx ON workspace_members (workspace_id) WHERE status = 'active';

-- +goose Down
DROP TABLE workspace_members;
