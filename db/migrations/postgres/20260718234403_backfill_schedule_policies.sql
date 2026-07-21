-- +goose Up
INSERT INTO casbin_rule (p_type, v0, v1, v2, v3)
SELECT 'p', r.role, w.id::text, 'schedules', r.action
FROM workspaces w
CROSS JOIN (VALUES ('admin', 'read'), ('admin', 'write'), ('member', 'read')) AS r(role, action)
WHERE NOT EXISTS (
    SELECT 1 FROM casbin_rule c
    WHERE c.p_type = 'p' AND c.v0 = r.role AND c.v1 = w.id::text AND c.v2 = 'schedules' AND c.v3 = r.action
);

-- +goose Down
DELETE FROM casbin_rule WHERE p_type = 'p' AND v2 = 'schedules';
