-- +goose Up
CREATE TABLE user_notification_rules (
    id                 uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id       uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id            uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quiet_enabled      boolean NOT NULL DEFAULT false,
    quiet_days         integer[] NOT NULL DEFAULT '{}',
    quiet_start_minute integer NOT NULL DEFAULT 1320,
    quiet_end_minute   integer NOT NULL DEFAULT 420,
    quiet_timezone     text NOT NULL DEFAULT 'UTC',
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_notification_rules_quiet_range CHECK (
        quiet_start_minute BETWEEN 0 AND 1439 AND quiet_end_minute BETWEEN 0 AND 1439),
    CONSTRAINT user_notification_rules_quiet_tz CHECK (btrim(quiet_timezone) <> '')
);
CREATE UNIQUE INDEX user_notification_rules_ws_user_uq ON user_notification_rules (workspace_id, user_id);

CREATE TABLE user_notification_rule_steps (
    id            uuid PRIMARY KEY DEFAULT uuidv7(),
    rule_id       uuid NOT NULL REFERENCES user_notification_rules(id) ON DELETE CASCADE,
    lane          text NOT NULL,
    position      integer NOT NULL,
    channel_type  text NOT NULL,
    delay_seconds integer NOT NULL DEFAULT 0,
    CONSTRAINT user_notification_rule_steps_lane_valid CHECK (lane IN ('high', 'low')),
    CONSTRAINT user_notification_rule_steps_channel_valid CHECK (channel_type IN ('slack', 'teams', 'discord', 'telegram', 'ntfy', 'email', 'webhook')),
    CONSTRAINT user_notification_rule_steps_delay_range CHECK (delay_seconds BETWEEN 0 AND 3600),
    CONSTRAINT user_notification_rule_steps_position CHECK (position >= 0 AND position < 12),
    CONSTRAINT user_notification_rule_steps_first_now CHECK (position > 0 OR delay_seconds = 0)
);
CREATE UNIQUE INDEX user_notification_rule_steps_uq ON user_notification_rule_steps (rule_id, lane, position);

-- +goose Down
DROP TABLE user_notification_rule_steps;
DROP TABLE user_notification_rules;
