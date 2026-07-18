package dashboard

import (
	"errors"
	"net/http"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func isValidation(err error) bool {
	return entity.IsValidationError(err) || errors.Is(err, entity.ErrWorkspaceSlugInvalid)
}

func validationDetail(err error) string {
	if entity.IsValidationError(err) {
		return entity.ValidationMessage(err)
	}
	if errors.Is(err, entity.ErrWorkspaceSlugInvalid) {
		return "A workspace URL uses lowercase letters, numbers, and hyphens, and starts with a letter."
	}
	return "One or more fields are invalid. Check your input and try again."
}

const problemBase = "https://opsybot.dev/problems/"

func prob(status int, title, detail, problemType string) api.Problem {
	pt := "about:blank"
	if problemType != "" {
		pt = problemBase + problemType
	}
	return api.Problem{Status: status, Title: title, Detail: &detail, Type: pt}
}

func ptr[T any](v T) *T { return &v }

func (h *handler) sessionCookie(token string, expires time.Time) string {
	c := &http.Cookie{
		Name:     h.cfg.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
	}
	return c.String()
}

const pendingCookieName = "opsybot_2fa"

func (h *handler) pendingCookie(token string) string {
	c := &http.Cookie{
		Name:     pendingCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((5 * time.Minute).Seconds()),
	}
	return c.String()
}

func (h *handler) clearPendingCookie() string {
	c := &http.Cookie{Name: pendingCookieName, Value: "", Path: "/", HttpOnly: true, Secure: h.cfg.CookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: -1}
	return c.String()
}

func (h *handler) clearCookie() string {
	c := &http.Cookie{
		Name:     h.cfg.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
	return c.String()
}
