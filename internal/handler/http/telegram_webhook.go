package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/service"
)

const maxTelegramBody = 1 << 20

type telegramWebhookRoutes struct {
	chats   service.Chats
	actions service.Actions
	cfg     config.Telegram
}

func (h *telegramWebhookRoutes) handle(w http.ResponseWriter, r *http.Request) {
	secret := entity.TelegramWebhookSecret(h.cfg.BotToken)
	if h.cfg.BotToken == "" || chi.URLParam(r, "secret") != secret {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if hdr := r.Header.Get("X-Telegram-Bot-Api-Secret-Token"); hdr != "" && hdr != secret {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	body, _ := io.ReadAll(io.LimitReader(r.Body, maxTelegramBody))
	var update struct {
		Message *struct {
			Text string `json:"text"`
			From struct {
				ID        int64  `json:"id"`
				Username  string `json:"username"`
				FirstName string `json:"first_name"`
			} `json:"from"`
		} `json:"message"`
		CallbackQuery *struct {
			ID   string `json:"id"`
			Data string `json:"data"`
		} `json:"callback_query"`
	}
	if err := json.Unmarshal(body, &update); err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	ctx := r.Context()
	switch {
	case update.Message != nil && strings.HasPrefix(update.Message.Text, "/start "):
		token := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/start "))
		handle := update.Message.From.Username
		if handle == "" {
			handle = update.Message.From.FirstName
		}
		if err := h.chats.CompleteTelegramLink(ctx, token, strconv.FormatInt(update.Message.From.ID, 10), handle); err != nil {
			logger.From(ctx).WarnContext(ctx, "telegram link failed", "error", err)
		}
	case update.CallbackQuery != nil && update.CallbackQuery.Data != "":
		outcome, err := h.actions.Redeem(ctx, update.CallbackQuery.Data, entity.RequestInfoFrom(ctx).IP)
		answer := "That action link has already been used or expired."
		if err != nil {
			logger.From(ctx).WarnContext(ctx, "telegram action failed", "error", err)
		} else {
			verb := "Acknowledged"
			if outcome.Action == entity.ActionKindResolve {
				verb = "Resolved"
			}
			answer = verb + ": " + outcome.AlertTitle
		}
		if aErr := h.chats.AnswerTelegramCallback(ctx, update.CallbackQuery.ID, answer); aErr != nil {
			logger.From(ctx).WarnContext(ctx, "telegram callback answer failed", "error", aErr)
		}
	}
	w.WriteHeader(http.StatusOK)
}
