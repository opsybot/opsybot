package incidents

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository/audit"
	"github.com/opsybot/opsybot/internal/repository/alert"
	"github.com/opsybot/opsybot/internal/repository/incident"
	"github.com/opsybot/opsybot/internal/repository/incident_field_def"
	"github.com/opsybot/opsybot/internal/repository/incident_severity"
	"github.com/opsybot/opsybot/internal/repository/lock"
	"github.com/opsybot/opsybot/internal/repository/member"
	"github.com/opsybot/opsybot/internal/repository/policy"
	servicerepo "github.com/opsybot/opsybot/internal/repository/service"
	"github.com/opsybot/opsybot/internal/repository/team"
	"github.com/opsybot/opsybot/internal/repository/workspace"
	"github.com/opsybot/opsybot/internal/service/escalations"
)

type fakeTx struct{}

func (fakeTx) WithTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type harness struct {
	srv        *srv
	lock       *lock.MockLock
	ws         *workspace.MockWorkspace
	members    *member.MockMember
	teams      *team.MockTeam
	incidents  *incident.MockIncident
	services   *servicerepo.MockService
	severities *incident_severity.MockIncidentSeverity
	fieldDefs  *incident_field_def.MockIncidentFieldDef
	alerts     *alert.MockAlert
	pol        *policy.MockPolicy
	audit      *audit.MockAudit
	esc        *escalations.MockEscalations
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	ctrl := gomock.NewController(t)
	h := &harness{
		lock:       lock.NewMockLock(ctrl),
		ws:         workspace.NewMockWorkspace(ctrl),
		members:    member.NewMockMember(ctrl),
		teams:      team.NewMockTeam(ctrl),
		incidents:  incident.NewMockIncident(ctrl),
		services:   servicerepo.NewMockService(ctrl),
		severities: incident_severity.NewMockIncidentSeverity(ctrl),
		fieldDefs:  incident_field_def.NewMockIncidentFieldDef(ctrl),
		alerts:     alert.NewMockAlert(ctrl),
		pol:        policy.NewMockPolicy(ctrl),
		audit:      audit.NewMockAudit(ctrl),
		esc:        escalations.NewMockEscalations(ctrl),
	}
	h.srv = &srv{
		tx: fakeTx{}, lock: h.lock, workspaces: h.ws, members: h.members, teams: h.teams,
		incidents: h.incidents, services: h.services, severities: h.severities, fieldDefs: h.fieldDefs,
		alerts: h.alerts, policy: h.pol, audit: h.audit, escalations: h.esc,
	}
	return h
}

func (h *harness) authorizeOK() {
	h.ws.EXPECT().GetBySlug(gomock.Any(), "acme").Return(entity.Workspace{ID: "ws-1", Slug: "acme"}, nil).AnyTimes()
	h.members.EXPECT().IsActive(gomock.Any(), "ws-1", "u1").Return(true, nil).AnyTimes()
	h.pol.EXPECT().Allowed(gomock.Any(), gomock.Any(), "ws-1", gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
}

func (h *harness) emptyDirectory() {
	h.members.EXPECT().ListByWorkspace(gomock.Any(), "ws-1").Return(nil, nil).AnyTimes()
	h.teams.EXPECT().ListByWorkspace(gomock.Any(), "ws-1", true).Return(nil, nil).AnyTimes()
}

func adminCtx() context.Context {
	return entity.WithIdentity(context.Background(), entity.Identity{Kind: entity.IdentityKindSession, UserID: "u1", Label: "Priya"})
}

func TestDeclareFromAlertPreservesContext(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()
	h.emptyDirectory()

	h.alerts.EXPECT().GetByID(gomock.Any(), "ws-1", "al-1").Return(entity.Alert{
		ID: "al-1", Title: "Checkout latency spike", Severity: entity.SeverityCritical, ServiceName: "checkout",
	}, nil)
	h.severities.EXPECT().Exists(gomock.Any(), "ws-1", "SEV1").Return(true, nil)
	h.lock.EXPECT().Workspace(gomock.Any(), "ws-1").Return(nil)
	h.incidents.EXPECT().NextNumber(gomock.Any(), "ws-1").Return(7, nil)

	var created entity.Incident
	h.incidents.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, in entity.Incident) (entity.Incident, error) {
			created = in
			created.ID = "inc-1"
			return created, nil
		})
	h.incidents.EXPECT().LinkAlert(gomock.Any(), "ws-1", "inc-1", "al-1").Return(nil)
	h.incidents.EXPECT().AppendEvent(gomock.Any(), gomock.Any()).Return(nil).Times(2)
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	h.incidents.EXPECT().GetByID(gomock.Any(), "ws-1", "inc-1").DoAndReturn(
		func(_ context.Context, _, id string) (entity.Incident, error) {
			return entity.Incident{ID: id, Number: 7, Name: created.Name, SeverityLevel: created.SeverityLevel, Status: entity.IncidentStatusDeclared}, nil
		})

	out, err := h.srv.Declare(adminCtx(), "acme", entity.IncidentDeclare{FromAlertID: "al-1"})
	if err != nil {
		t.Fatalf("declare from alert: %v", err)
	}
	if created.Name != "Checkout latency spike" {
		t.Errorf("name not carried from alert: %q", created.Name)
	}
	if created.SeverityLevel != "SEV1" {
		t.Errorf("severity not mapped from critical alert: %q", created.SeverityLevel)
	}
	if created.DeclaredBy != "u1" {
		t.Errorf("declaredBy = %q, want u1", created.DeclaredBy)
	}
	if out.Number != 7 {
		t.Errorf("number = %d, want 7", out.Number)
	}
}

func TestResolveResolvesLinkedAlerts(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()
	h.emptyDirectory()

	h.incidents.EXPECT().GetByID(gomock.Any(), "ws-1", "inc-1").Return(entity.Incident{
		ID: "inc-1", Number: 7, Status: entity.IncidentStatusMonitoring,
	}, nil).AnyTimes()
	h.incidents.EXPECT().SetStatus(gomock.Any(), "ws-1", "inc-1", entity.IncidentStatusMonitoring, entity.IncidentStatusResolved, gomock.Any(), "Root cause patched").Return(true, nil)
	h.incidents.EXPECT().LinkedAlertIDs(gomock.Any(), "inc-1").Return([]string{"al-1", "al-2"}, nil)
	h.alerts.EXPECT().Resolve(gomock.Any(), "ws-1", []string{"al-1", "al-2"}, gomock.Any(), entity.ResolveModeIncident).Return([]string{"al-1", "al-2"}, nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
	h.esc.EXPECT().OnResolved(gomock.Any(), "ws-1", []string{"al-1", "al-2"}, gomock.Any()).Return(nil)
	h.incidents.EXPECT().AppendEvent(gomock.Any(), gomock.Any()).Return(nil)
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	if _, err := h.srv.Resolve(adminCtx(), "acme", "inc-1", "Root cause patched"); err != nil {
		t.Fatalf("resolve: %v", err)
	}
}

func TestResolveRequiresSummary(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	if _, err := h.srv.Resolve(adminCtx(), "acme", "inc-1", "   "); !errors.Is(err, entity.ErrIncidentResolutionMissing) {
		t.Fatalf("expected resolution missing error, got %v", err)
	}
}

func TestChangeStatusRejectsIllegalJump(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()
	h.incidents.EXPECT().GetByID(gomock.Any(), "ws-1", "inc-1").Return(entity.Incident{
		ID: "inc-1", Status: entity.IncidentStatusDeclared,
	}, nil)

	_, err := h.srv.ChangeStatus(adminCtx(), "acme", "inc-1", entity.IncidentStatusMonitoring)
	if !errors.Is(err, entity.ErrIncidentInvalidTransition) {
		t.Fatalf("expected invalid transition, got %v", err)
	}
}

func TestChangeStatusToResolvedNeedsResolveEndpoint(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	_, err := h.srv.ChangeStatus(adminCtx(), "acme", "inc-1", entity.IncidentStatusResolved)
	if !errors.Is(err, entity.ErrIncidentResolutionMissing) {
		t.Fatalf("expected resolution missing error, got %v", err)
	}
}

func TestReopenRejectedWhenNotResolved(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()
	h.incidents.EXPECT().Reopen(gomock.Any(), "ws-1", "inc-1", entity.IncidentStatusInvestigating, gomock.Any()).Return(false, nil)

	_, err := h.srv.Reopen(adminCtx(), "acme", "inc-1")
	if !errors.Is(err, entity.ErrIncidentInvalidTransition) {
		t.Fatalf("expected invalid transition on reopen, got %v", err)
	}
}
