-- +goose Up
CREATE TABLE notification_runs (
    id               uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id     uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    alert_id         uuid NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    user_id          uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    escalation_id    uuid REFERENCES alert_escalations(id) ON DELETE SET NULL,
    escalation_cycle integer NOT NULL DEFAULT 0,
    level            integer NOT NULL DEFAULT 0,
    policy_slug      text NOT NULL DEFAULT '',
    label            text NOT NULL DEFAULT '',
    urgency          text NOT NULL,
    state            text NOT NULL DEFAULT 'running',
    stop_reason      text NOT NULL DEFAULT '',
    step_index       integer NOT NULL DEFAULT 0,
    plan             jsonb NOT NULL,
    next_at          timestamptz,
    started_at       timestamptz NOT NULL DEFAULT now(),
    ended_at         timestamptz,
    updated_at       timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT notification_runs_state_valid CHECK (state IN ('running', 'stopped', 'exhausted')),
    CONSTRAINT notification_runs_urgency_valid CHECK (urgency IN ('high', 'low')),
    CONSTRAINT notification_runs_reason_valid CHECK (stop_reason IN ('', 'acked', 'resolved', 'superseded')),
    CONSTRAINT notification_runs_counters CHECK (step_index >= 0 AND level >= 0 AND escalation_cycle >= 0)
);
CREATE UNIQUE INDEX notification_runs_page_uq ON notification_runs (alert_id, user_id, level, escalation_cycle);
CREATE INDEX notification_runs_due_idx ON notification_runs (next_at) WHERE state = 'running';
CREATE INDEX notification_runs_alert_idx ON notification_runs (alert_id, started_at DESC);

CREATE TABLE notification_attempts (
    id                  uuid PRIMARY KEY DEFAULT uuidv7(),
    run_id              uuid NOT NULL REFERENCES notification_runs(id) ON DELETE CASCADE,
    workspace_id        uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    alert_id            uuid NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    user_id             uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    step_index          integer NOT NULL,
    channel_type        text NOT NULL,
    channel_id          uuid REFERENCES user_channels(id) ON DELETE SET NULL,
    detail              text NOT NULL DEFAULT '',
    outcome             text NOT NULL,
    provider_message_id text NOT NULL DEFAULT '',
    error_detail        text NOT NULL DEFAULT '',
    at                  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT notification_attempts_outcome_valid CHECK (outcome IN ('delivered', 'failed', 'suppressed', 'skipped', 'throttled')),
    CONSTRAINT notification_attempts_channel_valid CHECK (channel_type IN ('slack', 'teams', 'discord', 'telegram', 'ntfy', 'email', 'webhook')),
    CONSTRAINT notification_attempts_step_index CHECK (step_index >= 0)
);
CREATE INDEX notification_attempts_alert_at_idx ON notification_attempts (alert_id, at);
CREATE INDEX notification_attempts_run_idx ON notification_attempts (run_id, step_index);

-- +goose Down
DROP TABLE notification_attempts;
DROP TABLE notification_runs;
