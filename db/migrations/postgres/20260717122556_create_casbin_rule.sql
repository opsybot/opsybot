-- +goose Up
CREATE TABLE IF NOT EXISTS casbin_rule(
    p_type VARCHAR(32)  DEFAULT '' NOT NULL,
    v0     VARCHAR(255) DEFAULT '' NOT NULL,
    v1     VARCHAR(255) DEFAULT '' NOT NULL,
    v2     VARCHAR(255) DEFAULT '' NOT NULL,
    v3     VARCHAR(255) DEFAULT '' NOT NULL,
    v4     VARCHAR(255) DEFAULT '' NOT NULL,
    v5     VARCHAR(255) DEFAULT '' NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_casbin_rule ON casbin_rule (p_type,v0,v1);

-- +goose Down
DROP INDEX IF EXISTS idx_casbin_rule;
DROP TABLE IF EXISTS casbin_rule;
