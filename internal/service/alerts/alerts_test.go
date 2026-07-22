package alerts

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository/alert"
	"github.com/opsybot/opsybot/internal/repository/audit"
	"github.com/opsybot/opsybot/internal/repository/member"
	"github.com/opsybot/opsybot/internal/repository/policy"
	"github.com/opsybot/opsybot/internal/repository/workspace"
	"github.com/opsybot/opsybot/internal/service/escalations"
)

type fakeTx struct{}

func (fakeTx) WithTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type harness struct {
	srv    *srv
	alerts *alert.MockAlert
	esc    *escalations.MockEscalations
	audits *audit.MockAudit
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	ctrl := gomock.NewController(t)
	ws := workspace.NewMockWorkspace(ctrl)
	members := member.NewMockMember(ctrl)
	pol := policy.NewMockPolicy(ctrl)
	h := &harness{
		alerts: alert.NewMockAlert(ctrl),
		esc:    escalations.NewMockEscalations(ctrl),
		audits: audit.NewMockAudit(ctrl),
	}
	h.srv = &srv{
		tx: fakeTx{}, workspaces: ws, members: members, alerts: h.alerts,
		policy: pol, audit: h.audits, escalations: h.esc,
	}
	ws.EXPECT().GetBySlug(gomock.Any(), "acme").Return(entity.Workspace{ID: "ws-1", Slug: "acme"}, nil).AnyTimes()
	members.EXPECT().IsActive(gomock.Any(), "ws-1", "u1").Return(true, nil).AnyTimes()
	pol.EXPECT().Allowed(gomock.Any(), gomock.Any(), "ws-1", gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
	return h
}

func actorCtx() context.Context {
	return entity.WithIdentity(context.Background(), entity.Identity{
		Kind: entity.IdentityKindSession, UserID: "u1", Label: "Priya",
	})
}

func TestAcknowledgeOnlyTouchesRowsTheRepoConfirmed(t *testing.T) {
	h := newHarness(t)
	ids := []string{"al-mine", "al-foreign"}

	h.alerts.EXPECT().
		Acknowledge(gomock.Any(), "ws-1", ids, "u1", "Priya", gomock.Any()).
		Return([]string{"al-mine"}, nil)
	h.alerts.EXPECT().
		AppendEvent(gomock.Any(), "al-mine", gomock.Cond(func(e entity.AlertEvent) bool {
			return e.Kind == entity.AlertEventAcked && e.Text == "Acknowledged by Priya"
		})).
		Return(nil)
	h.esc.EXPECT().
		OnAcked(gomock.Any(), "ws-1", []string{"al-mine"}, gomock.Any()).
		Return(nil)
	h.audits.EXPECT().
		Create(gomock.Any(), gomock.Cond(func(e entity.AuditEvent) bool {
			return e.Target == "1 alerts"
		})).
		Return(nil)

	updated, err := h.srv.Acknowledge(actorCtx(), "acme", ids)
	if err != nil {
		t.Fatalf("acknowledge: %v", err)
	}
	if updated != 1 {
		t.Fatalf("updated = %d, want 1", updated)
	}
}

func TestResolveOnlyTouchesRowsTheRepoConfirmed(t *testing.T) {
	h := newHarness(t)
	ids := []string{"al-mine", "al-foreign"}

	h.alerts.EXPECT().
		Resolve(gomock.Any(), "ws-1", ids, gomock.Any(), entity.ResolveModeManual).
		Return([]string{"al-mine"}, nil)
	h.alerts.EXPECT().
		AppendEvent(gomock.Any(), "al-mine", gomock.Cond(func(e entity.AlertEvent) bool {
			return e.Kind == entity.AlertEventResolved && e.Text == "Resolved by Priya"
		})).
		Return(nil)
	h.esc.EXPECT().
		OnResolved(gomock.Any(), "ws-1", []string{"al-mine"}, gomock.Any()).
		Return(nil)
	h.audits.EXPECT().
		Create(gomock.Any(), gomock.Cond(func(e entity.AuditEvent) bool {
			return e.Target == "1 alerts"
		})).
		Return(nil)

	updated, err := h.srv.Resolve(actorCtx(), "acme", ids)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if updated != 1 {
		t.Fatalf("updated = %d, want 1", updated)
	}
}
