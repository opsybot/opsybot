-- +goose Up
INSERT INTO incident_severities (workspace_id, level, label, definition, tone, position)
SELECT w.id, s.level, s.label, s.definition, s.tone, s.position
FROM workspaces w
CROSS JOIN (VALUES
    ('SEV1', 'SEV1', 'Customer-facing outage or data loss. All hands, page immediately.', 'critical', 0),
    ('SEV2', 'SEV2', 'Major degradation for many customers. Page the on-call now.', 'high', 1),
    ('SEV3', 'SEV3', 'Partial or contained impact. Fix during working hours.', 'warning', 2),
    ('SEV4', 'SEV4', 'Minor issue, no customer impact yet. Track it.', 'info', 3)
) AS s(level, label, definition, tone, position)
WHERE NOT EXISTS (
    SELECT 1 FROM incident_severities i WHERE i.workspace_id = w.id AND i.level = s.level
);

INSERT INTO casbin_rule (p_type, v0, v1, v2, v3)
SELECT 'p', r.role, w.id::text, r.object, r.action
FROM workspaces w
CROSS JOIN (VALUES
    ('admin', 'incidents', 'read'), ('admin', 'incidents', 'write'),
    ('member', 'incidents', 'read'), ('member', 'incidents', 'write'),
    ('admin', 'services', 'read'), ('admin', 'services', 'write'),
    ('member', 'services', 'read')
) AS r(role, object, action)
WHERE NOT EXISTS (
    SELECT 1 FROM casbin_rule c
    WHERE c.p_type = 'p' AND c.v0 = r.role AND c.v1 = w.id::text AND c.v2 = r.object AND c.v3 = r.action
);

-- +goose Down
DELETE FROM casbin_rule WHERE p_type = 'p' AND v2 IN ('incidents', 'services');
DELETE FROM incident_severities;
