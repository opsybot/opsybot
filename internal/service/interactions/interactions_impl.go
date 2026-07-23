package interactions

import (
	"context"
	"encoding/json"
	"net/url"
	"time"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	actions service.Actions
	slack   config.Slack
	discord config.Discord
	chat    config.Chat
}

func New(actions service.Actions, slack config.Slack, discord config.Discord, chat config.Chat) service.Interactions {
	return &srv{actions: actions, slack: slack, discord: discord, chat: chat}
}

func jsonResponse(status int, payload any) entity.InteractionResponse {
	body, _ := json.Marshal(payload)
	return entity.InteractionResponse{Status: status, ContentType: "application/json", Body: body}
}

func (s *srv) Slack(ctx context.Context, cb entity.ChatCallback) (entity.InteractionResponse, error) {
	if !entity.VerifySlackSignature(s.slack.SigningSecret, cb.Timestamp, cb.Signature, cb.Body, time.Now().UTC(), s.chat.InteractionSkew) {
		return entity.InteractionResponse{Status: 401}, nil
	}
	if challenge, ok := slackURLVerification(cb.Body); ok {
		return entity.InteractionResponse{Status: 200, ContentType: "text/plain", Body: []byte(challenge)}, nil
	}
	form, err := url.ParseQuery(string(cb.Body))
	if err != nil {
		return entity.InteractionResponse{Status: 400}, nil
	}
	var payload struct {
		Type    string `json:"type"`
		Actions []struct {
			ActionID string `json:"action_id"`
			Value    string `json:"value"`
		} `json:"actions"`
	}
	if err := json.Unmarshal([]byte(form.Get("payload")), &payload); err != nil {
		return entity.InteractionResponse{Status: 400}, nil
	}
	if len(payload.Actions) == 0 {
		return entity.InteractionResponse{Status: 200}, nil
	}
	outcome, err := s.actions.Redeem(ctx, payload.Actions[0].Value, cb.IP)
	text := interactionText(outcome, err)
	if err != nil {
		logger.From(ctx).WarnContext(ctx, "slack action redeem failed", "error", err)
	}
	return jsonResponse(200, map[string]any{"replace_original": true, "text": text}), nil
}

func slackURLVerification(body []byte) (string, bool) {
	var probe struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return "", false
	}
	if probe.Type == "url_verification" {
		return probe.Challenge, true
	}
	return "", false
}

func (s *srv) Discord(ctx context.Context, cb entity.ChatCallback) (entity.InteractionResponse, error) {
	if !entity.VerifyDiscordSignature(s.discord.PublicKey, cb.Timestamp, cb.Signature, cb.Body) {
		return entity.InteractionResponse{Status: 401}, nil
	}
	var payload struct {
		Type int `json:"type"`
		Data struct {
			CustomID string `json:"custom_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(cb.Body, &payload); err != nil {
		return entity.InteractionResponse{Status: 400}, nil
	}
	switch payload.Type {
	case 1:
		return jsonResponse(200, map[string]any{"type": 1}), nil
	case 3:
		outcome, err := s.actions.Redeem(ctx, payload.Data.CustomID, cb.IP)
		text := interactionText(outcome, err)
		if err != nil {
			logger.From(ctx).WarnContext(ctx, "discord action redeem failed", "error", err)
		}
		return jsonResponse(200, map[string]any{"type": 7, "data": map[string]any{"content": text, "components": []any{}}}), nil
	default:
		return jsonResponse(200, map[string]any{"type": 4, "data": map[string]any{"content": "Unsupported interaction."}}), nil
	}
}

func interactionText(outcome entity.ActionOutcome, err error) string {
	if err != nil {
		return "That action link has already been used or expired."
	}
	verb := "Acknowledged"
	if outcome.Action == entity.ActionKindResolve {
		verb = "Resolved"
	}
	return verb + " by " + outcome.Actor + " at " + outcome.At.Format("15:04") + " UTC."
}
