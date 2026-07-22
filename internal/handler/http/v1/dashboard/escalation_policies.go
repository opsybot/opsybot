package dashboard

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func escalationProblem(err error) (int, api.Problem) {
	switch {
	case errors.Is(err, entity.ErrForbidden):
		return http.StatusForbidden, prob(http.StatusForbidden, "Forbidden", "You do not have access to escalation policies in this workspace.", "")
	case errors.Is(err, entity.ErrUnauthenticated):
		return http.StatusUnauthorized, prob(http.StatusUnauthorized, "Unauthenticated", "Sign in to continue.", "")
	case isValidation(err):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Invalid policy", validationDetail(err), "")
	case errors.Is(err, entity.ErrEscalationPolicySlugTaken):
		return http.StatusConflict, prob(http.StatusConflict, "Name taken", "A policy already goes by that name.", "")
	case errors.Is(err, entity.ErrEscalationPolicyReferenced):
		return http.StatusConflict, prob(http.StatusConflict, "Policy in use", "Routing rules, monitors, or the default route still point at this policy. Re-point them first.", "")
	case errors.Is(err, entity.ErrEscalationWebhookSlugTaken):
		return http.StatusConflict, prob(http.StatusConflict, "Name taken", "A webhook already goes by that name.", "")
	case errors.Is(err, entity.ErrEscalationSecretUnavailable):
		return http.StatusConflict, prob(http.StatusConflict, "Secret storage unavailable", "This instance has no auth secret key configured, so the signing secret can't be stored. Ask an admin to set OPSYBOT_AUTH_SECRET_KEY, or add the webhook without a secret.", "")
	case errors.Is(err, entity.ErrEscalationWebhookInUse):
		return http.StatusConflict, prob(http.StatusConflict, "Webhook in use", "An escalation policy still targets this webhook. Remove it from the policy first.", "")
	case errors.Is(err, entity.ErrEscalationRunFinished):
		return http.StatusConflict, prob(http.StatusConflict, "Escalation finished", "This alert's escalation already reached a final state.", "")
	case errors.Is(err, entity.ErrTeamNotFound):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Unknown team", "Pick an existing team for this policy.", "")
	case errors.Is(err, entity.ErrEscalationPolicyNotFound), errors.Is(err, entity.ErrEscalationWebhookNotFound),
		errors.Is(err, entity.ErrEscalationRunNotFound), errors.Is(err, entity.ErrAlertNotFound),
		errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
		return http.StatusNotFound, prob(http.StatusNotFound, "Not found", "That escalation policy does not exist.", "")
	default:
		return 0, api.Problem{}
	}
}

func escalationNodesDTO(nodes []entity.EscalationNode) []api.EscalationNode {
	out := make([]api.EscalationNode, 0, len(nodes))
	for _, node := range nodes {
		switch {
		case node.Level != nil:
			targets := make([]api.EscalationTarget, 0, len(node.Level.Targets))
			for _, t := range node.Level.Targets {
				targets = append(targets, api.EscalationTarget{Type: api.EscalationTargetType(t.Type), Ref: t.Ref})
			}
			mode := api.EscalationNodeMode(node.Level.Mode)
			wait := int(node.Level.Wait / time.Second)
			out = append(out, api.EscalationNode{
				Type: api.Level, Id: node.Level.ID,
				Targets: &targets, Mode: &mode, WaitSeconds: &wait,
			})
		case node.Branch != nil:
			lanes := make([]api.EscalationLane, 0, len(node.Branch.Lanes))
			for _, lane := range node.Branch.Lanes {
				lanes = append(lanes, api.EscalationLane{Id: lane.ID, Key: lane.Key, Nodes: escalationNodesDTO(lane.Nodes)})
			}
			on := api.EscalationNodeOn(node.Branch.On)
			n := api.EscalationNode{Type: api.Branch, Id: node.Branch.ID, On: &on, Lanes: &lanes}
			if node.Branch.On == entity.EscalationBranchHours {
				n.Hours = &api.EscalationHours{
					Days:        node.Branch.Hours.Days,
					StartMinute: node.Branch.Hours.StartMinute,
					EndMinute:   node.Branch.Hours.EndMinute,
					Timezone:    node.Branch.Hours.Timezone,
				}
			}
			out = append(out, n)
		}
	}
	return out
}

func escalationNodesFromDTO(nodes []api.EscalationNode) []entity.EscalationNode {
	out := make([]entity.EscalationNode, 0, len(nodes))
	for _, n := range nodes {
		switch n.Type {
		case api.Level:
			level := &entity.EscalationLevel{ID: n.Id, Mode: entity.NotifyModeAll, Wait: 5 * time.Minute}
			if n.Targets != nil {
				for _, t := range *n.Targets {
					level.Targets = append(level.Targets, entity.EscalationTarget{Type: entity.EscalationTargetType(t.Type), Ref: t.Ref})
				}
			}
			if n.Mode != nil {
				level.Mode = entity.NotifyMode(*n.Mode)
			}
			if n.WaitSeconds != nil {
				level.Wait = time.Duration(*n.WaitSeconds) * time.Second
			}
			out = append(out, entity.EscalationNode{Level: level})
		case api.Branch:
			branch := &entity.EscalationBranch{ID: n.Id}
			if n.On != nil {
				branch.On = entity.EscalationBranchKind(*n.On)
			}
			if n.Hours != nil {
				branch.Hours = entity.HoursWindow{
					Days:        n.Hours.Days,
					StartMinute: n.Hours.StartMinute,
					EndMinute:   n.Hours.EndMinute,
					Timezone:    n.Hours.Timezone,
				}
			} else if branch.On == entity.EscalationBranchHours {
				branch.Hours = entity.DefaultHoursWindow()
			}
			if n.Lanes != nil {
				for _, lane := range *n.Lanes {
					branch.Lanes = append(branch.Lanes, entity.EscalationLane{
						ID: lane.Id, Key: lane.Key, Nodes: escalationNodesFromDTO(lane.Nodes),
					})
				}
			}
			out = append(out, entity.EscalationNode{Branch: branch})
		}
	}
	return out
}

func escalationPolicyDTO(p entity.EscalationPolicy) api.EscalationPolicy {
	return api.EscalationPolicy{
		Id:                p.ID,
		Slug:              p.Slug,
		Name:              p.Name,
		TeamSlug:          p.TeamSlug,
		Repeat:            p.Repeat,
		AckTimeoutSeconds: int(p.AckTimeout / time.Second),
		Nodes:             escalationNodesDTO(p.Nodes),
	}
}

func escalationPolicyFromRequest(body *api.SaveEscalationPolicyRequest) entity.EscalationPolicy {
	p := entity.EscalationPolicy{
		Name:  body.Name,
		Nodes: escalationNodesFromDTO(body.Nodes),
	}
	if body.Slug != nil {
		p.Slug = *body.Slug
	}
	if body.TeamSlug != nil {
		p.TeamSlug = *body.TeamSlug
	}
	if body.Repeat != nil {
		p.Repeat = *body.Repeat
	}
	if body.AckTimeoutSeconds != nil {
		p.AckTimeout = time.Duration(*body.AckTimeoutSeconds) * time.Second
	}
	return p
}

func (h *handler) ListEscalationPolicies(ctx context.Context, request api.ListEscalationPoliciesRequestObject) (api.ListEscalationPoliciesResponseObject, error) {
	list, err := h.escalations.List(ctx, request.WorkspaceId)
	if err != nil {
		status, p := escalationProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListEscalationPolicies401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListEscalationPolicies403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListEscalationPolicies404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	items := make([]api.EscalationPolicySummary, 0, len(list))
	for _, s := range list {
		items = append(items, api.EscalationPolicySummary{
			Id: s.ID, Slug: s.Slug, Name: s.Name, TeamSlug: s.TeamSlug,
			Routed: s.Routed, StepCount: s.StepCount, HasBranch: s.HasBranch,
			Nodes: escalationNodesDTO(s.Nodes),
		})
	}
	return api.ListEscalationPolicies200JSONResponse{Items: items}, nil
}

func (h *handler) GetEscalationPolicy(ctx context.Context, request api.GetEscalationPolicyRequestObject) (api.GetEscalationPolicyResponseObject, error) {
	detail, err := h.escalations.Get(ctx, request.WorkspaceId, request.PolicySlug)
	if err != nil {
		status, p := escalationProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.GetEscalationPolicy401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.GetEscalationPolicy403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.GetEscalationPolicy404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}

	routes := make([]api.AlertRoute, 0, len(detail.Routes))
	for _, r := range detail.Routes {
		routes = append(routes, routeDTO(r))
	}
	recent := make([]api.RecentEscalation, 0, len(detail.Recent))
	for _, r := range detail.Recent {
		item := api.RecentEscalation{
			AlertId:    r.AlertID,
			AlertTitle: r.AlertTitle,
			StartedAt:  r.StartedAt,
			State:      api.RecentEscalationState(r.State),
			Outcome:    r.Outcome,
		}
		if !r.EndedAt.IsZero() {
			at := r.EndedAt
			item.EndedAt = &at
		}
		if r.ByLabel != "" {
			by := r.ByLabel
			item.By = &by
		}
		recent = append(recent, item)
	}
	return api.GetEscalationPolicy200JSONResponse{
		Policy:    escalationPolicyDTO(detail.Policy),
		Routes:    routes,
		Recent:    recent,
		Routed:    detail.Routed,
		IsDefault: detail.IsDefault,
	}, nil
}

func (h *handler) CreateEscalationPolicy(ctx context.Context, request api.CreateEscalationPolicyRequestObject) (api.CreateEscalationPolicyResponseObject, error) {
	created, err := h.escalations.Create(ctx, request.WorkspaceId, escalationPolicyFromRequest(request.Body))
	if err != nil {
		status, p := escalationProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.CreateEscalationPolicy400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.CreateEscalationPolicy401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.CreateEscalationPolicy403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.CreateEscalationPolicy404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.CreateEscalationPolicy409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.CreateEscalationPolicy201JSONResponse(escalationPolicyDTO(created)), nil
}

func (h *handler) UpdateEscalationPolicy(ctx context.Context, request api.UpdateEscalationPolicyRequestObject) (api.UpdateEscalationPolicyResponseObject, error) {
	updated, err := h.escalations.Update(ctx, request.WorkspaceId, request.PolicySlug, escalationPolicyFromRequest(request.Body))
	if err != nil {
		status, p := escalationProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.UpdateEscalationPolicy400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.UpdateEscalationPolicy401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.UpdateEscalationPolicy403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.UpdateEscalationPolicy404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.UpdateEscalationPolicy200JSONResponse(escalationPolicyDTO(updated)), nil
}

func (h *handler) DeleteEscalationPolicy(ctx context.Context, request api.DeleteEscalationPolicyRequestObject) (api.DeleteEscalationPolicyResponseObject, error) {
	if err := h.escalations.Delete(ctx, request.WorkspaceId, request.PolicySlug); err != nil {
		status, p := escalationProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.DeleteEscalationPolicy401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.DeleteEscalationPolicy403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.DeleteEscalationPolicy404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.DeleteEscalationPolicy409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.DeleteEscalationPolicy204Response{}, nil
}

func (h *handler) GetEscalationDirectory(ctx context.Context, request api.GetEscalationDirectoryRequestObject) (api.GetEscalationDirectoryResponseObject, error) {
	directory, err := h.escalations.Directory(ctx, request.WorkspaceId)
	if err != nil {
		status, p := escalationProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.GetEscalationDirectory401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.GetEscalationDirectory403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.GetEscalationDirectory404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}

	members := make([]api.EscalationDirectoryMember, 0, len(directory.Members))
	for _, m := range directory.Members {
		members = append(members, api.EscalationDirectoryMember{
			Id: m.UserID, Name: m.Name, Email: m.Email, Active: m.Status == entity.MemberStatusActive,
		})
	}
	schedules := make([]api.EscalationDirectoryEntry, 0, len(directory.Schedules))
	for _, s := range directory.Schedules {
		schedules = append(schedules, api.EscalationDirectoryEntry{Id: s.ID, Slug: s.Slug, Name: s.Slug})
	}
	teams := make([]api.EscalationDirectoryEntry, 0, len(directory.Teams))
	for _, t := range directory.Teams {
		teams = append(teams, api.EscalationDirectoryEntry{Id: t.ID, Slug: t.Slug, Name: t.Name})
	}
	webhooks := make([]api.EscalationDirectoryEntry, 0, len(directory.Webhooks))
	for _, w := range directory.Webhooks {
		webhooks = append(webhooks, api.EscalationDirectoryEntry{Id: w.ID, Slug: w.Slug, Name: w.Name})
	}
	return api.GetEscalationDirectory200JSONResponse{
		Members: members, Schedules: schedules, Teams: teams, Webhooks: webhooks,
	}, nil
}

func webhookDTO(w entity.EscalationWebhook) api.EscalationWebhook {
	return api.EscalationWebhook{Id: w.ID, Slug: w.Slug, Name: w.Name, Url: w.URL, HasSecret: w.Secret != ""}
}

func (h *handler) ListEscalationWebhooks(ctx context.Context, request api.ListEscalationWebhooksRequestObject) (api.ListEscalationWebhooksResponseObject, error) {
	hooks, err := h.escalations.ListWebhooks(ctx, request.WorkspaceId)
	if err != nil {
		status, p := escalationProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListEscalationWebhooks401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListEscalationWebhooks403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListEscalationWebhooks404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	items := make([]api.EscalationWebhook, 0, len(hooks))
	for _, w := range hooks {
		items = append(items, webhookDTO(w))
	}
	return api.ListEscalationWebhooks200JSONResponse{Items: items}, nil
}

func (h *handler) CreateEscalationWebhook(ctx context.Context, request api.CreateEscalationWebhookRequestObject) (api.CreateEscalationWebhookResponseObject, error) {
	in := entity.NewEscalationWebhook{Name: request.Body.Name, URL: request.Body.Url}
	if request.Body.Slug != nil {
		in.Slug = *request.Body.Slug
	}
	secret := ""
	if request.Body.Secret != nil {
		secret = *request.Body.Secret
	}
	created, err := h.escalations.CreateWebhook(ctx, request.WorkspaceId, in, secret)
	if err != nil {
		status, p := escalationProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.CreateEscalationWebhook400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.CreateEscalationWebhook401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.CreateEscalationWebhook403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.CreateEscalationWebhook404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.CreateEscalationWebhook409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.CreateEscalationWebhook201JSONResponse(webhookDTO(created)), nil
}

func (h *handler) UpdateEscalationWebhook(ctx context.Context, request api.UpdateEscalationWebhookRequestObject) (api.UpdateEscalationWebhookResponseObject, error) {
	in := entity.NewEscalationWebhook{Name: request.Body.Name, URL: request.Body.Url}
	updated, err := h.escalations.UpdateWebhook(ctx, request.WorkspaceId, request.WebhookSlug, in)
	if err != nil {
		status, p := escalationProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.UpdateEscalationWebhook400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.UpdateEscalationWebhook401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.UpdateEscalationWebhook403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.UpdateEscalationWebhook404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.UpdateEscalationWebhook200JSONResponse(webhookDTO(updated)), nil
}

func (h *handler) DeleteEscalationWebhook(ctx context.Context, request api.DeleteEscalationWebhookRequestObject) (api.DeleteEscalationWebhookResponseObject, error) {
	if err := h.escalations.DeleteWebhook(ctx, request.WorkspaceId, request.WebhookSlug); err != nil {
		status, p := escalationProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.DeleteEscalationWebhook401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.DeleteEscalationWebhook403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.DeleteEscalationWebhook404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.DeleteEscalationWebhook409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.DeleteEscalationWebhook204Response{}, nil
}

func (h *handler) EscalateAlert(ctx context.Context, request api.EscalateAlertRequestObject) (api.EscalateAlertResponseObject, error) {
	if err := h.escalations.Escalate(ctx, request.WorkspaceId, request.AlertId); err != nil {
		status, p := escalationProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.EscalateAlert401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.EscalateAlert403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.EscalateAlert404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.EscalateAlert409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.EscalateAlert204Response{}, nil
}
