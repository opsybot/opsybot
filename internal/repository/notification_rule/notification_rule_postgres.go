package notification_rule

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/types"

	dbpostgres "github.com/opsybot/opsybot/internal/db/postgres"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.NotificationRule {
	return &repo{db: db}
}

func (r *repo) Get(ctx context.Context, workspaceID, userID string) (entity.NotificationRule, error) {
	m, err := dbpostgres.UserNotificationRules(
		qm.Where("workspace_id = ? AND user_id = ?", workspaceID, userID),
	).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.NotificationRule{}, entity.ErrNotificationRuleNotFound
		}
		return entity.NotificationRule{}, fmt.Errorf("get notification rule: %w", err)
	}
	steps, err := dbpostgres.UserNotificationRuleSteps(
		qm.Where("rule_id = ?", m.ID),
		qm.OrderBy("lane, position"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return entity.NotificationRule{}, fmt.Errorf("get notification rule steps: %w", err)
	}
	return toEntity(m, steps), nil
}

func (r *repo) Save(ctx context.Context, rule entity.NotificationRule) (entity.NotificationRule, error) {
	exec := r.db.Querier(ctx)
	m := &dbpostgres.UserNotificationRule{
		WorkspaceID:      rule.WorkspaceID,
		UserID:           rule.UserID,
		QuietEnabled:     rule.QuietHours.Enabled,
		QuietDays:        toInt64Array(rule.QuietHours.Window.Days),
		QuietStartMinute: rule.QuietHours.Window.StartMinute,
		QuietEndMinute:   rule.QuietHours.Window.EndMinute,
		QuietTimezone:    quietTimezone(rule.QuietHours.Window.Timezone),
		UpdatedAt:        time.Now().UTC(),
	}
	if err := m.Upsert(ctx, exec, true,
		[]string{"workspace_id", "user_id"},
		boil.Whitelist("quiet_enabled", "quiet_days", "quiet_start_minute", "quiet_end_minute", "quiet_timezone", "updated_at"),
		boil.Infer()); err != nil {
		return entity.NotificationRule{}, fmt.Errorf("save notification rule: %w", err)
	}
	if _, err := dbpostgres.UserNotificationRuleSteps(qm.Where("rule_id = ?", m.ID)).DeleteAll(ctx, exec); err != nil {
		return entity.NotificationRule{}, fmt.Errorf("clear notification rule steps: %w", err)
	}
	if err := r.insertSteps(ctx, exec, m.ID, string(entity.NotifyUrgencyHigh), rule.High); err != nil {
		return entity.NotificationRule{}, err
	}
	if err := r.insertSteps(ctx, exec, m.ID, string(entity.NotifyUrgencyLow), rule.Low); err != nil {
		return entity.NotificationRule{}, err
	}
	return r.Get(ctx, rule.WorkspaceID, rule.UserID)
}

func (r *repo) insertSteps(ctx context.Context, exec boil.ContextExecutor, ruleID, lane string, steps []entity.NotificationStep) error {
	for i, step := range steps {
		row := &dbpostgres.UserNotificationRuleStep{
			RuleID:       ruleID,
			Lane:         lane,
			Position:     i,
			ChannelType:  string(step.Channel),
			DelaySeconds: int(step.Delay / time.Second),
		}
		if err := row.Insert(ctx, exec, boil.Whitelist("rule_id", "lane", "position", "channel_type", "delay_seconds")); err != nil {
			return fmt.Errorf("insert notification rule step: %w", err)
		}
	}
	return nil
}

func (r *repo) DeleteByUser(ctx context.Context, workspaceID, userID string) error {
	_, err := dbpostgres.UserNotificationRules(
		qm.Where("workspace_id = ? AND user_id = ?", workspaceID, userID),
	).DeleteAll(ctx, r.db.Querier(ctx))
	if err != nil {
		return fmt.Errorf("delete notification rule: %w", err)
	}
	return nil
}

func toEntity(m *dbpostgres.UserNotificationRule, steps dbpostgres.UserNotificationRuleStepSlice) entity.NotificationRule {
	rule := entity.NotificationRule{
		WorkspaceID: m.WorkspaceID,
		UserID:      m.UserID,
		QuietHours: entity.QuietHours{
			Enabled: m.QuietEnabled,
			Window: entity.HoursWindow{
				Days:        fromInt64Array(m.QuietDays),
				StartMinute: m.QuietStartMinute,
				EndMinute:   m.QuietEndMinute,
				Timezone:    m.QuietTimezone,
			},
		},
		UpdatedAt: m.UpdatedAt,
	}
	for _, s := range steps {
		step := entity.NotificationStep{
			Channel: entity.ChannelType(s.ChannelType),
			Delay:   time.Duration(s.DelaySeconds) * time.Second,
		}
		if s.Lane == string(entity.NotifyUrgencyHigh) {
			rule.High = append(rule.High, step)
		} else {
			rule.Low = append(rule.Low, step)
		}
	}
	return rule
}

func quietTimezone(tz string) string {
	if tz == "" {
		return "UTC"
	}
	return tz
}

func toInt64Array(days []int) types.Int64Array {
	out := make(types.Int64Array, len(days))
	for i, d := range days {
		out[i] = int64(d)
	}
	return out
}

func fromInt64Array(days types.Int64Array) []int {
	out := make([]int, len(days))
	for i, d := range days {
		out[i] = int(d)
	}
	return out
}
