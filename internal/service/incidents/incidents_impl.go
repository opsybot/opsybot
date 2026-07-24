package incidents

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	tx          repository.Transactor
	lock        repository.Lock
	workspaces  repository.Workspace
	members     repository.Member
	teams       repository.Team
	incidents   repository.Incident
	services    repository.Service
	severities  repository.IncidentSeverity
	fieldDefs   repository.IncidentFieldDef
	alerts      repository.Alert
	blobs       repository.Blob
	policy      repository.Policy
	audit       repository.Audit
	escalations service.Escalations
}

func New(
	tx repository.Transactor,
	lock repository.Lock,
	workspaces repository.Workspace,
	members repository.Member,
	teams repository.Team,
	incidents repository.Incident,
	services repository.Service,
	severities repository.IncidentSeverity,
	fieldDefs repository.IncidentFieldDef,
	alerts repository.Alert,
	blobs repository.Blob,
	policy repository.Policy,
	audit repository.Audit,
	escalations service.Escalations,
) service.Incidents {
	return &srv{
		tx: tx, lock: lock, workspaces: workspaces, members: members, teams: teams,
		incidents: incidents, services: services, severities: severities, fieldDefs: fieldDefs,
		alerts: alerts, blobs: blobs, policy: policy, audit: audit, escalations: escalations,
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

func (s *srv) appendTimeline(ctx context.Context, ws entity.Workspace, incidentID string, kind entity.IncidentEventKind, text string, actor entity.Identity, at time.Time) error {
	_, err := s.incidents.AppendEvent(ctx, entity.IncidentEvent{
		IncidentID:  incidentID,
		WorkspaceID: ws.ID,
		At:          at,
		Kind:        kind,
		Category:    entity.CategoryForKind(kind),
		Source:      entity.IncidentSourceSystem,
		Text:        text,
		Actor:       actor.Label,
		ActorUserID: actor.UserID,
	})
	return err
}

type detailChange struct {
	kind entity.IncidentEventKind
	text string
}

func detailChanges(before entity.Incident, in entity.IncidentUpdate, teamID string, serviceIDs []string, names map[string]string) []detailChange {
	var out []detailChange
	if before.Name != in.Name {
		out = append(out, detailChange{entity.IncidentEventRenamed, fmt.Sprintf("Renamed to %q", in.Name)})
	}
	if before.Summary != in.Summary {
		out = append(out, detailChange{entity.IncidentEventSummaryChanged, "Summary updated"})
	}
	if before.LeadUserID != in.LeadUserID {
		text := "Incident lead cleared"
		if in.LeadUserID != "" {
			text = fmt.Sprintf("Incident lead set to %s", names[in.LeadUserID])
		}
		out = append(out, detailChange{entity.IncidentEventLeadChanged, text})
	}
	if before.TeamID != teamID {
		text := "Owning team cleared"
		if in.TeamSlug != "" {
			text = fmt.Sprintf("Owning team set to %s", in.TeamSlug)
		}
		out = append(out, detailChange{entity.IncidentEventUpdated, text})
	}
	if !sameServices(before.Services, serviceIDs) {
		out = append(out, detailChange{entity.IncidentEventUpdated, "Affected services updated"})
	}
	return out
}

func sameServices(before []entity.Service, serviceIDs []string) bool {
	if len(before) != len(serviceIDs) {
		return false
	}
	existing := make(map[string]bool, len(before))
	for _, svc := range before {
		existing[svc.ID] = true
	}
	for _, id := range serviceIDs {
		if !existing[id] {
			return false
		}
	}
	return true
}

func (s *srv) memberNames(ctx context.Context, workspaceID string) (map[string]string, error) {
	members, err := s.members.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(members))
	for _, m := range members {
		out[m.UserID] = m.Name
	}
	return out, nil
}

func (s *srv) teamSlugsByID(ctx context.Context, workspaceID string) (map[string]string, error) {
	teams, err := s.teams.ListByWorkspace(ctx, workspaceID, true)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(teams))
	for _, t := range teams {
		out[t.ID] = t.Slug
	}
	return out, nil
}

func (s *srv) enrich(inc *entity.Incident, names, teamSlugs map[string]string) {
	if inc.LeadUserID != "" {
		inc.LeadLabel = names[inc.LeadUserID]
	}
	if inc.TeamID != "" {
		inc.TeamSlug = teamSlugs[inc.TeamID]
	}
	for i := range inc.Services {
		inc.Services[i].TeamSlug = teamSlugs[inc.Services[i].TeamID]
	}
	for i := range inc.Followups {
		if inc.Followups[i].OwnerUserID != "" {
			inc.Followups[i].OwnerLabel = names[inc.Followups[i].OwnerUserID]
		}
	}
}

func (s *srv) List(ctx context.Context, workspaceSlug string, filter entity.IncidentFilter) (entity.IncidentPage, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.IncidentPage{}, err
	}
	page, err := s.incidents.List(ctx, ws.ID, filter)
	if err != nil {
		return entity.IncidentPage{}, err
	}
	names, err := s.memberNames(ctx, ws.ID)
	if err != nil {
		return entity.IncidentPage{}, err
	}
	teamSlugs, err := s.teamSlugsByID(ctx, ws.ID)
	if err != nil {
		return entity.IncidentPage{}, err
	}
	for i := range page.Incidents {
		s.enrich(&page.Incidents[i], names, teamSlugs)
	}
	return page, nil
}

func (s *srv) Get(ctx context.Context, workspaceSlug, id string) (entity.Incident, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.Incident{}, err
	}
	return s.hydrate(ctx, ws.ID, id)
}

func (s *srv) hydrate(ctx context.Context, workspaceID, id string) (entity.Incident, error) {
	inc, err := s.incidents.GetByID(ctx, workspaceID, id)
	if err != nil {
		return entity.Incident{}, err
	}
	names, err := s.memberNames(ctx, workspaceID)
	if err != nil {
		return entity.Incident{}, err
	}
	teamSlugs, err := s.teamSlugsByID(ctx, workspaceID)
	if err != nil {
		return entity.Incident{}, err
	}
	s.enrich(&inc, names, teamSlugs)
	page, err := s.timeline(ctx, inc, entity.TimelineFilter{})
	if err != nil {
		return entity.Incident{}, err
	}
	inc.Timeline = page.Entries
	return inc, nil
}

func (s *srv) resolveTeam(ctx context.Context, workspaceID, teamSlug string) (string, error) {
	teamSlug = strings.TrimSpace(teamSlug)
	if teamSlug == "" {
		return "", nil
	}
	team, err := s.teams.GetBySlug(ctx, workspaceID, teamSlug)
	if err != nil {
		return "", err
	}
	return team.ID, nil
}

func (s *srv) validateLead(ctx context.Context, workspaceID, userID string) error {
	if userID == "" {
		return nil
	}
	active, err := s.members.IsActive(ctx, workspaceID, userID)
	if err != nil {
		return err
	}
	if !active {
		return entity.ErrIncidentLeadUnknown
	}
	return nil
}

func (s *srv) validateServices(ctx context.Context, workspaceID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	found, err := s.services.ExistingIDs(ctx, workspaceID, ids)
	if err != nil {
		return err
	}
	if len(found) != len(dedupe(ids)) {
		return entity.ErrServiceNotFound
	}
	return nil
}

func dedupe(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func (s *srv) Declare(ctx context.Context, workspaceSlug string, in entity.IncidentDeclare) (entity.Incident, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.Incident{}, err
	}
	sev := strings.TrimSpace(in.SeverityLevel)
	linkAlertID := ""
	linkAlertTitle := ""
	if strings.TrimSpace(in.FromAlertID) != "" {
		alert, err := s.alerts.GetByID(ctx, ws.ID, in.FromAlertID)
		if err != nil {
			return entity.Incident{}, err
		}
		linkAlertID = alert.ID
		linkAlertTitle = alert.Title
		if strings.TrimSpace(in.Name) == "" {
			in.Name = alert.Title
		}
		if sev == "" {
			sev = entity.IncidentSeverityForAlert(alert.Severity)
		}
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Summary = strings.TrimSpace(in.Summary)
	if err := in.Validate(); err != nil {
		return entity.Incident{}, err
	}
	if sev == "" {
		return entity.Incident{}, entity.ErrIncidentSeverityUnknown
	}
	exists, err := s.severities.Exists(ctx, ws.ID, sev)
	if err != nil {
		return entity.Incident{}, err
	}
	if !exists {
		return entity.Incident{}, entity.ErrIncidentSeverityUnknown
	}
	teamID, err := s.resolveTeam(ctx, ws.ID, in.TeamSlug)
	if err != nil {
		return entity.Incident{}, err
	}
	if err := s.validateLead(ctx, ws.ID, in.LeadUserID); err != nil {
		return entity.Incident{}, err
	}
	serviceIDs := dedupe(in.ServiceIDs)
	if err := s.validateServices(ctx, ws.ID, serviceIDs); err != nil {
		return entity.Incident{}, err
	}

	now := time.Now().UTC()
	var created entity.Incident
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Workspace(ctx, ws.ID); err != nil {
			return err
		}
		number, err := s.incidents.NextNumber(ctx, ws.ID)
		if err != nil {
			return err
		}
		created, err = s.incidents.Create(ctx, entity.Incident{
			WorkspaceID:   ws.ID,
			Number:        number,
			Name:          in.Name,
			Summary:       in.Summary,
			SeverityLevel: sev,
			LeadUserID:    in.LeadUserID,
			TeamID:        teamID,
			DeclaredBy:    actor.UserID,
			DeclaredAt:    now,
		})
		if err != nil {
			return err
		}
		if len(serviceIDs) > 0 {
			if err := s.incidents.ReplaceServices(ctx, ws.ID, created.ID, serviceIDs); err != nil {
				return err
			}
		}
		if err := s.appendTimeline(ctx, ws, created.ID, entity.IncidentEventDeclared, fmt.Sprintf("Declared at %s", sev), actor, now); err != nil {
			return err
		}
		if linkAlertID != "" {
			if err := s.incidents.LinkAlert(ctx, ws.ID, created.ID, linkAlertID); err != nil {
				return err
			}
			if err := s.appendTimeline(ctx, ws, created.ID, entity.IncidentEventAlertLinked, fmt.Sprintf("Linked alert %q", linkAlertTitle), actor, now); err != nil {
				return err
			}
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentDeclared, fmt.Sprintf("INC-%d", number)))
	})
	if err != nil {
		return entity.Incident{}, err
	}
	return s.hydrate(ctx, ws.ID, created.ID)
}

func (s *srv) Update(ctx context.Context, workspaceSlug, id string, in entity.IncidentUpdate) (entity.Incident, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.Incident{}, err
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Summary = strings.TrimSpace(in.Summary)
	if err := in.Validate(); err != nil {
		return entity.Incident{}, err
	}
	teamID, err := s.resolveTeam(ctx, ws.ID, in.TeamSlug)
	if err != nil {
		return entity.Incident{}, err
	}
	if err := s.validateLead(ctx, ws.ID, in.LeadUserID); err != nil {
		return entity.Incident{}, err
	}
	serviceIDs := dedupe(in.ServiceIDs)
	if err := s.validateServices(ctx, ws.ID, serviceIDs); err != nil {
		return entity.Incident{}, err
	}
	names, err := s.memberNames(ctx, ws.ID)
	if err != nil {
		return entity.Incident{}, err
	}
	now := time.Now().UTC()
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		before, err := s.incidents.GetByID(ctx, ws.ID, id)
		if err != nil {
			return err
		}
		lead := in.LeadUserID
		team := teamID
		ok, err := s.incidents.Patch(ctx, ws.ID, id, entity.IncidentPatch{
			Name:       &in.Name,
			Summary:    &in.Summary,
			LeadUserID: &lead,
			TeamID:     &team,
		})
		if err != nil {
			return err
		}
		if !ok {
			return entity.ErrIncidentNotFound
		}
		if err := s.incidents.ReplaceServices(ctx, ws.ID, id, serviceIDs); err != nil {
			return err
		}
		for _, change := range detailChanges(before, in, teamID, serviceIDs, names) {
			if err := s.appendTimeline(ctx, ws, id, change.kind, change.text, actor, now); err != nil {
				return err
			}
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentUpdated, id))
	})
	if err != nil {
		return entity.Incident{}, err
	}
	return s.hydrate(ctx, ws.ID, id)
}

func (s *srv) ChangeStatus(ctx context.Context, workspaceSlug, id string, to entity.IncidentStatus) (entity.Incident, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.Incident{}, err
	}
	if !to.Valid() {
		return entity.Incident{}, entity.ErrIncidentInvalidTransition
	}
	if to == entity.IncidentStatusResolved {
		return entity.Incident{}, entity.ErrIncidentResolutionMissing
	}
	current, err := s.incidents.GetByID(ctx, ws.ID, id)
	if err != nil {
		return entity.Incident{}, err
	}
	if !current.Status.CanTransition(to) {
		return entity.Incident{}, entity.ErrIncidentInvalidTransition
	}
	now := time.Now().UTC()
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		ok, err := s.incidents.SetStatus(ctx, ws.ID, id, current.Status, to, now, "")
		if err != nil {
			return err
		}
		if !ok {
			return entity.ErrIncidentInvalidTransition
		}
		if err := s.appendTimeline(ctx, ws, id, entity.IncidentEventStatusChanged, fmt.Sprintf("Status moved to %s", to), actor, now); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentStatusChanged, fmt.Sprintf("%s → %s", current.Status, to)))
	})
	if err != nil {
		return entity.Incident{}, err
	}
	return s.hydrate(ctx, ws.ID, id)
}

func (s *srv) ChangeSeverity(ctx context.Context, workspaceSlug, id, level string) (entity.Incident, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.Incident{}, err
	}
	level = strings.TrimSpace(level)
	exists, err := s.severities.Exists(ctx, ws.ID, level)
	if err != nil {
		return entity.Incident{}, err
	}
	if !exists {
		return entity.Incident{}, entity.ErrIncidentSeverityUnknown
	}
	now := time.Now().UTC()
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		ok, err := s.incidents.Patch(ctx, ws.ID, id, entity.IncidentPatch{SeverityLevel: &level})
		if err != nil {
			return err
		}
		if !ok {
			return entity.ErrIncidentNotFound
		}
		if err := s.appendTimeline(ctx, ws, id, entity.IncidentEventSeverityChanged, fmt.Sprintf("Severity set to %s", level), actor, now); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentSeverityChanged, level))
	})
	if err != nil {
		return entity.Incident{}, err
	}
	return s.hydrate(ctx, ws.ID, id)
}

func (s *srv) Resolve(ctx context.Context, workspaceSlug, id, summary string) (entity.Incident, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.Incident{}, err
	}
	summary = strings.TrimSpace(summary)
	if summary == "" || len(summary) > entity.IncidentResolutionMaxLength {
		return entity.Incident{}, entity.ErrIncidentResolutionMissing
	}
	current, err := s.incidents.GetByID(ctx, ws.ID, id)
	if err != nil {
		return entity.Incident{}, err
	}
	if !current.Status.CanTransition(entity.IncidentStatusResolved) {
		return entity.Incident{}, entity.ErrIncidentInvalidTransition
	}
	now := time.Now().UTC()
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		ok, err := s.incidents.SetStatus(ctx, ws.ID, id, current.Status, entity.IncidentStatusResolved, now, summary)
		if err != nil {
			return err
		}
		if !ok {
			return entity.ErrIncidentInvalidTransition
		}
		linked, err := s.incidents.LinkedAlertIDs(ctx, id)
		if err != nil {
			return err
		}
		resolved, err := s.alerts.Resolve(ctx, ws.ID, linked, now, entity.ResolveModeIncident)
		if err != nil {
			return err
		}
		for _, alertID := range resolved {
			if err := s.alerts.AppendEvent(ctx, alertID, entity.AlertEvent{
				At:   now,
				Kind: entity.AlertEventResolved,
				Text: fmt.Sprintf("Resolved with incident INC-%d", current.Number),
			}); err != nil {
				return err
			}
		}
		if err := s.escalations.OnResolved(ctx, ws.ID, resolved, now); err != nil {
			return err
		}
		if err := s.appendTimeline(ctx, ws, id, entity.IncidentEventResolved, "Marked resolved", actor, now); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentResolved, fmt.Sprintf("INC-%d", current.Number)))
	})
	if err != nil {
		return entity.Incident{}, err
	}
	return s.hydrate(ctx, ws.ID, id)
}

func (s *srv) Reopen(ctx context.Context, workspaceSlug, id string) (entity.Incident, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.Incident{}, err
	}
	now := time.Now().UTC()
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		ok, err := s.incidents.Reopen(ctx, ws.ID, id, entity.IncidentStatusInvestigating, now)
		if err != nil {
			return err
		}
		if !ok {
			return entity.ErrIncidentInvalidTransition
		}
		if err := s.appendTimeline(ctx, ws, id, entity.IncidentEventReopened, "Reopened for investigation", actor, now); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentReopened, id))
	})
	if err != nil {
		return entity.Incident{}, err
	}
	return s.hydrate(ctx, ws.ID, id)
}

func (s *srv) SetCustomFields(ctx context.Context, workspaceSlug, id string, fields map[string]string) (entity.Incident, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.Incident{}, err
	}
	defs, err := s.fieldDefs.List(ctx, ws.ID)
	if err != nil {
		return entity.Incident{}, err
	}
	clean, err := validateCustomFields(defs, fields)
	if err != nil {
		return entity.Incident{}, err
	}
	now := time.Now().UTC()
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		ok, err := s.incidents.SetCustomFields(ctx, ws.ID, id, clean)
		if err != nil {
			return err
		}
		if !ok {
			return entity.ErrIncidentNotFound
		}
		if err := s.appendTimeline(ctx, ws, id, entity.IncidentEventFieldsChanged, "Custom fields updated", actor, now); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentUpdated, id))
	})
	if err != nil {
		return entity.Incident{}, err
	}
	return s.hydrate(ctx, ws.ID, id)
}

func validateCustomFields(defs []entity.IncidentFieldDef, fields map[string]string) (map[string]string, error) {
	byID := make(map[string]entity.IncidentFieldDef, len(defs))
	for _, d := range defs {
		byID[d.ID] = d
	}
	out := map[string]string{}
	for id, value := range fields {
		def, ok := byID[id]
		if !ok {
			return nil, entity.ErrIncidentFieldUnknown
		}
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		switch def.Kind {
		case entity.CustomFieldSelect:
			if !containsOption(def.Options, value) {
				return nil, entity.ErrIncidentFieldValueInvalid
			}
		case entity.CustomFieldMultiSelect:
			for _, v := range strings.Split(value, ",") {
				if !containsOption(def.Options, strings.TrimSpace(v)) {
					return nil, entity.ErrIncidentFieldValueInvalid
				}
			}
		case entity.CustomFieldNumber:
			if !isNumeric(value) {
				return nil, entity.ErrIncidentFieldValueInvalid
			}
		}
		out[id] = value
	}
	return out, nil
}

func containsOption(options []string, value string) bool {
	for _, o := range options {
		if o == value {
			return true
		}
	}
	return false
}

func isNumeric(v string) bool {
	dot := false
	for i, r := range v {
		switch {
		case r >= '0' && r <= '9':
		case r == '-' && i == 0:
		case r == '.' && !dot:
			dot = true
		default:
			return false
		}
	}
	return v != "" && v != "-"
}

func (s *srv) LinkAlert(ctx context.Context, workspaceSlug, id, alertID string) (entity.Incident, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.Incident{}, err
	}
	alert, err := s.alerts.GetByID(ctx, ws.ID, alertID)
	if err != nil {
		return entity.Incident{}, err
	}
	now := time.Now().UTC()
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.incidents.GetByID(ctx, ws.ID, id); err != nil {
			return err
		}
		if err := s.incidents.LinkAlert(ctx, ws.ID, id, alert.ID); err != nil {
			return err
		}
		if err := s.appendTimeline(ctx, ws, id, entity.IncidentEventAlertLinked, fmt.Sprintf("Linked alert %q", alert.Title), actor, now); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentUpdated, id))
	})
	if err != nil {
		return entity.Incident{}, err
	}
	return s.hydrate(ctx, ws.ID, id)
}

func (s *srv) UnlinkAlert(ctx context.Context, workspaceSlug, id, alertID string) (entity.Incident, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.Incident{}, err
	}
	now := time.Now().UTC()
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.incidents.UnlinkAlert(ctx, ws.ID, id, alertID); err != nil {
			return err
		}
		if err := s.appendTimeline(ctx, ws, id, entity.IncidentEventAlertUnlinked, "Unlinked an alert", actor, now); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentUpdated, id))
	})
	if err != nil {
		return entity.Incident{}, err
	}
	return s.hydrate(ctx, ws.ID, id)
}

func (s *srv) Relate(ctx context.Context, workspaceSlug, id, relatedID string, kind entity.IncidentRelationKind) (entity.Incident, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.Incident{}, err
	}
	if !kind.Valid() {
		return entity.Incident{}, entity.ErrIncidentFieldValueInvalid
	}
	if id == relatedID {
		return entity.Incident{}, entity.ErrIncidentSelfRelation
	}
	related, err := s.incidents.GetByID(ctx, ws.ID, relatedID)
	if err != nil {
		return entity.Incident{}, err
	}
	now := time.Now().UTC()
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.incidents.GetByID(ctx, ws.ID, id); err != nil {
			return err
		}
		if _, err := s.incidents.Relate(ctx, ws.ID, id, related.ID, kind); err != nil {
			return err
		}
		if err := s.appendTimeline(ctx, ws, id, entity.IncidentEventRelated, fmt.Sprintf("Marked %s INC-%d", relationVerb(kind), related.Number), actor, now); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentUpdated, id))
	})
	if err != nil {
		return entity.Incident{}, err
	}
	return s.hydrate(ctx, ws.ID, id)
}

func relationVerb(kind entity.IncidentRelationKind) string {
	switch kind {
	case entity.IncidentRelationDuplicate:
		return "duplicate of"
	case entity.IncidentRelationCausedBy:
		return "caused by"
	default:
		return "related to"
	}
}

func (s *srv) Unrelate(ctx context.Context, workspaceSlug, id, relationID string) (entity.Incident, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.Incident{}, err
	}
	now := time.Now().UTC()
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.incidents.Unrelate(ctx, ws.ID, id, relationID); err != nil {
			return err
		}
		if err := s.appendTimeline(ctx, ws, id, entity.IncidentEventUnrelated, "Removed a related incident", actor, now); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentUpdated, id))
	})
	if err != nil {
		return entity.Incident{}, err
	}
	return s.hydrate(ctx, ws.ID, id)
}

func (s *srv) AddFollowup(ctx context.Context, workspaceSlug, id string, in entity.NewFollowup) (entity.Incident, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.Incident{}, err
	}
	in.Title = strings.TrimSpace(in.Title)
	if err := in.Validate(); err != nil {
		return entity.Incident{}, err
	}
	if err := s.validateLead(ctx, ws.ID, in.OwnerUserID); err != nil {
		return entity.Incident{}, err
	}
	now := time.Now().UTC()
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.incidents.GetByID(ctx, ws.ID, id); err != nil {
			return err
		}
		if _, err := s.incidents.AddFollowup(ctx, entity.IncidentFollowup{
			WorkspaceID: ws.ID,
			IncidentID:  id,
			Title:       in.Title,
			OwnerUserID: in.OwnerUserID,
			DueAt:       in.DueAt,
		}); err != nil {
			return err
		}
		if err := s.appendTimeline(ctx, ws, id, entity.IncidentEventFollowupAdded, fmt.Sprintf("Added follow-up: %s", in.Title), actor, now); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentUpdated, id))
	})
	if err != nil {
		return entity.Incident{}, err
	}
	return s.hydrate(ctx, ws.ID, id)
}

func (s *srv) ToggleFollowup(ctx context.Context, workspaceSlug, id, followupID string, done bool) (entity.Incident, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.Incident{}, err
	}
	now := time.Now().UTC()
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		followup, err := s.incidents.SetFollowupDone(ctx, ws.ID, followupID, done, now)
		if err != nil {
			return err
		}
		if followup.IncidentID != id {
			return entity.ErrFollowupNotFound
		}
		text := fmt.Sprintf("Reopened follow-up: %s", followup.Title)
		if done {
			text = fmt.Sprintf("Completed follow-up: %s", followup.Title)
		}
		if err := s.appendTimeline(ctx, ws, id, entity.IncidentEventFollowupDone, text, actor, now); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentUpdated, id))
	})
	if err != nil {
		return entity.Incident{}, err
	}
	return s.hydrate(ctx, ws.ID, id)
}

func (s *srv) ListFollowups(ctx context.Context, workspaceSlug string) ([]entity.IncidentFollowup, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectIncidents)
	if err != nil {
		return nil, err
	}
	followups, err := s.incidents.ListFollowups(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	names, err := s.memberNames(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	for i := range followups {
		if followups[i].OwnerUserID != "" {
			followups[i].OwnerLabel = names[followups[i].OwnerUserID]
		}
	}
	return followups, nil
}

func (s *srv) ListSeverities(ctx context.Context, workspaceSlug string) ([]entity.IncidentSeverity, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectIncidents)
	if err != nil {
		return nil, err
	}
	return s.severities.List(ctx, ws.ID)
}

func (s *srv) SaveSeverities(ctx context.Context, workspaceSlug string, severities []entity.IncidentSeverity) ([]entity.IncidentSeverity, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectSettings)
	if err != nil {
		return nil, err
	}
	if len(severities) == 0 {
		return nil, entity.ErrIncidentSeverityUnknown
	}
	seen := map[string]struct{}{}
	for i := range severities {
		severities[i].Level = strings.TrimSpace(severities[i].Level)
		severities[i].Label = strings.TrimSpace(severities[i].Label)
		if err := severities[i].Validate(); err != nil {
			return nil, err
		}
		if _, ok := seen[severities[i].Level]; ok {
			return nil, entity.ErrIncidentSeverityUnknown
		}
		seen[severities[i].Level] = struct{}{}
	}
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.severities.Replace(ctx, ws.ID, severities); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentSeveritiesSaved, workspaceSlug))
	})
	if err != nil {
		return nil, err
	}
	return s.severities.List(ctx, ws.ID)
}

func (s *srv) ListFieldDefs(ctx context.Context, workspaceSlug string) ([]entity.IncidentFieldDef, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectIncidents)
	if err != nil {
		return nil, err
	}
	return s.fieldDefs.List(ctx, ws.ID)
}

func (s *srv) SaveFieldDefs(ctx context.Context, workspaceSlug string, defs []entity.IncidentFieldDef) ([]entity.IncidentFieldDef, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectSettings)
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	for i := range defs {
		defs[i].Name = strings.TrimSpace(defs[i].Name)
		if err := defs[i].Validate(); err != nil {
			return nil, err
		}
		slug := strings.TrimSpace(defs[i].Slug)
		if slug == "" {
			slug = entity.Slugify(defs[i].Name)
		}
		defs[i].Slug = slug
		if _, ok := seen[slug]; ok {
			return nil, entity.ErrFieldSlugTaken
		}
		seen[slug] = struct{}{}
	}
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.fieldDefs.Replace(ctx, ws.ID, defs); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentFieldsSaved, workspaceSlug))
	})
	if err != nil {
		return nil, err
	}
	return s.fieldDefs.List(ctx, ws.ID)
}
