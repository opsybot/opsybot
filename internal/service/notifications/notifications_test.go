package notifications

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository/alert"
	"github.com/opsybot/opsybot/internal/repository/channel"
	"github.com/opsybot/opsybot/internal/repository/lock"
	"github.com/opsybot/opsybot/internal/repository/notification_rule"
	"github.com/opsybot/opsybot/internal/repository/notification_run"
	"github.com/opsybot/opsybot/internal/repository/workspace"
	"github.com/opsybot/opsybot/internal/service/notifier"
	"github.com/opsybot/opsybot/internal/service/ratelimiter"
)

type fakeTx struct{}

func (fakeTx) WithTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type harness struct {
	srv      *srv
	runs     *notification_run.MockNotificationRun
	rules    *notification_rule.MockNotificationRule
	channels *channel.MockChannel
	alerts   *alert.MockAlert
	ws       *workspace.MockWorkspace
	notify   *notifier.MockNotifier
	limiter  *ratelimiter.MockRateLimiter
	lock     *lock.MockLock
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	ctrl := gomock.NewController(t)
	h := &harness{
		runs:     notification_run.NewMockNotificationRun(ctrl),
		rules:    notification_rule.NewMockNotificationRule(ctrl),
		channels: channel.NewMockChannel(ctrl),
		alerts:   alert.NewMockAlert(ctrl),
		ws:       workspace.NewMockWorkspace(ctrl),
		notify:   notifier.NewMockNotifier(ctrl),
		limiter:  ratelimiter.NewMockRateLimiter(ctrl),
		lock:     lock.NewMockLock(ctrl),
	}
	h.srv = &srv{
		tx: fakeTx{}, lock: h.lock, runs: h.runs, rules: h.rules, channels: h.channels,
		alerts: h.alerts, workspaces: h.ws, notifier: h.notify, limiter: h.limiter,
		cfg: config.Auth{BaseURL: "http://localhost:8080"},
	}
	return h
}

func step(kind entity.ChannelType, delay time.Duration, detail string) entity.NotificationPlanStep {
	return entity.NotificationPlanStep{Channel: kind, Delay: delay, Detail: detail}
}

func runWith(steps []entity.NotificationPlanStep, index int, next time.Time) entity.NotificationRun {
	return entity.NotificationRun{
		ID: "nrun-1", WorkspaceID: "ws-1", AlertID: "al-1", UserID: "u1", Label: "Priya",
		State: entity.NotifyRunRunning, StepIndex: index, NextAt: next,
		Plan: entity.NotificationPlan{Steps: steps},
	}
}

func openAlert() entity.Alert {
	return entity.Alert{ID: "al-1", WorkspaceID: "ws-1", Status: entity.AlertStatusOpen, Severity: entity.SeverityCritical}
}

func TestPageCreatesRunFromRule(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()
	rule := entity.NotificationRule{High: []entity.NotificationStep{{Channel: entity.ChannelTypeNtfy}}}
	h.rules.EXPECT().Get(gomock.Any(), "ws-1", "u1").Return(rule, nil)
	h.channels.EXPECT().ListByUser(gomock.Any(), "u1").Return([]entity.Channel{
		{ID: "c1", Type: entity.ChannelTypeNtfy, Detail: "ntfy.sh/x", Verified: true},
	}, nil)
	h.runs.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, run entity.NotificationRun) (entity.NotificationRun, bool, error) {
			if len(run.Plan.Steps) != 1 || run.Plan.Steps[0].Channel != entity.ChannelTypeNtfy {
				t.Errorf("plan = %+v", run.Plan.Steps)
			}
			run.ID = "nrun-1"
			return run, true, nil
		})

	out, err := h.srv.Page(context.Background(), entity.NotifyRequest{
		WorkspaceID: "ws-1", AlertID: "al-1", UserID: "u1", Email: "p@acme.test",
		Severity: entity.SeverityCritical, At: now,
	})
	if err != nil || out.ID != "nrun-1" {
		t.Fatalf("page = %+v, %v", out, err)
	}
}

func TestPageFallsBackToDefaultRuleAndEmail(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()
	h.rules.EXPECT().Get(gomock.Any(), "ws-1", "u1").Return(entity.NotificationRule{}, entity.ErrNotificationRuleNotFound)
	h.channels.EXPECT().ListByUser(gomock.Any(), "u1").Return(nil, nil)
	h.runs.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, run entity.NotificationRun) (entity.NotificationRun, bool, error) {
			if len(run.Plan.Steps) != 1 || run.Plan.Steps[0].Channel != entity.ChannelTypeEmail || run.Plan.Steps[0].Detail != "p@acme.test" {
				t.Errorf("fallback plan = %+v", run.Plan.Steps)
			}
			return run, true, nil
		})

	if _, err := h.srv.Page(context.Background(), entity.NotifyRequest{
		WorkspaceID: "ws-1", AlertID: "al-1", UserID: "u1", Email: "p@acme.test", Severity: entity.SeverityCritical, At: now,
	}); err != nil {
		t.Fatalf("page: %v", err)
	}
}

func TestAdvanceSendsStepAndSchedulesNext(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()
	steps := []entity.NotificationPlanStep{step(entity.ChannelTypeEmail, 0, "p@acme.test"), step(entity.ChannelTypeEmail, 5*time.Minute, "p@acme.test")}
	run := runWith(steps, 0, now)

	h.runs.EXPECT().ListDue(gomock.Any(), now, entity.NotificationRunSweepBatch).Return([]entity.NotificationRun{run}, nil)
	h.lock.EXPECT().TryJob(gomock.Any(), "notify:nrun-1").Return(true, nil)
	h.runs.EXPECT().GetByID(gomock.Any(), "nrun-1").Return(run, nil)
	h.alerts.EXPECT().GetByID(gomock.Any(), "ws-1", "al-1").Return(openAlert(), nil)
	h.runs.EXPECT().SaveProgress(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, saved entity.NotificationRun) (bool, error) {
			if saved.StepIndex != 1 || !saved.NextAt.Equal(now.Add(5*time.Minute)) {
				t.Errorf("saved step %d next %v", saved.StepIndex, saved.NextAt)
			}
			return true, nil
		})
	h.ws.EXPECT().GetByID(gomock.Any(), "ws-1").Return(entity.Workspace{ID: "ws-1", Slug: "acme"}, nil)
	h.runs.EXPECT().GetByID(gomock.Any(), "nrun-1").Return(run, nil)
	h.limiter.EXPECT().Allow(gomock.Any(), entity.RateScopeNotify, "u1").Return(entity.RateResult{Allowed: true}, nil)
	h.notify.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).Return(entity.NotifyResult{Delivered: true, Detail: "email sent"})
	h.runs.EXPECT().AppendAttempt(gomock.Any(), gomock.Any()).Return(nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-1", gomock.Any()).Return(nil)

	advanced, err := h.srv.Advance(context.Background(), now)
	if err != nil || advanced != 1 {
		t.Fatalf("advance = %d, %v", advanced, err)
	}
}

func TestFailedDeliveryPullsNextStepForward(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()
	steps := []entity.NotificationPlanStep{step(entity.ChannelTypeWebhook, 0, "https://x"), step(entity.ChannelTypeEmail, 5*time.Minute, "p@acme.test")}
	run := runWith(steps, 0, now)

	h.runs.EXPECT().ListDue(gomock.Any(), now, entity.NotificationRunSweepBatch).Return([]entity.NotificationRun{run}, nil)
	h.lock.EXPECT().TryJob(gomock.Any(), "notify:nrun-1").Return(true, nil)
	h.runs.EXPECT().GetByID(gomock.Any(), "nrun-1").Return(run, nil)
	h.alerts.EXPECT().GetByID(gomock.Any(), "ws-1", "al-1").Return(openAlert(), nil)
	h.runs.EXPECT().SaveProgress(gomock.Any(), gomock.Any()).Return(true, nil)
	h.ws.EXPECT().GetByID(gomock.Any(), "ws-1").Return(entity.Workspace{ID: "ws-1", Slug: "acme"}, nil)
	h.runs.EXPECT().GetByID(gomock.Any(), "nrun-1").Return(run, nil)
	h.limiter.EXPECT().Allow(gomock.Any(), entity.RateScopeNotify, "u1").Return(entity.RateResult{Allowed: true}, nil)
	h.notify.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).Return(entity.NotifyResult{Detail: "connection refused"})
	h.runs.EXPECT().AppendAttempt(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, a entity.NotificationAttempt) error {
			if a.Outcome != entity.NotifyOutcomeFailed {
				t.Errorf("outcome = %s, want failed", a.Outcome)
			}
			return nil
		})
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-1", gomock.Any()).Return(nil)
	h.runs.EXPECT().Reschedule(gomock.Any(), "nrun-1", 1, now).Return(true, nil)

	if _, err := h.srv.Advance(context.Background(), now); err != nil {
		t.Fatalf("advance: %v", err)
	}
}

func TestQuietHoursSuppressesWithoutSending(t *testing.T) {
	h := newHarness(t)
	inWindow := time.Date(2026, 7, 22, 23, 30, 0, 0, time.UTC)
	steps := []entity.NotificationPlanStep{step(entity.ChannelTypeEmail, 0, "p@acme.test")}
	run := runWith(steps, 0, inWindow)
	run.Urgency = entity.NotifyUrgencyLow
	run.Plan.Urgency = entity.NotifyUrgencyLow
	run.Plan.QuietHours = entity.QuietHours{Enabled: true, Window: entity.HoursWindow{
		Days: []int{0, 1, 2, 3, 4, 5, 6}, StartMinute: 22 * 60, EndMinute: 7 * 60, Timezone: "UTC",
	}}

	h.runs.EXPECT().ListDue(gomock.Any(), inWindow, entity.NotificationRunSweepBatch).Return([]entity.NotificationRun{run}, nil)
	h.lock.EXPECT().TryJob(gomock.Any(), "notify:nrun-1").Return(true, nil)
	h.runs.EXPECT().GetByID(gomock.Any(), "nrun-1").Return(run, nil)
	h.alerts.EXPECT().GetByID(gomock.Any(), "ws-1", "al-1").Return(openAlert(), nil)
	h.runs.EXPECT().SaveProgress(gomock.Any(), gomock.Any()).Return(true, nil)
	h.runs.EXPECT().AppendAttempt(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, a entity.NotificationAttempt) error {
			if a.Outcome != entity.NotifyOutcomeSuppressed {
				t.Errorf("outcome = %s, want suppressed", a.Outcome)
			}
			return nil
		})
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-1", gomock.Any()).Return(nil)

	if _, err := h.srv.Advance(context.Background(), inWindow); err != nil {
		t.Fatalf("advance: %v", err)
	}
}

func TestDeliveryAbortsIfRunStoppedBetweenCasAndSend(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()
	steps := []entity.NotificationPlanStep{step(entity.ChannelTypeEmail, 0, "p@acme.test"), step(entity.ChannelTypeEmail, 5*time.Minute, "p@acme.test")}
	run := runWith(steps, 0, now)

	h.runs.EXPECT().ListDue(gomock.Any(), now, entity.NotificationRunSweepBatch).Return([]entity.NotificationRun{run}, nil)
	h.lock.EXPECT().TryJob(gomock.Any(), "notify:nrun-1").Return(true, nil)
	h.runs.EXPECT().GetByID(gomock.Any(), "nrun-1").Return(run, nil)
	h.alerts.EXPECT().GetByID(gomock.Any(), "ws-1", "al-1").Return(openAlert(), nil)
	h.runs.EXPECT().SaveProgress(gomock.Any(), gomock.Any()).Return(true, nil)
	h.ws.EXPECT().GetByID(gomock.Any(), "ws-1").Return(entity.Workspace{ID: "ws-1", Slug: "acme"}, nil)
	stopped := run
	stopped.State = entity.NotifyRunStopped
	stopped.StopReason = entity.NotifyStopAcked
	h.runs.EXPECT().GetByID(gomock.Any(), "nrun-1").Return(stopped, nil)
	h.runs.EXPECT().AppendAttempt(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, a entity.NotificationAttempt) error {
			if a.Outcome != entity.NotifyOutcomeSkipped {
				t.Errorf("outcome = %s, want skipped", a.Outcome)
			}
			return nil
		})
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-1", gomock.Any()).Return(nil)

	if _, err := h.srv.Advance(context.Background(), now); err != nil {
		t.Fatalf("advance: %v", err)
	}
}

func TestAdvanceStopsRunWhenAlertResolved(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()
	run := runWith([]entity.NotificationPlanStep{step(entity.ChannelTypeEmail, 0, "p@acme.test")}, 0, now)

	h.runs.EXPECT().ListDue(gomock.Any(), now, entity.NotificationRunSweepBatch).Return([]entity.NotificationRun{run}, nil)
	h.lock.EXPECT().TryJob(gomock.Any(), "notify:nrun-1").Return(true, nil)
	h.runs.EXPECT().GetByID(gomock.Any(), "nrun-1").Return(run, nil)
	resolved := openAlert()
	resolved.Status = entity.AlertStatusResolved
	h.alerts.EXPECT().GetByID(gomock.Any(), "ws-1", "al-1").Return(resolved, nil)
	h.runs.EXPECT().SaveProgress(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, saved entity.NotificationRun) (bool, error) {
			if saved.State != entity.NotifyRunStopped || saved.StopReason != entity.NotifyStopResolved {
				t.Errorf("saved state=%s reason=%s", saved.State, saved.StopReason)
			}
			return true, nil
		})

	if _, err := h.srv.Advance(context.Background(), now); err != nil {
		t.Fatalf("advance: %v", err)
	}
}

func TestAdvanceSkipsRunHeldByAnotherReplica(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()
	run := runWith([]entity.NotificationPlanStep{step(entity.ChannelTypeEmail, 0, "p@acme.test")}, 0, now)

	h.runs.EXPECT().ListDue(gomock.Any(), now, entity.NotificationRunSweepBatch).Return([]entity.NotificationRun{run}, nil)
	h.lock.EXPECT().TryJob(gomock.Any(), "notify:nrun-1").Return(false, nil)

	advanced, err := h.srv.Advance(context.Background(), now)
	if err != nil || advanced != 0 {
		t.Fatalf("advance = %d, %v; expected a skipped run to page nobody", advanced, err)
	}
}
