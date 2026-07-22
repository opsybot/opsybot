-- +goose Up
INSERT INTO casbin_rule (p_type, v0, v1, v2, v3)
SELECT 'p', r.role, w.id::text, r.object, r.action
FROM workspaces w
CROSS JOIN (VALUES
    ('admin', 'alerts', 'read'), ('admin', 'alerts', 'write'),
    ('admin', 'alert_sources', 'read'), ('admin', 'alert_sources', 'write'),
    ('member', 'alerts', 'read'), ('member', 'alerts', 'write'),
    ('member', 'alert_sources', 'read')
) AS r(role, object, action)
WHERE NOT EXISTS (
    SELECT 1 FROM casbin_rule c
    WHERE c.p_type = 'p' AND c.v0 = r.role AND c.v1 = w.id::text AND c.v2 = r.object AND c.v3 = r.action
);

-- +goose Down
DELETE FROM casbin_rule WHERE p_type = 'p' AND v2 IN ('alerts', 'alert_sources');
