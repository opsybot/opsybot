-- +goose Up
CREATE TABLE escalation_webhooks (
    id           uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    slug         text NOT NULL,
    name         text NOT NULL,
    url          text NOT NULL,
    secret       bytea NOT NULL DEFAULT ''::bytea,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT escalation_webhooks_slug_format CHECK (slug ~ '^[a-z][a-z0-9-]{0,39}$'),
    CONSTRAINT escalation_webhooks_name_nonempty CHECK (btrim(name) <> ''),
    CONSTRAINT escalation_webhooks_url_nonempty CHECK (btrim(url) <> '')
);
CREATE UNIQUE INDEX escalation_webhooks_ws_slug_uq ON escalation_webhooks (workspace_id, slug);

-- +goose Down
DROP TABLE escalation_webhooks;
