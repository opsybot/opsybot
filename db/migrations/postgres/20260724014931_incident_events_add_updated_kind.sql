-- +goose Up
ALTER TABLE incident_events DROP CONSTRAINT incident_events_kind_valid;
ALTER TABLE incident_events ADD CONSTRAINT incident_events_kind_valid CHECK (
    kind IN ('declared', 'status_changed', 'severity_changed', 'lead_changed', 'renamed', 'summary_changed', 'fields_changed', 'reopened', 'resolved', 'alert_linked', 'alert_unlinked', 'related', 'followup_added', 'updated')
);

-- +goose Down
ALTER TABLE incident_events DROP CONSTRAINT incident_events_kind_valid;
ALTER TABLE incident_events ADD CONSTRAINT incident_events_kind_valid CHECK (
    kind IN ('declared', 'status_changed', 'severity_changed', 'lead_changed', 'renamed', 'summary_changed', 'fields_changed', 'reopened', 'resolved', 'alert_linked', 'alert_unlinked', 'related', 'followup_added')
);
