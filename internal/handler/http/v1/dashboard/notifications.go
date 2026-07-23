package dashboard

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func notificationProblem(err error) (int, api.Problem) {
	switch {
	case errors.Is(err, entity.ErrForbidden):
		return http.StatusForbidden, prob(http.StatusForbidden, "Forbidden", "You do not have access to notifications in this workspace.", "")
	case errors.Is(err, entity.ErrUnauthenticated):
		return http.StatusUnauthorized, prob(http.StatusUnauthorized, "Unauthenticated", "Sign in to continue.", "")
	case isValidation(err):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Invalid rules", validationDetail(err), "")
	case errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
		return http.StatusNotFound, prob(http.StatusNotFound, "Not found", "That workspace does not exist.", "")
	default:
		return 0, api.Problem{}
	}
}

func (h *handler) GetNotificationRules(ctx context.Context, request api.GetNotificationRulesRequestObject) (api.GetNotificationRulesResponseObject, error) {
	settings, err := h.notifications.Get(ctx, request.WorkspaceId)
	if err != nil {
		status, p := notificationProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.GetNotificationRules401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.GetNotificationRules403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.GetNotificationRules404ApplicationProblemPlusJSONResponse(p), nil
		default:
			return nil, err
		}
	}
	channels := make([]api.Channel, 0, len(settings.Channels))
	for _, c := range settings.Channels {
		channels = append(channels, channelDTO(c))
	}
	return api.GetNotificationRules200JSONResponse{
		Rules:    notificationRulesDTO(settings.Rule),
		Channels: channels,
	}, nil
}

func (h *handler) PutNotificationRules(ctx context.Context, request api.PutNotificationRulesRequestObject) (api.PutNotificationRulesResponseObject, error) {
	if request.Body == nil {
		return api.PutNotificationRules400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	saved, err := h.notifications.Save(ctx, request.WorkspaceId, notificationRuleFromDTO(*request.Body))
	if err != nil {
		status, p := notificationProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.PutNotificationRules400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.PutNotificationRules401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.PutNotificationRules403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.PutNotificationRules404ApplicationProblemPlusJSONResponse(p), nil
		default:
			return nil, err
		}
	}
	return api.PutNotificationRules200JSONResponse(notificationRulesDTO(saved)), nil
}

func notificationRulesDTO(rule entity.NotificationRule) api.NotificationRules {
	return api.NotificationRules{
		High:       notificationStepsDTO(rule.High),
		Low:        notificationStepsDTO(rule.Low),
		QuietHours: quietHoursDTO(rule.QuietHours),
	}
}

func notificationStepsDTO(steps []entity.NotificationStep) []api.NotificationStep {
	out := make([]api.NotificationStep, 0, len(steps))
	for _, s := range steps {
		out = append(out, api.NotificationStep{
			ChannelType:  api.NotificationStepChannelType(s.Channel),
			DelayMinutes: int(s.Delay / time.Minute),
		})
	}
	return out
}

func quietHoursDTO(q entity.QuietHours) api.NotificationQuietHours {
	days := make([]int, 0, len(q.Window.Days))
	days = append(days, q.Window.Days...)
	return api.NotificationQuietHours{
		Enabled:     q.Enabled,
		Days:        days,
		StartMinute: q.Window.StartMinute,
		EndMinute:   q.Window.EndMinute,
		Timezone:    q.Window.Timezone,
	}
}

func notificationRuleFromDTO(in api.NotificationRules) entity.NotificationRule {
	return entity.NotificationRule{
		High:       notificationStepsFromDTO(in.High),
		Low:        notificationStepsFromDTO(in.Low),
		QuietHours: quietHoursFromDTO(in.QuietHours),
	}
}

func notificationStepsFromDTO(steps []api.NotificationStep) []entity.NotificationStep {
	out := make([]entity.NotificationStep, 0, len(steps))
	for _, s := range steps {
		out = append(out, entity.NotificationStep{
			Channel: entity.ChannelType(s.ChannelType),
			Delay:   time.Duration(s.DelayMinutes) * time.Minute,
		})
	}
	return out
}

func quietHoursFromDTO(in api.NotificationQuietHours) entity.QuietHours {
	days := make([]int, 0, len(in.Days))
	days = append(days, in.Days...)
	return entity.QuietHours{
		Enabled: in.Enabled,
		Window: entity.HoursWindow{
			Days:        days,
			StartMinute: in.StartMinute,
			EndMinute:   in.EndMinute,
			Timezone:    in.Timezone,
		},
	}
}
