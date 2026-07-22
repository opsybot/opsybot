package escalations

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository/alert"
	"github.com/opsybot/opsybot/internal/repository/alert_route"
	"github.com/opsybot/opsybot/internal/repository/escalation_policy"
	"github.com/opsybot/opsybot/internal/repository/escalation_run"
	"github.com/opsybot/opsybot/internal/repository/lock"
	"github.com/opsybot/opsybot/internal/repository/member"
	"github.com/opsybot/opsybot/internal/repository/schedule"
	"github.com/opsybot/opsybot/internal/repository/team"
	"github.com/opsybot/opsybot/internal/repository/workspace"
	"github.com/opsybot/opsybot/internal/service/notifier"
)

type fakeTx struct{}

func (fakeTx) WithTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type harness struct {
	srv      *srv
	runs     *escalation_run.MockEscalationRun
	policies *escalation_policy.MockEscalationPolicy
	alerts   *alert.MockAlert
	members  *member.MockMember
	teams    *team.MockTeam
	scheds   *schedule.MockSchedule
	notify   *notifier.MockNotifier
	lock     *lock.MockLock
	ws       *workspace.MockWorkspace
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	ctrl := gomock.NewController(t)
	h := &harness{
		runs:     escalation_run.NewMockEscalationRun(ctrl),
		policies: escalation_policy.NewMockEscalationPolicy(ctrl),
		alerts:   alert.NewMockAlert(ctrl),
		members:  member.NewMockMember(ctrl),
		teams:    team.NewMockTeam(ctrl),
		scheds:   schedule.NewMockSchedule(ctrl),
		notify:   notifier.NewMockNotifier(ctrl),
		lock:     lock.NewMockLock(ctrl),
		ws:       workspace.NewMockWorkspace(ctrl),
	}
	h.srv = &srv{
		tx: fakeTx{}, lock: h.lock, workspaces: h.ws, members: h.members, teams: h.teams,
		schedules: h.scheds, policies: h.policies, runs: h.runs, alerts: h.alerts,
		routes: alert_route.NewMockAlertRoute(ctrl), policy: nil, audit: nil,
		notifier: h.notify, cfg: config.Auth{BaseURL: "http://localhost:8080"},
	}
	return h
}

func (h *harness) emptyDirectory() {
	h.members.EXPECT().ListByWorkspace(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	h.scheds.EXPECT().ListByWorkspace(gomock.Any(), gomock.Any(), false).Return(nil, nil).AnyTimes()
	h.teams.EXPECT().ListByWorkspace(gomock.Any(), gomock.Any(), false).Return(nil, nil).AnyTimes()
	h.policies.EXPECT().ListWebhooks(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
}

func activeMember(userID, name string) entity.Member {
	return entity.Member{UserID: userID, Name: name, Email: name + "@acme.dev", Status: entity.MemberStatusActive}
}

func (h *harness) directoryWith(members ...entity.Member) {
	h.members.EXPECT().ListByWorkspace(gomock.Any(), gomock.Any()).Return(members, nil).AnyTimes()
	h.scheds.EXPECT().ListByWorkspace(gomock.Any(), gomock.Any(), false).Return(nil, nil).AnyTimes()
	h.teams.EXPECT().ListByWorkspace(gomock.Any(), gomock.Any(), false).Return(nil, nil).AnyTimes()
	h.policies.EXPECT().ListWebhooks(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
}

func runningRun(now time.Time) entity.EscalationRun {
	return entity.EscalationRun{
		ID: "run-1", WorkspaceID: "ws-1", AlertID: "al-1", PolicyID: "pol-1", PolicySlug: "payments-primary",
		State: entity.EscalationRunning,
		Path: []entity.EscalationLevel{
			{ID: "l1", Targets: []entity.EscalationTarget{{Type: entity.EscalationTargetPerson, Ref: "u1"}}, Mode: entity.NotifyModeAll, Wait: 5 * time.Minute},
		},
		Snapshot: entity.EscalationSnapshot{Repeat: 0},
		NextAt:   now,
	}
}

func TestStartCreatesRunOnceAndSkipsSuppressed(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()
	alertRow := entity.Alert{ID: "al-1", WorkspaceID: "ws-1", Severity: entity.SeverityCritical}
	policy := entity.EscalationPolicy{ID: "pol-1", Slug: "payments-primary", Nodes: []entity.EscalationNode{
		{Level: &entity.EscalationLevel{ID: "l1", Targets: []entity.EscalationTarget{{Type: entity.EscalationTargetPerson, Ref: "u1"}}, Mode: entity.NotifyModeAll, Wait: 5 * time.Minute}},
	}}

	h.policies.EXPECT().GetByID(gomock.Any(), "ws-1", "pol-1").Return(policy, nil).Times(2)
	h.runs.EXPECT().Create(gomock.Any(), gomock.Any()).Return(runningRun(now), true, nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-1", gomock.Any()).Return(nil)
	if err := h.srv.Start(context.Background(), alertRow, "pol-1"); err != nil {
		t.Fatalf("start: %v", err)
	}

	h.runs.EXPECT().Create(gomock.Any(), gomock.Any()).Return(runningRun(now), false, nil)
	if err := h.srv.Start(context.Background(), alertRow, "pol-1"); err != nil {
		t.Fatalf("second start: %v", err)
	}

	suppressed := alertRow
	suppressed.SuppressedBySilenceID = "sil-1"
	if err := h.srv.Start(context.Background(), suppressed, "pol-1"); err != nil {
		t.Fatalf("suppressed start: %v", err)
	}
}

func TestOnAckedSetsExpiryFromPolicySnapshot(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()

	withTimeout := runningRun(now)
	withTimeout.Snapshot.AckTimeout = 30 * time.Minute
	h.runs.EXPECT().GetByAlertID(gomock.Any(), "al-1").Return(withTimeout, nil)
	h.runs.EXPECT().MarkAcked(gomock.Any(), "al-1", now, now.Add(30*time.Minute)).Return(true, nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-1", gomock.Any()).Return(nil)

	if err := h.srv.OnAcked(context.Background(), "ws-1", []string{"al-1"}, now); err != nil {
		t.Fatalf("on acked: %v", err)
	}
}

func TestOnAckedWithoutTimeoutIsTerminal(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()

	h.runs.EXPECT().GetByAlertID(gomock.Any(), "al-1").Return(runningRun(now), nil)
	h.runs.EXPECT().MarkAcked(gomock.Any(), "al-1", now, time.Time{}).Return(true, nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-1", gomock.Any()).Return(nil)

	if err := h.srv.OnAcked(context.Background(), "ws-1", []string{"al-1"}, now); err != nil {
		t.Fatalf("on acked: %v", err)
	}
}

func TestOnAckedLostRaceWritesNoEvent(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()

	h.runs.EXPECT().GetByAlertID(gomock.Any(), "al-1").Return(runningRun(now), nil)
	h.runs.EXPECT().MarkAcked(gomock.Any(), "al-1", now, time.Time{}).Return(false, nil)

	if err := h.srv.OnAcked(context.Background(), "ws-1", []string{"al-1"}, now); err != nil {
		t.Fatalf("on acked: %v", err)
	}
}

func TestAdvanceNotifiesAndSchedulesNextStep(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()
	run := runningRun(now)

	h.directoryWith(activeMember("u1", "Priya"))
	h.runs.EXPECT().ListDue(gomock.Any(), now, entity.EscalationSweepBatch).Return([]entity.EscalationRun{run}, nil)
	h.lock.EXPECT().TryJob(gomock.Any(), "escalation:run-1").Return(true, nil)
	h.runs.EXPECT().GetByAlertID(gomock.Any(), "al-1").Return(run, nil)
	h.alerts.EXPECT().GetByID(gomock.Any(), "ws-1", "al-1").
		Return(entity.Alert{ID: "al-1", WorkspaceID: "ws-1", Status: entity.AlertStatusOpen, Title: "p99 high", Severity: entity.SeverityCritical}, nil)
	h.runs.EXPECT().SaveProgress(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, saved entity.EscalationRun) (bool, error) {
			if saved.StepIndex != 1 || !saved.NextAt.Equal(now.Add(5*time.Minute)) {
				t.Errorf("saved step %d next %v, want step 1 due in 5m", saved.StepIndex, saved.NextAt)
			}
			return true, nil
		})
	h.ws.EXPECT().GetByID(gomock.Any(), "ws-1").Return(entity.Workspace{ID: "ws-1", Slug: "acme"}, nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-1", gomock.Any()).Return(nil).Times(2)
	h.notify.EXPECT().PageUser(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, m entity.Member, page entity.AlertPage) entity.NotifyResult {
			if m.UserID != "u1" || page.Level != 1 || page.AlertURL != "http://localhost:8080/acme/alerts/al-1" {
				t.Errorf("paged %s level %d url %s", m.UserID, page.Level, page.AlertURL)
			}
			return entity.NotifyResult{Delivered: true, Detail: "email sent"}
		})

	advanced, err := h.srv.Advance(context.Background(), now)
	if err != nil || advanced != 1 {
		t.Fatalf("advance = %d, %v", advanced, err)
	}
}

func TestAdvanceTimeoutLosesRaceAgainstAck(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()
	run := runningRun(now)

	h.directoryWith(activeMember("u1", "Priya"))
	h.runs.EXPECT().ListDue(gomock.Any(), now, entity.EscalationSweepBatch).Return([]entity.EscalationRun{run}, nil)
	h.lock.EXPECT().TryJob(gomock.Any(), "escalation:run-1").Return(true, nil)
	h.runs.EXPECT().GetByAlertID(gomock.Any(), "al-1").Return(run, nil)
	h.alerts.EXPECT().GetByID(gomock.Any(), "ws-1", "al-1").
		Return(entity.Alert{ID: "al-1", WorkspaceID: "ws-1", Status: entity.AlertStatusOpen, Severity: entity.SeverityCritical}, nil)
	h.runs.EXPECT().SaveProgress(gomock.Any(), gomock.Any()).Return(false, nil)

	advanced, err := h.srv.Advance(context.Background(), now)
	if err != nil || advanced != 1 {
		t.Fatalf("advance = %d, %v", advanced, err)
	}
}

func TestAdvanceResumesExpiredAckAndReopensAlert(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()
	run := runningRun(now)
	run.State = entity.EscalationAcked
	run.AckExpiresAt = now.Add(-time.Second)

	h.runs.EXPECT().ListDue(gomock.Any(), now, entity.EscalationSweepBatch).Return([]entity.EscalationRun{run}, nil)
	h.lock.EXPECT().TryJob(gomock.Any(), "escalation:run-1").Return(true, nil)
	h.runs.EXPECT().GetByAlertID(gomock.Any(), "al-1").Return(run, nil)
	h.alerts.EXPECT().GetByID(gomock.Any(), "ws-1", "al-1").
		Return(entity.Alert{ID: "al-1", WorkspaceID: "ws-1", Status: entity.AlertStatusAcked}, nil)
	h.runs.EXPECT().Resume(gomock.Any(), "run-1", now).Return(true, nil)
	h.alerts.EXPECT().Reopen(gomock.Any(), "al-1", now).Return(true, nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-1", gomock.Any()).DoAndReturn(
		func(_ context.Context, _ string, ev entity.AlertEvent) error {
			if ev.Kind != entity.AlertEventTimeout {
				t.Errorf("event kind %s, want timeout", ev.Kind)
			}
			return nil
		})

	if _, err := h.srv.Advance(context.Background(), now); err != nil {
		t.Fatalf("advance: %v", err)
	}
}

func TestAdvanceClosesRunWhenAlertAlreadyResolved(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()
	run := runningRun(now)

	h.runs.EXPECT().ListDue(gomock.Any(), now, entity.EscalationSweepBatch).Return([]entity.EscalationRun{run}, nil)
	h.lock.EXPECT().TryJob(gomock.Any(), "escalation:run-1").Return(true, nil)
	h.runs.EXPECT().GetByAlertID(gomock.Any(), "al-1").Return(run, nil)
	h.alerts.EXPECT().GetByID(gomock.Any(), "ws-1", "al-1").
		Return(entity.Alert{ID: "al-1", WorkspaceID: "ws-1", Status: entity.AlertStatusResolved}, nil)
	h.runs.EXPECT().Finish(gomock.Any(), "run-1", entity.EscalationResolved, now).Return(true, nil)

	if _, err := h.srv.Advance(context.Background(), now); err != nil {
		t.Fatalf("advance: %v", err)
	}
}

func TestAdvanceExhaustsAfterFinalLevel(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()
	run := runningRun(now)
	run.StepIndex = 1

	h.runs.EXPECT().ListDue(gomock.Any(), now, entity.EscalationSweepBatch).Return([]entity.EscalationRun{run}, nil)
	h.lock.EXPECT().TryJob(gomock.Any(), "escalation:run-1").Return(true, nil)
	h.runs.EXPECT().GetByAlertID(gomock.Any(), "al-1").Return(run, nil)
	h.alerts.EXPECT().GetByID(gomock.Any(), "ws-1", "al-1").
		Return(entity.Alert{ID: "al-1", WorkspaceID: "ws-1", Status: entity.AlertStatusOpen}, nil)
	h.runs.EXPECT().Finish(gomock.Any(), "run-1", entity.EscalationExhausted, now).Return(true, nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-1", gomock.Any()).DoAndReturn(
		func(_ context.Context, _ string, ev entity.AlertEvent) error {
			if ev.Kind != entity.AlertEventExhausted {
				t.Errorf("event kind %s, want exhausted", ev.Kind)
			}
			return nil
		})

	if _, err := h.srv.Advance(context.Background(), now); err != nil {
		t.Fatalf("advance: %v", err)
	}
}

func TestAdvanceSkipsRunHeldByAnotherReplica(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()
	run := runningRun(now)

	h.runs.EXPECT().ListDue(gomock.Any(), now, entity.EscalationSweepBatch).Return([]entity.EscalationRun{run}, nil)
	h.lock.EXPECT().TryJob(gomock.Any(), "escalation:run-1").Return(false, nil)

	advanced, err := h.srv.Advance(context.Background(), now)
	if err != nil || advanced != 1 {
		t.Fatalf("advance = %d, %v: a held run is skipped, not an error", advanced, err)
	}
}

func TestNotifierFailureIsRecordedAndDoesNotStopEscalation(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC()
	run := runningRun(now)

	h.directoryWith(activeMember("u1", "Priya"))
	h.runs.EXPECT().ListDue(gomock.Any(), now, entity.EscalationSweepBatch).Return([]entity.EscalationRun{run}, nil)
	h.lock.EXPECT().TryJob(gomock.Any(), "escalation:run-1").Return(true, nil)
	h.runs.EXPECT().GetByAlertID(gomock.Any(), "al-1").Return(run, nil)
	h.alerts.EXPECT().GetByID(gomock.Any(), "ws-1", "al-1").
		Return(entity.Alert{ID: "al-1", WorkspaceID: "ws-1", Status: entity.AlertStatusOpen, Severity: entity.SeverityHigh}, nil)
	h.runs.EXPECT().SaveProgress(gomock.Any(), gomock.Any()).Return(true, nil)
	h.ws.EXPECT().GetByID(gomock.Any(), "ws-1").Return(entity.Workspace{ID: "ws-1", Slug: "acme"}, nil)
	h.notify.EXPECT().PageUser(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(entity.NotifyResult{Detail: "smtp connection refused"})

	failedRecorded := false
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-1", gomock.Any()).DoAndReturn(
		func(_ context.Context, _ string, ev entity.AlertEvent) error {
			if ev.Kind == entity.AlertEventNotified && ev.Result == "smtp connection refused" {
				failedRecorded = true
			}
			return nil
		}).Times(2)

	advanced, err := h.srv.Advance(context.Background(), now)
	if err != nil || advanced != 1 {
		t.Fatalf("advance = %d, %v", advanced, err)
	}
	if !failedRecorded {
		t.Fatal("delivery failure never landed on the timeline")
	}
}

func TestValidateRejectsUnknownTargets(t *testing.T) {
	h := newHarness(t)
	h.emptyDirectory()

	policy := entity.EscalationPolicy{
		WorkspaceID: "ws-1", Slug: "p", Name: "p",
		Nodes: []entity.EscalationNode{
			{Level: &entity.EscalationLevel{ID: "l1", Targets: []entity.EscalationTarget{{Type: entity.EscalationTargetPerson, Ref: "ghost"}}, Mode: entity.NotifyModeAll, Wait: 5 * time.Minute}},
		},
	}
	if err := h.srv.validatePolicy(context.Background(), "ws-1", policy); !entity.IsValidationError(err) {
		t.Fatalf("validatePolicy = %v, want a validation error for an unknown target", err)
	}
}

func TestReassignRewritesPersonTargets(t *testing.T) {
	h := newHarness(t)
	policy := entity.EscalationPolicy{
		ID: "pol-1", WorkspaceID: "ws-1", Slug: "p", Name: "p",
		Nodes: []entity.EscalationNode{
			{Level: &entity.EscalationLevel{ID: "l1", Targets: []entity.EscalationTarget{{Type: entity.EscalationTargetPerson, Ref: "u-old"}}, Mode: entity.NotifyModeAll, Wait: 5 * time.Minute}},
		},
	}
	h.policies.EXPECT().GetByID(gomock.Any(), "ws-1", "pol-1").Return(policy, nil)
	h.policies.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, p entity.EscalationPolicy) (entity.EscalationPolicy, error) {
			if p.Nodes[0].Level.Targets[0].Ref != "u-new" {
				t.Errorf("target ref = %s, want u-new", p.Nodes[0].Level.Targets[0].Ref)
			}
			return p, nil
		})

	if err := h.srv.Reassign(context.Background(), "ws-1", "pol-1:u-old", "u-new"); err != nil {
		t.Fatalf("reassign: %v", err)
	}
}
