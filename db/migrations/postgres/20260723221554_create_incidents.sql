-- +goose Up
CREATE TABLE incident_severities (
    id           uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    level        text NOT NULL,
    label        text NOT NULL,
    definition   text NOT NULL DEFAULT '',
    tone         text NOT NULL DEFAULT 'neutral',
    position     integer NOT NULL DEFAULT 0,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT incident_severities_level_format CHECK (level ~ '^[A-Z][A-Z0-9]{0,11}$'),
    CONSTRAINT incident_severities_label_nonempty CHECK (btrim(label) <> '')
);
CREATE UNIQUE INDEX incident_severities_ws_level_uq ON incident_severities (workspace_id, level);

CREATE TABLE incident_field_defs (
    id           uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    slug         text NOT NULL,
    name         text NOT NULL,
    kind         text NOT NULL,
    options      jsonb NOT NULL DEFAULT '[]'::jsonb,
    position     integer NOT NULL DEFAULT 0,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT incident_field_defs_kind_valid CHECK (kind IN ('text', 'select', 'multi_select', 'number')),
    CONSTRAINT incident_field_defs_slug_format CHECK (slug ~ '^[a-z][a-z0-9-]{0,39}$'),
    CONSTRAINT incident_field_defs_name_nonempty CHECK (btrim(name) <> '')
);
CREATE UNIQUE INDEX incident_field_defs_ws_slug_uq ON incident_field_defs (workspace_id, slug);

CREATE TABLE incidents (
    id                 uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id       uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    number             integer NOT NULL,
    name               text NOT NULL,
    summary            text NOT NULL DEFAULT '',
    severity_level     text NOT NULL,
    status             text NOT NULL DEFAULT 'declared',
    lead_user_id       uuid REFERENCES users(id) ON DELETE SET NULL,
    team_id            uuid REFERENCES teams(id) ON DELETE SET NULL,
    custom_fields      jsonb NOT NULL DEFAULT '{}'::jsonb,
    resolution_summary text NOT NULL DEFAULT '',
    declared_by        uuid REFERENCES users(id) ON DELETE SET NULL,
    declared_at        timestamptz NOT NULL DEFAULT now(),
    resolved_at        timestamptz,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT incidents_status_valid CHECK (status IN ('declared', 'investigating', 'identified', 'monitoring', 'resolved')),
    CONSTRAINT incidents_name_nonempty CHECK (btrim(name) <> ''),
    CONSTRAINT incidents_number_positive CHECK (number >= 1),
    CONSTRAINT incidents_resolved_has_time CHECK ((status = 'resolved') = (resolved_at IS NOT NULL))
);
CREATE UNIQUE INDEX incidents_ws_number_uq ON incidents (workspace_id, number);
CREATE UNIQUE INDEX incidents_id_ws_uq ON incidents (id, workspace_id);
CREATE INDEX incidents_ws_declared_idx ON incidents (workspace_id, declared_at DESC, id DESC);
CREATE INDEX incidents_ws_status_idx ON incidents (workspace_id, status);

CREATE TABLE incident_services (
    incident_id  uuid NOT NULL,
    service_id   uuid NOT NULL,
    workspace_id uuid NOT NULL,
    PRIMARY KEY (incident_id, service_id),
    FOREIGN KEY (incident_id, workspace_id) REFERENCES incidents(id, workspace_id) ON DELETE CASCADE,
    FOREIGN KEY (service_id, workspace_id) REFERENCES services(id, workspace_id) ON DELETE CASCADE
);
CREATE INDEX incident_services_service_idx ON incident_services (service_id);

CREATE TABLE incident_alerts (
    incident_id  uuid NOT NULL,
    alert_id     uuid NOT NULL,
    workspace_id uuid NOT NULL,
    PRIMARY KEY (incident_id, alert_id),
    FOREIGN KEY (incident_id, workspace_id) REFERENCES incidents(id, workspace_id) ON DELETE CASCADE,
    FOREIGN KEY (alert_id, workspace_id) REFERENCES alerts(id, workspace_id) ON DELETE CASCADE
);
CREATE INDEX incident_alerts_alert_idx ON incident_alerts (alert_id);

CREATE TABLE incident_relations (
    id                  uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id        uuid NOT NULL,
    incident_id         uuid NOT NULL,
    related_incident_id uuid NOT NULL,
    kind                text NOT NULL,
    created_at          timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT incident_relations_kind_valid CHECK (kind IN ('related', 'duplicate', 'caused_by')),
    CONSTRAINT incident_relations_distinct CHECK (incident_id <> related_incident_id),
    FOREIGN KEY (incident_id, workspace_id) REFERENCES incidents(id, workspace_id) ON DELETE CASCADE,
    FOREIGN KEY (related_incident_id, workspace_id) REFERENCES incidents(id, workspace_id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX incident_relations_uq ON incident_relations (incident_id, related_incident_id, kind);

CREATE TABLE incident_followups (
    id            uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id  uuid NOT NULL,
    incident_id   uuid NOT NULL,
    title         text NOT NULL,
    owner_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    due_at        timestamptz,
    done          boolean NOT NULL DEFAULT false,
    done_at       timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT incident_followups_title_nonempty CHECK (btrim(title) <> ''),
    FOREIGN KEY (incident_id, workspace_id) REFERENCES incidents(id, workspace_id) ON DELETE CASCADE
);
CREATE INDEX incident_followups_ws_open_idx ON incident_followups (workspace_id, done, due_at);
CREATE INDEX incident_followups_incident_idx ON incident_followups (incident_id);

CREATE TABLE incident_events (
    id          uuid PRIMARY KEY DEFAULT uuidv7(),
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    at          timestamptz NOT NULL DEFAULT now(),
    kind        text NOT NULL,
    text        text NOT NULL DEFAULT '',
    actor       text NOT NULL DEFAULT '',
    CONSTRAINT incident_events_kind_valid CHECK (
        kind IN ('declared', 'status_changed', 'severity_changed', 'lead_changed', 'renamed', 'summary_changed', 'fields_changed', 'reopened', 'resolved', 'alert_linked', 'alert_unlinked', 'related', 'followup_added')
    )
);
CREATE INDEX incident_events_incident_at_idx ON incident_events (incident_id, at);

-- +goose Down
DROP TABLE incident_events;
DROP TABLE incident_followups;
DROP TABLE incident_relations;
DROP TABLE incident_alerts;
DROP TABLE incident_services;
DROP TABLE incidents;
DROP TABLE incident_field_defs;
DROP TABLE incident_severities;
