-- +goose Up
CREATE TABLE alert_escalations (
    id             uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id   uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    alert_id       uuid NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    policy_id      uuid NOT NULL REFERENCES escalation_policies(id) ON DELETE CASCADE,
    state          text NOT NULL DEFAULT 'running',
    cycle          integer NOT NULL DEFAULT 0,
    step_index     integer NOT NULL DEFAULT 0,
    plan           jsonb NOT NULL,
    next_at        timestamptz,
    acked_at       timestamptz,
    ack_expires_at timestamptz,
    started_at     timestamptz NOT NULL DEFAULT now(),
    ended_at       timestamptz,
    updated_at     timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT alert_escalations_state_valid CHECK (state IN ('running', 'acked', 'resolved', 'exhausted')),
    CONSTRAINT alert_escalations_counters CHECK (cycle >= 0 AND step_index >= 0)
);
CREATE UNIQUE INDEX alert_escalations_alert_uq ON alert_escalations (alert_id);
CREATE INDEX alert_escalations_due_idx ON alert_escalations (next_at) WHERE state IN ('running', 'acked');
CREATE INDEX alert_escalations_policy_started_idx ON alert_escalations (policy_id, started_at DESC);

CREATE TABLE escalation_rr_state (
    policy_id uuid NOT NULL REFERENCES escalation_policies(id) ON DELETE CASCADE,
    node_id   text NOT NULL,
    position  integer NOT NULL DEFAULT 0,
    PRIMARY KEY (policy_id, node_id)
);

-- +goose Down
DROP TABLE escalation_rr_state;
DROP TABLE alert_escalations;
