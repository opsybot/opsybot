-- +goose Up
CREATE TABLE alert_source_mappings (
    id        uuid PRIMARY KEY DEFAULT uuidv7(),
    source_id uuid NOT NULL REFERENCES alert_sources(id) ON DELETE CASCADE,
    field     text NOT NULL,
    path      text NOT NULL,
    position  integer NOT NULL DEFAULT 0,
    CONSTRAINT alert_source_mappings_field_valid CHECK (
        field IN ('title', 'description', 'severity', 'service', 'source', 'dedup_key', 'status', 'starts_at', 'ends_at', 'labels')
    ),
    CONSTRAINT alert_source_mappings_path_nonempty CHECK (btrim(path) <> '')
);
CREATE UNIQUE INDEX alert_source_mappings_source_field_uq ON alert_source_mappings (source_id, field);

-- +goose Down
DROP TABLE alert_source_mappings;
