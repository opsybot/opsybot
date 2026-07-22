-- +goose Up
CREATE TABLE alert_routes (
    id           uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    position     integer NOT NULL DEFAULT 0,
    policy_ref   text NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT alert_routes_policy_ref_nonempty CHECK (btrim(policy_ref) <> '')
);
CREATE INDEX alert_routes_ws_position_idx ON alert_routes (workspace_id, position);

CREATE TABLE alert_route_conditions (
    id       uuid PRIMARY KEY DEFAULT uuidv7(),
    route_id uuid NOT NULL REFERENCES alert_routes(id) ON DELETE CASCADE,
    field    text NOT NULL,
    op       text NOT NULL,
    value    text NOT NULL,
    position integer NOT NULL DEFAULT 0,
    CONSTRAINT alert_route_conditions_op_valid CHECK (op IN ('is', 'is not', 'contains', 'matches')),
    CONSTRAINT alert_route_conditions_value_nonempty CHECK (btrim(value) <> '')
);
CREATE INDEX alert_route_conditions_route_idx ON alert_route_conditions (route_id, position);

-- +goose Down
DROP TABLE alert_route_conditions;
DROP TABLE alert_routes;
