package http

import (
	"io"
	"net/http"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/service"
)

const maxCallbackBody = 1 << 20

type chatCallbackRoutes struct {
	interactions service.Interactions
}

func (h *chatCallbackRoutes) slack(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxCallbackBody))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	resp, err := h.interactions.Slack(r.Context(), entity.ChatCallback{
		Provider:  entity.ChatProviderSlack,
		Body:      body,
		Signature: r.Header.Get("X-Slack-Signature"),
		Timestamp: r.Header.Get("X-Slack-Request-Timestamp"),
		IP:        clientIP(r),
	})
	writeInteraction(w, resp, err)
}

func (h *chatCallbackRoutes) discord(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxCallbackBody))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	resp, err := h.interactions.Discord(r.Context(), entity.ChatCallback{
		Provider:  entity.ChatProviderDiscord,
		Body:      body,
		Signature: r.Header.Get("X-Signature-Ed25519"),
		Timestamp: r.Header.Get("X-Signature-Timestamp"),
		IP:        clientIP(r),
	})
	writeInteraction(w, resp, err)
}

func writeInteraction(w http.ResponseWriter, resp entity.InteractionResponse, err error) {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if resp.Status == 0 {
		resp.Status = http.StatusOK
	}
	if resp.ContentType != "" {
		w.Header().Set("Content-Type", resp.ContentType)
	}
	w.WriteHeader(resp.Status)
	if len(resp.Body) > 0 {
		_, _ = w.Write(resp.Body)
	}
}

func clientIP(r *http.Request) string {
	if info := entity.RequestInfoFrom(r.Context()); info.IP != "" {
		return info.IP
	}
	return r.RemoteAddr
}
