package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/service"
)

type channelVerifyRoutes struct {
	channels service.Channels
}

func (h *channelVerifyRoutes) confirm(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	err := h.channels.CompleteByToken(r.Context(), token)
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err != nil {
		if errors.Is(err, entity.ErrChannelVerifyInvalid) || errors.Is(err, entity.ErrChannelVerifyExpired) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(verifyPage("This confirmation link is invalid or has expired.")))
			return
		}
		logger.From(r.Context()).WarnContext(r.Context(), "channel verify failed", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(verifyPage("Something went wrong confirming this channel.")))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(verifyPage("Channel confirmed. You can close this tab.")))
}

func verifyPage(message string) string {
	return `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">` +
		`<title>Opsybot</title><style>body{font-family:system-ui,sans-serif;background:#0b0b0f;color:#e7e7ea;display:grid;place-items:center;height:100vh;margin:0}` +
		`.card{max-width:24rem;padding:2rem;text-align:center;line-height:1.5}</style></head><body><div class="card"><p>` +
		htmlEscape(message) + `</p></div></body></html>`
}

func htmlEscape(s string) string {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		switch r {
		case '<':
			out = append(out, "&lt;"...)
		case '>':
			out = append(out, "&gt;"...)
		case '&':
			out = append(out, "&amp;"...)
		default:
			out = append(out, string(r)...)
		}
	}
	return string(out)
}
