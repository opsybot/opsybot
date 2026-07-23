-- +goose Up
ALTER TABLE notification_runs ADD COLUMN step_attempts integer NOT NULL DEFAULT 0;
ALTER TABLE notification_runs ADD COLUMN leased_until timestamptz;

ALTER TABLE notification_attempts DROP CONSTRAINT notification_attempts_outcome_valid;
ALTER TABLE notification_attempts ADD CONSTRAINT notification_attempts_outcome_valid
    CHECK (outcome IN ('delivered', 'accepted', 'failed', 'suppressed', 'skipped', 'throttled'));

-- +goose Down
ALTER TABLE notification_attempts DROP CONSTRAINT notification_attempts_outcome_valid;
ALTER TABLE notification_attempts ADD CONSTRAINT notification_attempts_outcome_valid
    CHECK (outcome IN ('delivered', 'failed', 'suppressed', 'skipped', 'throttled'));

ALTER TABLE notification_runs DROP COLUMN leased_until;
ALTER TABLE notification_runs DROP COLUMN step_attempts;
