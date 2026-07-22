package escalations

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	tx            repository.Transactor
	lock          repository.Lock
	workspaces    repository.Workspace
	members       repository.Member
	teams         repository.Team
	schedules     repository.Schedule
	policies      repository.EscalationPolicy
	runs          repository.EscalationRun
	alerts        repository.Alert
	routes        repository.AlertRoute
	policy        repository.Policy
	audit         repository.Audit
	notifier      service.Notifier
	notifications service.Notifications
	cfg           config.Auth
}

func New(
	tx repository.Transactor,
	lock repository.Lock,
	workspaces repository.Workspace,
	members repository.Member,
	teams repository.Team,
	schedules repository.Schedule,
	policies repository.EscalationPolicy,
	runs repository.EscalationRun,
	alerts repository.Alert,
	routes repository.AlertRoute,
	policy repository.Policy,
	audit repository.Audit,
	notifier service.Notifier,
	notifications service.Notifications,
	cfg config.Auth,
) service.Escalations {
	return &srv{
		tx: tx, lock: lock, workspaces: workspaces, members: members, teams: teams,
		schedules: schedules, policies: policies, runs: runs, alerts: alerts,
		routes: routes, policy: policy, audit: audit, notifier: notifier,
		notifications: notifications, cfg: cfg,
	}
}

func (s *srv) authorize(ctx context.Context, workspaceSlug string, act entity.PolicyAction, obj entity.PolicyObject) (entity.Identity, entity.Workspace, error) {
	id, ok := entity.IdentityFrom(ctx)
	if !ok {
		return entity.Identity{}, entity.Workspace{}, entity.ErrUnauthenticated
	}
	ws, err := s.workspaces.GetBySlug(ctx, workspaceSlug)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	if id.Kind == entity.IdentityKindAPIKey && id.WorkspaceID != ws.ID {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	if id.UserID != "" {
		active, err := s.members.IsActive(ctx, ws.ID, id.UserID)
		if err != nil {
			return entity.Identity{}, entity.Workspace{}, err
		}
		if !active {
			return entity.Identity{}, entity.Workspace{}, entity.ErrNotMember
		}
	}
	if !id.ScopePermits(obj, act) {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	allowed, err := s.policy.Allowed(ctx, id.Subject(), ws.ID, obj, act)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	if !allowed {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	return id, ws, nil
}

func (s *srv) event(actor entity.Identity, workspaceID, action, target string) entity.AuditEvent {
	return entity.AuditEvent{
		WorkspaceID: workspaceID,
		ActorType:   entity.AuditActorUser,
		ActorUserID: actor.UserID,
		ActorLabel:  actor.Label,
		Action:      action,
		Target:      target,
		IP:          actor.IP,
	}
}

func summarize(p entity.EscalationPolicy, routed int) entity.EscalationPolicySummary {
	steps := 0
	hasBranch := false
	var walk func(nodes []entity.EscalationNode)
	walk = func(nodes []entity.EscalationNode) {
		for _, node := range nodes {
			switch {
			case node.Level != nil:
				steps++
			case node.Branch != nil:
				hasBranch = true
				for _, lane := range node.Branch.Lanes {
					walk(lane.Nodes)
				}
			}
		}
	}
	walk(p.Nodes)
	return entity.EscalationPolicySummary{
		ID: p.ID, Slug: p.Slug, Name: p.Name, TeamSlug: p.TeamSlug,
		Routed: routed, StepCount: steps, HasBranch: hasBranch, Nodes: p.Nodes,
	}
}

func (s *srv) List(ctx context.Context, workspaceSlug string) ([]entity.EscalationPolicySummary, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectPolicies)
	if err != nil {
		return nil, err
	}
	policies, err := s.policies.List(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	routed, err := s.policies.RoutedCounts(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	out := make([]entity.EscalationPolicySummary, 0, len(policies))
	for _, p := range policies {
		out = append(out, summarize(p, routed[p.ID]))
	}
	return out, nil
}

func (s *srv) Get(ctx context.Context, workspaceSlug, policySlug string) (entity.EscalationPolicyDetail, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectPolicies)
	if err != nil {
		return entity.EscalationPolicyDetail{}, err
	}
	p, err := s.policies.GetBySlug(ctx, ws.ID, policySlug)
	if err != nil {
		return entity.EscalationPolicyDetail{}, err
	}
	routes, err := s.routes.List(ctx, ws.ID)
	if err != nil {
		return entity.EscalationPolicyDetail{}, err
	}
	linked := make([]entity.AlertRoute, 0, len(routes))
	for _, route := range routes {
		if route.PolicyID == p.ID {
			linked = append(linked, route)
		}
	}
	recent, err := s.runs.RecentByPolicy(ctx, p.ID, entity.EscalationRecentLimit)
	if err != nil {
		return entity.EscalationPolicyDetail{}, err
	}
	routed, err := s.policies.RoutedCounts(ctx, ws.ID)
	if err != nil {
		return entity.EscalationPolicyDetail{}, err
	}
	settings, err := s.routes.Settings(ctx, ws.ID)
	if err != nil {
		return entity.EscalationPolicyDetail{}, err
	}
	return entity.EscalationPolicyDetail{
		Policy: p, Routes: linked, Recent: recent,
		Routed: routed[p.ID], IsDefault: settings.DefaultPolicyID == p.ID,
	}, nil
}

type directoryIndex struct {
	members   map[string]entity.Member
	schedules map[string]entity.Schedule
	teams     map[string]entity.Team
	webhooks  map[string]entity.EscalationWebhook
}

func (s *srv) directoryIndex(ctx context.Context, workspaceID string) (directoryIndex, error) {
	idx := directoryIndex{
		members:   map[string]entity.Member{},
		schedules: map[string]entity.Schedule{},
		teams:     map[string]entity.Team{},
		webhooks:  map[string]entity.EscalationWebhook{},
	}
	members, err := s.members.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return idx, err
	}
	for _, m := range members {
		idx.members[m.UserID] = m
	}
	schedules, err := s.schedules.ListByWorkspace(ctx, workspaceID, false)
	if err != nil {
		return idx, err
	}
	for _, sched := range schedules {
		idx.schedules[sched.ID] = sched
	}
	teams, err := s.teams.ListByWorkspace(ctx, workspaceID, false)
	if err != nil {
		return idx, err
	}
	for _, team := range teams {
		idx.teams[team.ID] = team
	}
	webhooks, err := s.policies.ListWebhooks(ctx, workspaceID)
	if err != nil {
		return idx, err
	}
	for _, hook := range webhooks {
		idx.webhooks[hook.ID] = hook
	}
	return idx, nil
}

func (idx directoryIndex) known(t entity.EscalationTarget) bool {
	switch t.Type {
	case entity.EscalationTargetPerson:
		_, ok := idx.members[t.Ref]
		return ok
	case entity.EscalationTargetSchedule:
		_, ok := idx.schedules[t.Ref]
		return ok
	case entity.EscalationTargetTeam:
		_, ok := idx.teams[t.Ref]
		return ok
	case entity.EscalationTargetWebhook:
		_, ok := idx.webhooks[t.Ref]
		return ok
	default:
		return false
	}
}

func (idx directoryIndex) reachable(t entity.EscalationTarget) bool {
	switch t.Type {
	case entity.EscalationTargetPerson:
		m, ok := idx.members[t.Ref]
		return ok && m.Status == entity.MemberStatusActive
	default:
		return idx.known(t)
	}
}

func (s *srv) validatePolicy(ctx context.Context, workspaceID string, p entity.EscalationPolicy) error {
	if err := p.Validate(); err != nil {
		return err
	}
	idx, err := s.directoryIndex(ctx, workspaceID)
	if err != nil {
		return err
	}
	return entity.ValidateEscalationReach(p.Nodes, idx.known, idx.reachable)
}

func (s *srv) resolveTeam(ctx context.Context, workspaceID string, p *entity.EscalationPolicy) error {
	if p.TeamSlug == "" {
		p.TeamID = ""
		return nil
	}
	team, err := s.teams.GetBySlug(ctx, workspaceID, p.TeamSlug)
	if err != nil {
		return err
	}
	p.TeamID = team.ID
	return nil
}

func (s *srv) Create(ctx context.Context, workspaceSlug string, in entity.EscalationPolicy) (entity.EscalationPolicy, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectPolicies)
	if err != nil {
		return entity.EscalationPolicy{}, err
	}
	in.WorkspaceID = ws.ID
	in.Name = strings.TrimSpace(in.Name)
	if in.Slug == "" {
		in.Slug = entity.Slugify(in.Name)
	}
	if err := s.resolveTeam(ctx, ws.ID, &in); err != nil {
		return entity.EscalationPolicy{}, err
	}
	if err := s.validatePolicy(ctx, ws.ID, in); err != nil {
		return entity.EscalationPolicy{}, err
	}

	var created entity.EscalationPolicy
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		created, err = s.policies.Create(ctx, in)
		if err != nil {
			return fmt.Errorf("create escalation policy: %w", err)
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionPolicyCreated, created.Slug))
	})
	return created, err
}

func (s *srv) Update(ctx context.Context, workspaceSlug, policySlug string, in entity.EscalationPolicy) (entity.EscalationPolicy, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectPolicies)
	if err != nil {
		return entity.EscalationPolicy{}, err
	}
	existing, err := s.policies.GetBySlug(ctx, ws.ID, policySlug)
	if err != nil {
		return entity.EscalationPolicy{}, err
	}
	in.ID = existing.ID
	in.WorkspaceID = ws.ID
	in.Slug = existing.Slug
	in.Name = strings.TrimSpace(in.Name)
	if err := s.resolveTeam(ctx, ws.ID, &in); err != nil {
		return entity.EscalationPolicy{}, err
	}
	if err := s.validatePolicy(ctx, ws.ID, in); err != nil {
		return entity.EscalationPolicy{}, err
	}

	var updated entity.EscalationPolicy
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		updated, err = s.policies.Update(ctx, in)
		if err != nil {
			return fmt.Errorf("update escalation policy: %w", err)
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionPolicyUpdated, updated.Slug))
	})
	return updated, err
}

func (s *srv) Delete(ctx context.Context, workspaceSlug, policySlug string) error {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectPolicies)
	if err != nil {
		return err
	}
	p, err := s.policies.GetBySlug(ctx, ws.ID, policySlug)
	if err != nil {
		return err
	}
	refs, err := s.policies.Refs(ctx, ws.ID, p.ID)
	if err != nil {
		return err
	}
	if refs.Routes > 0 || refs.Monitors > 0 || refs.Default {
		return entity.ErrEscalationPolicyReferenced
	}
	if refs.ActiveRuns > 0 {
		return entity.ErrEscalationPolicyActive
	}
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.policies.Delete(ctx, ws.ID, p.ID); err != nil {
			return fmt.Errorf("delete escalation policy: %w", err)
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionPolicyDeleted, p.Slug))
	})
}

func (s *srv) Directory(ctx context.Context, workspaceSlug string) (entity.EscalationDirectory, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectPolicies)
	if err != nil {
		return entity.EscalationDirectory{}, err
	}
	members, err := s.members.ListByWorkspace(ctx, ws.ID)
	if err != nil {
		return entity.EscalationDirectory{}, err
	}
	schedules, err := s.schedules.ListByWorkspace(ctx, ws.ID, false)
	if err != nil {
		return entity.EscalationDirectory{}, err
	}
	teams, err := s.teams.ListByWorkspace(ctx, ws.ID, false)
	if err != nil {
		return entity.EscalationDirectory{}, err
	}
	webhooks, err := s.policies.ListWebhooks(ctx, ws.ID)
	if err != nil {
		return entity.EscalationDirectory{}, err
	}
	return entity.EscalationDirectory{Members: members, Schedules: schedules, Teams: teams, Webhooks: webhooks}, nil
}

func (s *srv) ListWebhooks(ctx context.Context, workspaceSlug string) ([]entity.EscalationWebhook, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectPolicies)
	if err != nil {
		return nil, err
	}
	return s.policies.ListWebhooks(ctx, ws.ID)
}

func (s *srv) CreateWebhook(ctx context.Context, workspaceSlug string, in entity.NewEscalationWebhook, secret string) (entity.EscalationWebhook, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectPolicies)
	if err != nil {
		return entity.EscalationWebhook{}, err
	}
	in.Name = strings.TrimSpace(in.Name)
	in.URL = strings.TrimSpace(in.URL)
	if in.Slug == "" {
		in.Slug = entity.Slugify(in.Name)
	}
	if err := in.Validate(); err != nil {
		return entity.EscalationWebhook{}, err
	}

	var created entity.EscalationWebhook
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		created, err = s.policies.CreateWebhook(ctx, ws.ID, in, secret)
		if err != nil {
			return fmt.Errorf("create escalation webhook: %w", err)
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionEscalationWebhookCreated, created.Slug))
	})
	return created, err
}

func (s *srv) UpdateWebhook(ctx context.Context, workspaceSlug, webhookSlug string, in entity.NewEscalationWebhook) (entity.EscalationWebhook, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectPolicies)
	if err != nil {
		return entity.EscalationWebhook{}, err
	}
	in.Slug = webhookSlug
	in.Name = strings.TrimSpace(in.Name)
	in.URL = strings.TrimSpace(in.URL)
	if err := in.Validate(); err != nil {
		return entity.EscalationWebhook{}, err
	}

	var updated entity.EscalationWebhook
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		updated, err = s.policies.UpdateWebhook(ctx, ws.ID, webhookSlug, in)
		if err != nil {
			return fmt.Errorf("update escalation webhook: %w", err)
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionEscalationWebhookUpdated, webhookSlug))
	})
	return updated, err
}

func (s *srv) DeleteWebhook(ctx context.Context, workspaceSlug, webhookSlug string) error {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectPolicies)
	if err != nil {
		return err
	}
	hook, err := s.policies.GetWebhook(ctx, ws.ID, webhookSlug)
	if err != nil {
		return err
	}
	referencing, err := s.policies.ListReferencingWebhook(ctx, ws.ID, hook.ID)
	if err != nil {
		return err
	}
	if len(referencing) > 0 {
		return entity.ErrEscalationWebhookInUse
	}
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.policies.DeleteWebhook(ctx, ws.ID, webhookSlug); err != nil {
			return fmt.Errorf("delete escalation webhook: %w", err)
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionEscalationWebhookDeleted, webhookSlug))
	})
}

func (s *srv) Start(ctx context.Context, alert entity.Alert, policyID string) error {
	if policyID == "" || alert.SuppressedBySilenceID != "" {
		return nil
	}
	policy, err := s.policies.GetByID(ctx, alert.WorkspaceID, policyID)
	if err != nil {
		if errors.Is(err, entity.ErrEscalationPolicyNotFound) {
			return nil
		}
		return err
	}
	run := entity.StartEscalationRun(alert, policy, time.Now().UTC())
	created, isNew, err := s.runs.Create(ctx, run)
	if err != nil {
		return err
	}
	if !isNew {
		return nil
	}
	return s.alerts.AppendEvent(ctx, alert.ID, entity.AlertEvent{
		At:     created.StartedAt,
		Kind:   entity.AlertEventEscalation,
		Text:   "Escalation started through " + policy.Slug,
		Result: strconv.Itoa(len(run.Path)) + " levels",
	})
}

func (s *srv) OnAcked(ctx context.Context, workspaceID string, alertIDs []string, now time.Time) error {
	for _, alertID := range alertIDs {
		run, err := s.runs.GetByAlertID(ctx, alertID)
		if err != nil {
			if errors.Is(err, entity.ErrEscalationRunNotFound) {
				continue
			}
			return err
		}
		if run.WorkspaceID != workspaceID {
			continue
		}
		if run.State != entity.EscalationRunning {
			continue
		}
		expiry := time.Time{}
		text := "Escalation stopped by the acknowledgement"
		if run.Snapshot.AckTimeout > 0 {
			expiry = now.Add(run.Snapshot.AckTimeout)
			text = "Escalation paused; resumes " + expiry.UTC().Format("15:04") + " UTC unless resolved"
		}
		marked, err := s.runs.MarkAcked(ctx, workspaceID, alertID, now, expiry)
		if err != nil {
			return err
		}
		if !marked {
			continue
		}
		if err := s.alerts.AppendEvent(ctx, alertID, entity.AlertEvent{
			At:   now,
			Kind: entity.AlertEventEscalation,
			Text: text,
		}); err != nil {
			return err
		}
	}
	return s.notifications.StopForAlerts(ctx, workspaceID, alertIDs, entity.NotifyStopAcked, now)
}

func (s *srv) OnResolved(ctx context.Context, workspaceID string, alertIDs []string, now time.Time) error {
	if _, err := s.runs.MarkResolved(ctx, workspaceID, alertIDs, now); err != nil {
		return err
	}
	return s.notifications.StopForAlerts(ctx, workspaceID, alertIDs, entity.NotifyStopResolved, now)
}

func (s *srv) RunForAlert(ctx context.Context, alertID string) (entity.EscalationRun, error) {
	return s.runs.GetByAlertID(ctx, alertID)
}

func (s *srv) Escalate(ctx context.Context, workspaceSlug, alertID string) error {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectAlerts)
	if err != nil {
		return err
	}
	run, err := s.runs.GetByAlertID(ctx, alertID)
	if err != nil {
		return err
	}
	if run.WorkspaceID != ws.ID {
		return entity.ErrEscalationRunNotFound
	}
	if !run.Active() {
		return entity.ErrEscalationRunFinished
	}
	now := time.Now().UTC()
	if run.State == entity.EscalationAcked {
		if err := s.resume(ctx, run, now); err != nil {
			return err
		}
		run, err = s.runs.GetByAlertID(ctx, alertID)
		if err != nil {
			return err
		}
	}
	if err := s.alerts.AppendEvent(ctx, run.AlertID, entity.AlertEvent{
		At:   now,
		Kind: entity.AlertEventEscalation,
		Text: fmt.Sprintf("Manually escalated by %s", actor.Label),
	}); err != nil {
		return err
	}
	if err := s.executeRun(ctx, run, now, true); err != nil {
		return err
	}
	return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionAlertEscalated, alertID))
}

func (s *srv) Advance(ctx context.Context, now time.Time) (int, error) {
	due, err := s.runs.ListDue(ctx, now, entity.EscalationSweepBatch)
	if err != nil {
		return 0, err
	}
	advanced := 0
	for _, run := range due {
		if err := s.executeRun(ctx, run, now, false); err != nil {
			logger.From(ctx).ErrorContext(ctx, "escalation advance failed", "run", run.ID, "alert", run.AlertID, "error", err)
			continue
		}
		advanced++
	}
	return advanced, nil
}

type pendingNotify struct {
	webhooks []entity.EscalationWebhook
	runIDs   []string
	alert    entity.Alert
	page     entity.AlertPage
}

func (s *srv) executeRun(ctx context.Context, stale entity.EscalationRun, now time.Time, force bool) error {
	var pending *pendingNotify
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		held, err := s.lock.TryJob(ctx, "escalation:"+stale.ID)
		if err != nil || !held {
			return err
		}
		run, err := s.runs.GetByAlertID(ctx, stale.AlertID)
		if err != nil {
			return err
		}
		alert, err := s.alerts.GetByID(ctx, run.WorkspaceID, run.AlertID)
		if err != nil {
			return err
		}
		if alert.Status == entity.AlertStatusResolved {
			_, err := s.runs.Finish(ctx, run.ID, entity.EscalationResolved, now)
			return err
		}
		if force && run.State == entity.EscalationRunning {
			run.NextAt = now
		}
		tick, due := run.NextTick(now)
		if !due {
			return nil
		}
		switch tick.Kind {
		case entity.EscalationTickResume:
			return s.resumeLocked(ctx, run, now)
		case entity.EscalationTickRepeat:
			repeated := run.Repeated(alert.Severity, now)
			saved, err := s.runs.SaveProgress(ctx, repeated)
			if err != nil || !saved {
				return err
			}
			return s.alerts.AppendEvent(ctx, run.AlertID, entity.AlertEvent{
				At:     now,
				Kind:   entity.AlertEventEscalation,
				Text:   "Nobody acknowledged. Repeating the policy, cycle " + strconv.Itoa(repeated.Cycle+1),
				Result: "cycle " + strconv.Itoa(repeated.Cycle),
			})
		case entity.EscalationTickExhaust:
			finished, err := s.runs.Finish(ctx, run.ID, entity.EscalationExhausted, now)
			if err != nil || !finished {
				return err
			}
			return s.alerts.AppendEvent(ctx, run.AlertID, entity.AlertEvent{
				At:   now,
				Kind: entity.AlertEventExhausted,
				Text: "Escalation exhausted. The alert stays open, but no one else will be paged.",
			})
		case entity.EscalationTickNotify:
			idx, err := s.directoryIndex(ctx, run.WorkspaceID)
			if err != nil {
				return err
			}
			valid := make([]entity.EscalationTarget, 0, len(tick.Level.Targets))
			for _, t := range tick.Level.Targets {
				if idx.reachable(t) {
					valid = append(valid, t)
				}
			}
			position := 0
			if tick.Level.Mode == entity.NotifyModeRoundRobin && len(valid) > 0 {
				position, err = s.runs.NextRoundRobin(ctx, run.PolicyID, tick.Level.ID)
				if err != nil {
					return err
				}
			}
			picked := tick.Level.PickTargets(valid, position)
			levelNum := run.StepIndex + 1

			advanced := run
			advanced.StepIndex++
			advanced.NextAt = now.Add(tick.Level.Wait)
			saved, err := s.runs.SaveProgress(ctx, advanced)
			if err != nil || !saved {
				return err
			}

			text := "Level " + strconv.Itoa(levelNum) + ": " + describeTargets(picked, idx)
			if len(picked) == 0 {
				text = "Level " + strconv.Itoa(levelNum) + ": no reachable targets, skipping"
			}
			if err := s.alerts.AppendEvent(ctx, run.AlertID, entity.AlertEvent{
				At:     now,
				Kind:   entity.AlertEventEscalation,
				Text:   text,
				Result: string(tick.Level.Mode),
			}); err != nil {
				return err
			}
			if len(picked) == 0 {
				return nil
			}
			workspace, err := s.workspaces.GetByID(ctx, run.WorkspaceID)
			if err != nil {
				return err
			}
			collected := &pendingNotify{
				alert: alert,
				page:  entity.BuildAlertPage(alert, workspace.Slug, run.PolicySlug, s.cfg.BaseURL, levelNum),
			}
			for _, target := range picked {
				for _, d := range s.resolveDeliveries(ctx, run, target, idx, now) {
					if d.webhook.ID != "" {
						collected.webhooks = append(collected.webhooks, d.webhook)
						continue
					}
					if d.gapResult != "" {
						if err := s.alerts.AppendEvent(ctx, run.AlertID, entity.AlertEvent{
							At: now, Kind: entity.AlertEventNotified, Text: d.label + " failed", Result: d.gapResult,
						}); err != nil {
							return err
						}
						continue
					}
					started, err := s.notifications.Page(ctx, entity.NotifyRequest{
						WorkspaceID: run.WorkspaceID, AlertID: run.AlertID, UserID: d.userID, Email: d.email,
						Label: d.name, PolicySlug: run.PolicySlug, EscalationID: run.ID,
						EscalationCycle: run.Cycle, Level: levelNum, Severity: alert.Severity, At: now,
					})
					if err != nil {
						return err
					}
					if len(started.Plan.Steps) == 0 {
						if err := s.alerts.AppendEvent(ctx, run.AlertID, entity.AlertEvent{
							At: now, Kind: entity.AlertEventNotified, Text: d.name + " has no reachable channel", Result: "no channel",
						}); err != nil {
							return err
						}
						continue
					}
					if err := s.alerts.AppendEvent(ctx, run.AlertID, entity.AlertEvent{
						At: now, Kind: entity.AlertEventNotified, Text: d.label, Result: entity.NotifyChannelSummary(started.Plan),
					}); err != nil {
						return err
					}
					collected.runIDs = append(collected.runIDs, started.ID)
				}
			}
			if len(collected.webhooks) > 0 || len(collected.runIDs) > 0 {
				pending = collected
			}
			return nil
		}
		return nil
	})
	if err != nil || pending == nil {
		return err
	}
	s.deliver(ctx, stale, *pending, now)
	return nil
}

func describeTargets(targets []entity.EscalationTarget, idx directoryIndex) string {
	names := make([]string, 0, len(targets))
	for _, t := range targets {
		switch t.Type {
		case entity.EscalationTargetPerson:
			names = append(names, idx.members[t.Ref].Name)
		case entity.EscalationTargetSchedule:
			names = append(names, "on-call for "+idx.schedules[t.Ref].Slug)
		case entity.EscalationTargetTeam:
			names = append(names, "team "+idx.teams[t.Ref].Slug)
		case entity.EscalationTargetWebhook:
			names = append(names, "webhook "+idx.webhooks[t.Ref].Slug)
		}
	}
	return "paging " + strings.Join(names, ", ")
}

func (s *srv) deliver(ctx context.Context, run entity.EscalationRun, p pendingNotify, now time.Time) {
	for _, hook := range p.webhooks {
		result := s.notifier.CallWebhook(ctx, hook, p.alert, p.page)
		text := "Called webhook " + hook.Slug
		if !result.Delivered {
			text = text + " failed"
		}
		if err := s.alerts.AppendEvent(ctx, run.AlertID, entity.AlertEvent{
			At:     time.Now().UTC(),
			Kind:   entity.AlertEventNotified,
			Text:   text,
			Result: result.Detail,
		}); err != nil {
			logger.From(ctx).ErrorContext(ctx, "record webhook delivery failed", "alert", run.AlertID, "error", err)
		}
	}
	if len(p.runIDs) > 0 {
		if err := s.notifications.RunNow(ctx, p.runIDs, now); err != nil {
			logger.From(ctx).ErrorContext(ctx, "run notification ladders failed", "alert", run.AlertID, "error", err)
		}
	}
}

type delivery struct {
	label     string
	name      string
	userID    string
	email     string
	webhook   entity.EscalationWebhook
	gapResult string
}

func (s *srv) resolveDeliveries(ctx context.Context, run entity.EscalationRun, target entity.EscalationTarget, idx directoryIndex, now time.Time) []delivery {
	switch target.Type {
	case entity.EscalationTargetPerson:
		member, ok := idx.members[target.Ref]
		if !ok || member.Status != entity.MemberStatusActive {
			return nil
		}
		return []delivery{{label: "Paging " + member.Name, name: member.Name, userID: member.UserID, email: member.Email}}
	case entity.EscalationTargetSchedule:
		sched, ok := idx.schedules[target.Ref]
		if !ok {
			return nil
		}
		cover := sched.OnCallAt(now)
		if cover.UserID == "" {
			return []delivery{{label: "No one is on call for " + sched.Slug, gapResult: "schedule gap"}}
		}
		member, ok := idx.members[cover.UserID]
		if !ok || member.Status != entity.MemberStatusActive {
			return []delivery{{label: "On-call for " + sched.Slug + " can't be paged", gapResult: "deactivated member"}}
		}
		return []delivery{{
			label: "Paging " + member.Name + " (on call for " + sched.Slug + ")",
			name:  member.Name, userID: member.UserID, email: member.Email,
		}}
	case entity.EscalationTargetTeam:
		team, ok := idx.teams[target.Ref]
		if !ok {
			return nil
		}
		out := make([]delivery, 0, len(team.MemberIDs))
		for _, userID := range team.MemberIDs {
			member, ok := idx.members[userID]
			if !ok || member.Status != entity.MemberStatusActive {
				continue
			}
			out = append(out, delivery{
				label: "Paging " + member.Name + " (team " + team.Slug + ")",
				name:  member.Name, userID: member.UserID, email: member.Email,
			})
		}
		return out
	case entity.EscalationTargetWebhook:
		hook, ok := idx.webhooks[target.Ref]
		if !ok {
			return nil
		}
		return []delivery{{label: "Called webhook " + hook.Slug, webhook: hook}}
	default:
		return nil
	}
}

func (s *srv) resume(ctx context.Context, run entity.EscalationRun, now time.Time) error {
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		held, err := s.lock.TryJob(ctx, "escalation:"+run.ID)
		if err != nil || !held {
			return err
		}
		return s.resumeLocked(ctx, run, now)
	})
}

func (s *srv) resumeLocked(ctx context.Context, run entity.EscalationRun, now time.Time) error {
	resumed, err := s.runs.Resume(ctx, run.ID, now)
	if err != nil || !resumed {
		return err
	}
	if _, err := s.alerts.Reopen(ctx, run.AlertID, now); err != nil {
		return err
	}
	return s.alerts.AppendEvent(ctx, run.AlertID, entity.AlertEvent{
		At:   now,
		Kind: entity.AlertEventTimeout,
		Text: "Acknowledgement expired. The alert is open again and escalation resumes.",
	})
}

func (s *srv) ListByUser(ctx context.Context, workspaceID, userID string) ([]entity.MemberReference, error) {
	policies, err := s.policies.ListReferencingUser(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	out := make([]entity.MemberReference, 0, len(policies))
	for _, p := range policies {
		out = append(out, entity.MemberReference{
			ID:     p.ID + ":" + userID,
			Kind:   entity.ReferenceKindPolicy,
			Icon:   "arrow-up-right",
			Label:  p.Name,
			Detail: "Escalation policy target",
		})
	}
	return out, nil
}

func (s *srv) Reassign(ctx context.Context, workspaceID, referenceID, toUserID string) error {
	policyID, fromUserID, ok := strings.Cut(referenceID, ":")
	if !ok {
		return entity.ErrReferenceUnknown
	}
	p, err := s.policies.GetByID(ctx, workspaceID, policyID)
	if err != nil {
		return err
	}
	replaceTargets(p.Nodes, fromUserID, toUserID)
	if _, err := s.policies.Update(ctx, p); err != nil {
		return fmt.Errorf("reassign policy target: %w", err)
	}
	return nil
}

func replaceTargets(nodes []entity.EscalationNode, fromUserID, toUserID string) {
	for _, node := range nodes {
		switch {
		case node.Level != nil:
			for i, t := range node.Level.Targets {
				if t.Type == entity.EscalationTargetPerson && t.Ref == fromUserID {
					node.Level.Targets[i].Ref = toUserID
				}
			}
		case node.Branch != nil:
			for _, lane := range node.Branch.Lanes {
				replaceTargets(lane.Nodes, fromUserID, toUserID)
			}
		}
	}
}
