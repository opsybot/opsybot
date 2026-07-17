-- +goose Up
CREATE TABLE invites (
    id           uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invited_by   uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    token_hash   text NOT NULL,
    status       text NOT NULL DEFAULT 'pending',
    expires_at   timestamptz NOT NULL,
    accepted_at  timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT invites_status CHECK (status IN ('pending','accepted','revoked'))
);
CREATE UNIQUE INDEX invites_token_hash_uq ON invites (token_hash);
CREATE UNIQUE INDEX invites_pending_uq ON invites (workspace_id, user_id) WHERE status = 'pending';

-- +goose Down
DROP TABLE invites;
