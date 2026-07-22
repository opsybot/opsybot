-- +goose Up
DELETE FROM alert_routes;
ALTER TABLE alert_routes DROP COLUMN policy_ref;
ALTER TABLE alert_routes ADD COLUMN escalation_policy_id uuid NOT NULL REFERENCES escalation_policies(id);

ALTER TABLE alert_settings DROP COLUMN default_policy_ref;
ALTER TABLE alert_settings ADD COLUMN default_escalation_policy_id uuid REFERENCES escalation_policies(id) ON DELETE SET NULL;

ALTER TABLE alert_monitors DROP COLUMN policy_ref;
ALTER TABLE alert_monitors ADD COLUMN escalation_policy_id uuid REFERENCES escalation_policies(id) ON DELETE SET NULL;

ALTER TABLE alerts DROP COLUMN routed_policy_ref;
ALTER TABLE alerts ADD COLUMN escalation_policy_id uuid REFERENCES escalation_policies(id) ON DELETE SET NULL;

ALTER TABLE alert_events DROP CONSTRAINT alert_events_kind_valid;
ALTER TABLE alert_events ADD CONSTRAINT alert_events_kind_valid CHECK (
    kind IN ('received', 'deduped', 'grouped', 'routed', 'suppressed', 'escalation', 'notified', 'push', 'sms', 'timeout', 'chat', 'acked', 'resolved', 'exhausted')
);

-- +goose Down
ALTER TABLE alert_events DROP CONSTRAINT alert_events_kind_valid;
ALTER TABLE alert_events ADD CONSTRAINT alert_events_kind_valid CHECK (
    kind IN ('received', 'deduped', 'grouped', 'routed', 'suppressed', 'escalation', 'push', 'sms', 'timeout', 'chat', 'acked', 'resolved')
);
ALTER TABLE alerts DROP COLUMN escalation_policy_id;
ALTER TABLE alerts ADD COLUMN routed_policy_ref text NOT NULL DEFAULT '';
ALTER TABLE alert_monitors DROP COLUMN escalation_policy_id;
ALTER TABLE alert_monitors ADD COLUMN policy_ref text NOT NULL DEFAULT '';
ALTER TABLE alert_settings DROP COLUMN default_escalation_policy_id;
ALTER TABLE alert_settings ADD COLUMN default_policy_ref text NOT NULL DEFAULT 'platform-default';
ALTER TABLE alert_routes DROP COLUMN escalation_policy_id;
ALTER TABLE alert_routes ADD COLUMN policy_ref text NOT NULL DEFAULT 'platform-default';
