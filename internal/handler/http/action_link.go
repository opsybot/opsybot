package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/service"
)

type actionLinkRoutes struct {
	actions service.Actions
}

func (h *actionLinkRoutes) prompt(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	actionHeaders(w)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(actionPromptPage(token)))
}

func (h *actionLinkRoutes) act(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	outcome, err := h.actions.Redeem(r.Context(), token, clientIP(r))
	actionHeaders(w)
	if err != nil {
		if errors.Is(err, entity.ErrActionTokenInvalid) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(actionResultPage("This action link is invalid or has already been used.")))
			return
		}
		logger.From(r.Context()).WarnContext(r.Context(), "action redeem failed", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(actionResultPage("Something went wrong performing this action.")))
		return
	}
	verb := "Acknowledged"
	if outcome.Action == entity.ActionKindResolve {
		verb = "Resolved"
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(actionResultPage(verb + " at " + outcome.At.Format("15:04") + " UTC.")))
}

func actionHeaders(w http.ResponseWriter) {
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

func actionPromptPage(token string) string {
	return actionShell(`<p>Confirm this action for your alert.</p>` +
		`<form method="post" action="/v1/act/` + htmlEscape(token) + `"><button type="submit">Confirm</button></form>`)
}

func actionResultPage(message string) string {
	return actionShell(`<p>` + htmlEscape(message) + `</p>`)
}

func actionShell(inner string) string {
	return `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">` +
		`<title>Opsybot</title><style>body{font-family:system-ui,sans-serif;background:#0b0b0f;color:#e7e7ea;display:grid;place-items:center;height:100vh;margin:0}` +
		`.card{max-width:24rem;padding:2rem;text-align:center;line-height:1.6}button{font:inherit;padding:.6rem 1.4rem;border-radius:.5rem;border:0;background:#635bff;color:#fff;cursor:pointer}</style>` +
		`</head><body><div class="card">` + inner + `</div></body></html>`
}
