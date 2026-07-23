package actions

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	tokens     repository.ActionToken
	workspaces repository.Workspace
	members    repository.Member
	alerts     repository.Alert
	alertSvc   service.Alerts
}

func New(
	tokens repository.ActionToken,
	workspaces repository.Workspace,
	members repository.Member,
	alerts repository.Alert,
	alertSvc service.Alerts,
) service.Actions {
	return &srv{tokens: tokens, workspaces: workspaces, members: members, alerts: alerts, alertSvc: alertSvc}
}

func (s *srv) Redeem(ctx context.Context, rawToken, ip string) (entity.ActionOutcome, error) {
	now := time.Now().UTC()
	claim, err := s.tokens.Consume(ctx, entity.HashToken(rawToken), ip, now)
	if err != nil {
		return entity.ActionOutcome{}, err
	}
	ws, err := s.workspaces.GetByID(ctx, claim.WorkspaceID)
	if err != nil {
		return entity.ActionOutcome{}, err
	}
	member, err := s.members.Get(ctx, claim.WorkspaceID, claim.UserID)
	if err != nil {
		return entity.ActionOutcome{}, err
	}
	if member.Status != entity.MemberStatusActive {
		return entity.ActionOutcome{}, entity.ErrNotMember
	}
	alert, err := s.alerts.GetByID(ctx, claim.WorkspaceID, claim.AlertID)
	if err != nil {
		return entity.ActionOutcome{}, err
	}
	actorCtx := entity.WithIdentity(ctx, entity.Identity{
		Kind: entity.IdentityKindSession, UserID: claim.UserID, Label: member.Name, IP: ip,
	})
	switch claim.Action {
	case entity.ActionKindResolve:
		if _, err := s.alertSvc.Resolve(actorCtx, ws.Slug, []string{claim.AlertID}); err != nil {
			return entity.ActionOutcome{}, err
		}
	default:
		if _, err := s.alertSvc.Acknowledge(actorCtx, ws.Slug, []string{claim.AlertID}); err != nil {
			return entity.ActionOutcome{}, err
		}
	}
	return entity.ActionOutcome{Action: claim.Action, AlertTitle: alert.Title, Actor: member.Name, At: now}, nil
}
