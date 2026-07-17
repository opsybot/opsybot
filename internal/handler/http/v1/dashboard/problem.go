package dashboard

import (
	"errors"
	"net/http"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

var validationErrors = []error{
	entity.ErrUserInvalidEmail, entity.ErrUserInvalidName, entity.ErrUserWeakPassword,
	entity.ErrUserInvalidTimezone, entity.ErrWorkspaceSlugInvalid, entity.ErrWorkspaceSlugReserved,
	entity.ErrWorkspaceNameInvalid, entity.ErrRoleInvalid,
}

func isValidation(err error) bool {
	for _, v := range validationErrors {
		if errors.Is(err, v) {
			return true
		}
	}
	return false
}

func validationDetail(err error) string {
	switch {
	case errors.Is(err, entity.ErrUserWeakPassword):
		return "Password must be at least 12 characters. Choose a longer password and try again."
	case errors.Is(err, entity.ErrUserInvalidEmail):
		return "That email address isn't valid. Enter a valid address like name@example.com."
	case errors.Is(err, entity.ErrUserInvalidTimezone):
		return "That timezone isn't recognised. Pick an IANA timezone such as Europe/Berlin."
	case errors.Is(err, entity.ErrWorkspaceSlugReserved):
		return "That workspace name is reserved. Choose a different name."
	default:
		return "One or more fields are invalid. Check your input and try again."
	}
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
