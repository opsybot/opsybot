-- +goose Up
CREATE UNIQUE INDEX alerts_open_group_parent_uq ON alerts (workspace_id, group_key)
    WHERE group_key <> '' AND parent_alert_id IS NULL AND resolved_at IS NULL;

-- +goose Down
DROP INDEX alerts_open_group_parent_uq;
