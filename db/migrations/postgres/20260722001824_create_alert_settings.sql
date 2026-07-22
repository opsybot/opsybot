-- +goose Up
CREATE TABLE alert_settings (
    workspace_id       uuid PRIMARY KEY REFERENCES workspaces(id) ON DELETE CASCADE,
    default_policy_ref text NOT NULL DEFAULT 'platform-default',
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT alert_settings_default_policy_nonempty CHECK (btrim(default_policy_ref) <> '')
);

-- +goose Down
DROP TABLE alert_settings;
