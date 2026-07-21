-- +goose Up
CREATE TABLE schedule_layer_participants (
    layer_id     uuid NOT NULL,
    workspace_id uuid NOT NULL,
    user_id      uuid NOT NULL,
    position     int NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (layer_id, user_id),
    FOREIGN KEY (layer_id, workspace_id) REFERENCES schedule_layers(id, workspace_id) ON DELETE CASCADE,
    FOREIGN KEY (workspace_id, user_id) REFERENCES workspace_members(workspace_id, user_id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX schedule_layer_participants_position_uq ON schedule_layer_participants (layer_id, position);
CREATE INDEX schedule_layer_participants_ws_user_idx ON schedule_layer_participants (workspace_id, user_id);

-- +goose Down
DROP TABLE schedule_layer_participants;
