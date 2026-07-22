package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/service"
)

type ingestRoutes struct {
	ingest  service.Ingest
	sem     chan struct{}
	maxBody int64
}

func newIngestRoutes(ingest service.Ingest, cfg config.Ingest) *ingestRoutes {
	concurrent := cfg.MaxConcurrent
	if concurrent <= 0 {
		concurrent = 1
	}
	maxBody := cfg.MaxBodyBytes
	if maxBody <= 0 {
		maxBody = 1 << 20
	}
	return &ingestRoutes{ingest: ingest, sem: make(chan struct{}, concurrent), maxBody: maxBody}
}

type ingestResponseAlert struct {
	ID       string `json:"id"`
	DedupKey string `json:"dedupKey"`
	Outcome  string `json:"outcome"`
}

type ingestResponse struct {
	Accepted int                   `json:"accepted"`
	Alerts   []ingestResponseAlert `json:"alerts"`
}

func (h *ingestRoutes) webhook(w http.ResponseWriter, r *http.Request) {
	select {
	case h.sem <- struct{}{}:
		defer func() { <-h.sem }()
	default:
		w.Header().Set("Retry-After", "1")
		writeIngestProblem(w, http.StatusTooManyRequests, "Too many requests", "The ingest queue is full. Retry shortly.")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, h.maxBody+1))
	if err != nil {
		writeIngestProblem(w, http.StatusBadRequest, "Unreadable body", "The request body could not be read.")
		return
	}

	results, err := h.ingest.Webhook(r.Context(), entity.IngestRequest{
		Token:       chi.URLParam(r, "token"),
		Signature:   r.Header.Get(entity.SourceSignatureHeader),
		Body:        body,
		ContentType: r.Header.Get("Content-Type"),
		Method:      r.Method,
		RemoteIP:    entity.RequestInfoFrom(r.Context()).IP,
		ReceivedAt:  time.Now().UTC(),
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	out := ingestResponse{Accepted: len(results), Alerts: make([]ingestResponseAlert, 0, len(results))}
	for _, res := range results {
		out.Alerts = append(out.Alerts, ingestResponseAlert{ID: res.AlertID, DedupKey: res.DedupKey, Outcome: string(res.Outcome)})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(out)
}

func (h *ingestRoutes) checkIn(w http.ResponseWriter, r *http.Request) {
	select {
	case h.sem <- struct{}{}:
		defer func() { <-h.sem }()
	default:
		w.Header().Set("Retry-After", "1")
		writeIngestProblem(w, http.StatusTooManyRequests, "Too many requests", "The ingest queue is full. Retry shortly.")
		return
	}

	result, err := h.ingest.CheckIn(r.Context(), entity.CheckInRequest{
		Token:      chi.URLParam(r, "token"),
		RemoteIP:   entity.RequestInfoFrom(r.Context()).IP,
		ReceivedAt: time.Now().UTC(),
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(ingestResponse{
		Accepted: 1,
		Alerts:   []ingestResponseAlert{{ID: result.AlertID, DedupKey: result.DedupKey, Outcome: string(result.Outcome)}},
	})
}

func (h *ingestRoutes) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, entity.ErrAlertSourceNotFound), errors.Is(err, entity.ErrAlertMonitorNotFound):
		writeIngestProblem(w, http.StatusNotFound, "Unknown endpoint", "That ingest URL does not match any source.")
	case errors.Is(err, entity.ErrAlertMonitorFormat):
		writeIngestProblem(w, http.StatusConflict, "Not a monitor", "That URL belongs to a webhook source, not a heartbeat monitor.")
	case errors.Is(err, entity.ErrAlertSourcePaused):
		writeIngestProblem(w, http.StatusConflict, "Source paused", "This source is paused, so events are not accepted.")
	case errors.Is(err, entity.ErrAlertSourceSignature):
		writeIngestProblem(w, http.StatusUnauthorized, "Signature rejected", "The request signature did not match the source secret.")
	case errors.Is(err, entity.ErrIngestFlooded):
		w.Header().Set("Retry-After", "60")
		writeIngestProblem(w, http.StatusTooManyRequests, "Too many events", "This source is over its ingest budget.")
	case errors.Is(err, entity.ErrIngestUnparseable):
		detail := "The payload could not be read. It is recorded on the failures page."
		if pe, ok := entity.ParseFailureOf(err); ok {
			detail = pe.Detail
			if pe.Reason == entity.FailureBodyTooLarge {
				writeIngestProblem(w, http.StatusRequestEntityTooLarge, "Payload too large", detail)
				return
			}
		}
		writeIngestProblem(w, http.StatusBadRequest, "Payload rejected", detail)
	default:
		logger.From(r.Context()).ErrorContext(r.Context(), "alert ingestion failed", "error", err)
		writeIngestProblem(w, http.StatusInternalServerError, "Ingestion failed", "The event could not be processed.")
	}
}

func writeIngestProblem(w http.ResponseWriter, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":   "about:blank",
		"title":  title,
		"status": status,
		"detail": detail,
	})
}
