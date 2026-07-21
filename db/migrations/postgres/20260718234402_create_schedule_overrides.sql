-- +goose Up
CREATE TABLE schedule_overrides (
    id                 uuid PRIMARY KEY DEFAULT uuidv7(),
    schedule_id        uuid NOT NULL,
    workspace_id       uuid NOT NULL,
    user_id            uuid NOT NULL,
    starts_at          timestamptz NOT NULL,
    ends_at            timestamptz NOT NULL,
    reason             text NOT NULL DEFAULT '',
    created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    created_at         timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT schedule_overrides_window CHECK (ends_at > starts_at),
    FOREIGN KEY (schedule_id, workspace_id) REFERENCES schedules(id, workspace_id) ON DELETE CASCADE,
    FOREIGN KEY (workspace_id, user_id) REFERENCES workspace_members(workspace_id, user_id) ON DELETE CASCADE
);
CREATE INDEX schedule_overrides_schedule_idx ON schedule_overrides (schedule_id, starts_at);

-- +goose Down
DROP TABLE schedule_overrides;
