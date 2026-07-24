-- +goose Up
ALTER TABLE incident_events ADD COLUMN workspace_id uuid;
UPDATE incident_events e SET workspace_id = i.workspace_id FROM incidents i WHERE i.id = e.incident_id;
ALTER TABLE incident_events ALTER COLUMN workspace_id SET NOT NULL;

ALTER TABLE incident_events DROP CONSTRAINT incident_events_incident_id_fkey;
ALTER TABLE incident_events ADD CONSTRAINT incident_events_incident_fkey
    FOREIGN KEY (incident_id, workspace_id) REFERENCES incidents(id, workspace_id) ON DELETE CASCADE;

ALTER TABLE incident_events ADD COLUMN category text NOT NULL DEFAULT 'status';
ALTER TABLE incident_events ADD COLUMN source text NOT NULL DEFAULT 'system';
ALTER TABLE incident_events ADD COLUMN retroactive boolean NOT NULL DEFAULT false;
ALTER TABLE incident_events ADD COLUMN actor_user_id uuid REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE incident_events ADD COLUMN edited_at timestamptz;
ALTER TABLE incident_events ADD COLUMN edited_by uuid REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE incident_events ADD COLUMN idempotency_key text NOT NULL DEFAULT '';

UPDATE incident_events SET category = 'action'
 WHERE kind IN ('alert_linked', 'alert_unlinked', 'related');

ALTER TABLE incident_events ADD CONSTRAINT incident_events_category_valid CHECK (
    category IN ('status', 'communication', 'action', 'observation', 'decision')
);
ALTER TABLE incident_events ADD CONSTRAINT incident_events_source_valid CHECK (
    source IN ('system', 'ui', 'api', 'chat')
);

ALTER TABLE incident_events DROP CONSTRAINT incident_events_kind_valid;
ALTER TABLE incident_events ADD CONSTRAINT incident_events_kind_valid CHECK (
    kind IN ('declared', 'status_changed', 'severity_changed', 'lead_changed', 'renamed', 'summary_changed', 'fields_changed', 'reopened', 'resolved', 'alert_linked', 'alert_unlinked', 'related', 'unrelated', 'followup_added', 'followup_done', 'updated', 'note')
);

CREATE UNIQUE INDEX incident_events_id_ws_uq ON incident_events (id, workspace_id);
CREATE INDEX incident_events_incident_order_idx ON incident_events (incident_id, at, id);
CREATE UNIQUE INDEX incident_events_idem_uq ON incident_events (incident_id, idempotency_key)
    WHERE idempotency_key <> '';

CREATE TABLE incident_event_revisions (
    id             uuid PRIMARY KEY DEFAULT uuidv7(),
    event_id       uuid NOT NULL,
    workspace_id   uuid NOT NULL,
    at             timestamptz NOT NULL DEFAULT now(),
    editor_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    editor_label   text NOT NULL DEFAULT '',
    text           text NOT NULL DEFAULT '',
    category       text NOT NULL DEFAULT 'status',
    FOREIGN KEY (event_id, workspace_id) REFERENCES incident_events(id, workspace_id) ON DELETE CASCADE
);
CREATE INDEX incident_event_revisions_event_at_idx ON incident_event_revisions (event_id, at);

CREATE TABLE incident_event_attachments (
    id            uuid PRIMARY KEY DEFAULT uuidv7(),
    event_id      uuid NOT NULL,
    workspace_id  uuid NOT NULL,
    kind          text NOT NULL,
    label         text NOT NULL DEFAULT '',
    url           text NOT NULL DEFAULT '',
    body          text NOT NULL DEFAULT '',
    object_key    text NOT NULL DEFAULT '',
    content_type  text NOT NULL DEFAULT '',
    size_bytes    bigint NOT NULL DEFAULT 0,
    created_at    timestamptz NOT NULL DEFAULT now(),
    created_by    uuid REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT incident_event_attachments_kind_valid CHECK (kind IN ('image', 'log', 'link')),
    FOREIGN KEY (event_id, workspace_id) REFERENCES incident_events(id, workspace_id) ON DELETE CASCADE
);
CREATE INDEX incident_event_attachments_event_idx ON incident_event_attachments (event_id, created_at);

-- +goose Down
DROP TABLE incident_event_attachments;
DROP TABLE incident_event_revisions;

DROP INDEX incident_events_idem_uq;
DROP INDEX incident_events_incident_order_idx;
DROP INDEX incident_events_id_ws_uq;

ALTER TABLE incident_events DROP CONSTRAINT incident_events_kind_valid;
ALTER TABLE incident_events ADD CONSTRAINT incident_events_kind_valid CHECK (
    kind IN ('declared', 'status_changed', 'severity_changed', 'lead_changed', 'renamed', 'summary_changed', 'fields_changed', 'reopened', 'resolved', 'alert_linked', 'alert_unlinked', 'related', 'followup_added', 'updated')
);

ALTER TABLE incident_events DROP CONSTRAINT incident_events_source_valid;
ALTER TABLE incident_events DROP CONSTRAINT incident_events_category_valid;

ALTER TABLE incident_events DROP COLUMN idempotency_key;
ALTER TABLE incident_events DROP COLUMN edited_by;
ALTER TABLE incident_events DROP COLUMN edited_at;
ALTER TABLE incident_events DROP COLUMN actor_user_id;
ALTER TABLE incident_events DROP COLUMN retroactive;
ALTER TABLE incident_events DROP COLUMN source;
ALTER TABLE incident_events DROP COLUMN category;

ALTER TABLE incident_events DROP CONSTRAINT incident_events_incident_fkey;
ALTER TABLE incident_events ADD CONSTRAINT incident_events_incident_id_fkey
    FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE;

ALTER TABLE incident_events DROP COLUMN workspace_id;
