-- +goose Up
CREATE TABLE schedule_layers (
    id            uuid PRIMARY KEY DEFAULT uuidv7(),
    schedule_id   uuid NOT NULL,
    workspace_id  uuid NOT NULL,
    position      int NOT NULL,
    rotation      text NOT NULL,
    interval_days int NOT NULL DEFAULT 1,
    handover_hour int NOT NULL,
    starts_on     date NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT schedule_layers_rotation_valid CHECK (rotation IN ('daily', 'weekly', 'custom')),
    CONSTRAINT schedule_layers_interval_range CHECK (interval_days BETWEEN 1 AND 30),
    CONSTRAINT schedule_layers_handover_range CHECK (handover_hour BETWEEN 0 AND 23),
    FOREIGN KEY (schedule_id, workspace_id) REFERENCES schedules(id, workspace_id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX schedule_layers_schedule_position_uq ON schedule_layers (schedule_id, position);
CREATE UNIQUE INDEX schedule_layers_id_ws_uq ON schedule_layers (id, workspace_id);

-- +goose Down
DROP TABLE schedule_layers;
