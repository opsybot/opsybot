package http

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/service"
)

type chatOAuthRoutes struct {
	chats service.Chats
	auth  service.Auth
	cfg   config.Auth
}

func (h *chatOAuthRoutes) callback(provider entity.ChatProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := h.withSession(r)
		q := r.URL.Query()
		if q.Get("error") != "" {
			h.redirect(w, r, "", "denied")
			return
		}
		code, state := q.Get("code"), q.Get("state")
		if code == "" || state == "" {
			h.redirect(w, r, "", "invalid_state")
			return
		}
		slug, err := h.chats.CompleteOAuth(ctx, provider, code, state)
		if err != nil {
			logger.From(ctx).WarnContext(ctx, "chat oauth failed", "error", err, "provider", string(provider))
			h.redirect(w, r, slug, chatOAuthErrorCode(err))
			return
		}
		http.Redirect(w, r, h.base()+"/"+slug+"/chat?connected="+url.QueryEscape(string(provider)), http.StatusFound)
	}
}

func (h *chatOAuthRoutes) identityCallback(provider entity.ChatProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := h.withSession(r)
		q := r.URL.Query()
		if q.Get("error") != "" {
			h.redirect(w, r, "", "denied")
			return
		}
		code, state := q.Get("code"), q.Get("state")
		if code == "" || state == "" {
			h.redirect(w, r, "", "invalid_state")
			return
		}
		slug, err := h.chats.CompleteIdentityOAuth(ctx, provider, code, state)
		if err != nil {
			logger.From(ctx).WarnContext(ctx, "chat identity oauth failed", "error", err, "provider", string(provider))
			h.redirect(w, r, slug, chatOAuthErrorCode(err))
			return
		}
		http.Redirect(w, r, h.base()+"/"+slug+"/chat?linked="+url.QueryEscape(string(provider)), http.StatusFound)
	}
}

func (h *chatOAuthRoutes) withSession(r *http.Request) context.Context {
	ctx := r.Context()
	cookie, err := r.Cookie(h.cfg.CookieName)
	if err != nil {
		return ctx
	}
	id, err := h.auth.Resolve(ctx, cookie.Value)
	if err != nil {
		return ctx
	}
	info := entity.RequestInfoFrom(ctx)
	id.IP = info.IP
	id.UserAgent = info.UserAgent
	return entity.WithIdentity(ctx, id)
}

func (h *chatOAuthRoutes) redirect(w http.ResponseWriter, r *http.Request, slug, code string) {
	var dest string
	if slug != "" {
		dest = h.base() + "/" + slug + "/chat?chat_error=" + url.QueryEscape(code)
	} else {
		dest = h.base() + "/chat-error?code=" + url.QueryEscape(code)
	}
	http.Redirect(w, r, dest, http.StatusFound)
}

func (h *chatOAuthRoutes) base() string {
	return strings.TrimRight(h.cfg.BaseURL, "/")
}

func chatOAuthErrorCode(err error) string {
	switch {
	case errors.Is(err, entity.ErrChatOAuthStateInvalid):
		return "invalid_state"
	case errors.Is(err, entity.ErrForbidden):
		return "forbidden"
	case errors.Is(err, entity.ErrChatOAuthExchange):
		return "exchange_failed"
	case errors.Is(err, entity.ErrChatProviderNotConfigured):
		return "not_configured"
	case errors.Is(err, entity.ErrChatSecretUnavailable):
		return "secret_unavailable"
	default:
		return "error"
	}
}
