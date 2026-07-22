-- +goose Up
CREATE TABLE escalation_policy_targets (
    id          uuid PRIMARY KEY DEFAULT uuidv7(),
    policy_id   uuid NOT NULL REFERENCES escalation_policies(id) ON DELETE CASCADE,
    node_id     text NOT NULL,
    target_type text NOT NULL,
    target_ref  uuid NOT NULL,
    CONSTRAINT escalation_policy_targets_type_valid CHECK (target_type IN ('person', 'schedule', 'team', 'webhook'))
);
CREATE INDEX escalation_policy_targets_policy_idx ON escalation_policy_targets (policy_id);
CREATE INDEX escalation_policy_targets_ref_idx ON escalation_policy_targets (target_type, target_ref);

-- +goose Down
DROP TABLE escalation_policy_targets;
