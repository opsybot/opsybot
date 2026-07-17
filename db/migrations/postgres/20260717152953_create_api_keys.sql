-- +goose Up
CREATE TABLE api_keys (
    id            uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id  uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    kind          text NOT NULL,
    owner_user_id uuid REFERENCES users(id) ON DELETE CASCADE,
    created_by    uuid REFERENCES users(id) ON DELETE SET NULL,
    name          text NOT NULL,
    token_hash    text NOT NULL UNIQUE,
    token_hint    text NOT NULL,
    scopes        text[] NOT NULL,
    last_used_at  timestamptz,
    revoked_at    timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT api_keys_kind CHECK (kind IN ('personal','workspace')),
    CONSTRAINT api_keys_owner_pair CHECK ((kind = 'personal') = (owner_user_id IS NOT NULL)),
    CONSTRAINT api_keys_scopes_nonempty CHECK (cardinality(scopes) > 0),
    CONSTRAINT api_keys_scopes_valid CHECK (scopes <@ ARRAY['incidents:read','incidents:write','alerts:read','alerts:write','schedules:write','policies:write','config:write','audit:read']::text[]),
    CONSTRAINT api_keys_name_nonempty CHECK (btrim(name) <> '')
);
CREATE INDEX api_keys_ws_kind_idx ON api_keys (workspace_id, kind) WHERE revoked_at IS NULL;
CREATE INDEX api_keys_owner_idx ON api_keys (owner_user_id) WHERE owner_user_id IS NOT NULL AND revoked_at IS NULL;

-- +goose Down
DROP TABLE api_keys;
