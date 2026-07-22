-- +goose Up
CREATE TABLE escalation_policies (
    id                  uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id        uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    slug                text NOT NULL,
    name                text NOT NULL,
    team_id             uuid REFERENCES teams(id) ON DELETE SET NULL,
    repeat_count        integer NOT NULL DEFAULT 0,
    ack_timeout_seconds integer NOT NULL DEFAULT 0,
    definition          jsonb NOT NULL,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT escalation_policies_slug_format CHECK (slug ~ '^[a-z][a-z0-9-]{0,39}$'),
    CONSTRAINT escalation_policies_name_nonempty CHECK (btrim(name) <> ''),
    CONSTRAINT escalation_policies_repeat_range CHECK (repeat_count BETWEEN 0 AND 3),
    CONSTRAINT escalation_policies_ack_timeout_range CHECK (ack_timeout_seconds BETWEEN 0 AND 86400)
);
CREATE UNIQUE INDEX escalation_policies_ws_slug_uq ON escalation_policies (workspace_id, slug);

-- +goose Down
DROP TABLE escalation_policies;
