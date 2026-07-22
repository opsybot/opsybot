-- +goose Up
CREATE TABLE alert_silences (
    id            uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id  uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    kind          text NOT NULL DEFAULT 'silence',
    reason        text NOT NULL DEFAULT '',
    created_by    text NOT NULL DEFAULT '',
    created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    starts_at     timestamptz NOT NULL,
    ends_at       timestamptz NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT alert_silences_kind_valid CHECK (kind IN ('silence', 'maintenance')),
    CONSTRAINT alert_silences_window_ordered CHECK (ends_at > starts_at)
);
CREATE INDEX alert_silences_ws_window_idx ON alert_silences (workspace_id, starts_at, ends_at);

CREATE TABLE alert_silence_conditions (
    id         uuid PRIMARY KEY DEFAULT uuidv7(),
    silence_id uuid NOT NULL REFERENCES alert_silences(id) ON DELETE CASCADE,
    field      text NOT NULL,
    value      text NOT NULL,
    position   integer NOT NULL DEFAULT 0,
    CONSTRAINT alert_silence_conditions_field_valid CHECK (field IN ('source', 'service', 'label')),
    CONSTRAINT alert_silence_conditions_value_nonempty CHECK (btrim(value) <> '')
);
CREATE INDEX alert_silence_conditions_silence_idx ON alert_silence_conditions (silence_id, position);

ALTER TABLE alerts
    ADD CONSTRAINT alerts_suppressed_by_silence_fk
    FOREIGN KEY (suppressed_by_silence_id) REFERENCES alert_silences(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE alerts DROP CONSTRAINT alerts_suppressed_by_silence_fk;
DROP TABLE alert_silence_conditions;
DROP TABLE alert_silences;
