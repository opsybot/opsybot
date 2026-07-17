package http

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/service"
)

type ssoRoutes struct {
	sso service.SSO
	cfg config.Auth
}

func (h *ssoRoutes) start(w http.ResponseWriter, r *http.Request) {
	redirect, err := h.sso.StartLogin(r.Context(), chi.URLParam(r, "workspace"))
	if err != nil {
		h.fail(w, r, err)
		return
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (h *ssoRoutes) callback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("error") != "" {
		h.redirect(w, r, "idp_error")
		return
	}
	code, state := q.Get("code"), q.Get("state")
	if code == "" || state == "" {
		h.redirect(w, r, "invalid_state")
		return
	}
	info := entity.RequestInfoFrom(r.Context())
	result, err := h.sso.CompleteLogin(r.Context(), chi.URLParam(r, "workspace"), code, state, info.IP, info.UserAgent)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.CookieName,
		Value:    result.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  result.Session.ExpiresAt,
	})
	http.Redirect(w, r, h.base()+"/"+chi.URLParam(r, "workspace"), http.StatusFound)
}

func (h *ssoRoutes) fail(w http.ResponseWriter, r *http.Request, err error) {
	logger.From(r.Context()).WarnContext(r.Context(), "sso login failed", "error", err, "path", r.URL.Path)
	h.redirect(w, r, ssoErrorCode(err))
}

func (h *ssoRoutes) redirect(w http.ResponseWriter, r *http.Request, code string) {
	http.Redirect(w, r, h.base()+"/sso-error?code="+url.QueryEscape(code), http.StatusFound)
}

func (h *ssoRoutes) base() string {
	return strings.TrimRight(h.cfg.BaseURL, "/")
}

func ssoErrorCode(err error) string {
	switch {
	case errors.Is(err, entity.ErrSSONotConfigured), errors.Is(err, entity.ErrSSONotEnabled), errors.Is(err, entity.ErrSSOInvalid):
		return "not_enabled"
	case errors.Is(err, entity.ErrSSOStateInvalid):
		return "invalid_state"
	case errors.Is(err, entity.ErrSSOExchange):
		return "exchange_failed"
	case errors.Is(err, entity.ErrSSOProvisioningDisabled):
		return "not_provisioned"
	case errors.Is(err, entity.ErrSSODomainNotAllowed):
		return "domain_not_allowed"
	case errors.Is(err, entity.ErrMemberDeactivated):
		return "deactivated"
	case errors.Is(err, entity.ErrSSOEmailMissing):
		return "email_missing"
	case errors.Is(err, entity.ErrWorkspaceNotFound):
		return "not_found"
	default:
		return "error"
	}
}
