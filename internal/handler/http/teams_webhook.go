package http

import (
	"io"
	"net/http"
)

const maxTeamsBody = 1 << 20

type teamsWebhookRoutes struct{}

func (h *teamsWebhookRoutes) handle(w http.ResponseWriter, r *http.Request) {
	_, _ = io.Copy(io.Discard, io.LimitReader(r.Body, maxTeamsBody))
	w.WriteHeader(http.StatusOK)
}
