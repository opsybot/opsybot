-- +goose Up
CREATE TABLE schedule_layer_restrictions (
    id           uuid PRIMARY KEY DEFAULT uuidv7(),
    layer_id     uuid NOT NULL REFERENCES schedule_layers(id) ON DELETE CASCADE,
    weekday      int NOT NULL,
    start_minute int NOT NULL,
    end_minute   int NOT NULL,
    CONSTRAINT schedule_layer_restrictions_weekday_range CHECK (weekday BETWEEN 0 AND 6),
    CONSTRAINT schedule_layer_restrictions_window CHECK (start_minute >= 0 AND end_minute <= 1440 AND start_minute < end_minute)
);
CREATE INDEX schedule_layer_restrictions_layer_idx ON schedule_layer_restrictions (layer_id);

-- +goose Down
DROP TABLE schedule_layer_restrictions;
