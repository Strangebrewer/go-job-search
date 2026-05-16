package rube

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Strangebrewer/go-job-search/pubsub"
	"github.com/Strangebrewer/go-job-search/tracer"
	"github.com/google/uuid"
)

type ChainRequest struct {
	UserId string `json:"userId"`
	Title  string `json:"title"`
	Link   string `json:"link"`
}

type ChainResponse struct {
	Link  string `json:"link"`
	Title string `json:"title"`
}

type Handler struct {
	publisher       *pubsub.Publisher
	tracer          *tracer.Client
	rubeOwidTopicId string
}

func NewHandler(tc *tracer.Client, publisher *pubsub.Publisher, rubeOwidTopicId string) *Handler {
	return &Handler{
		publisher:       publisher,
		tracer:          tc,
		rubeOwidTopicId: rubeOwidTopicId,
	}
}

func (h *Handler) Chain(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req ChainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	traceID := r.Header.Get("X-Trace-ID")

	if req.Link == "" || req.Title == "" {
		slog.Error("rube missing link or title")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	link := req.Link
	title := req.Title

	if h.publisher != nil && h.rubeOwidTopicId != "" {
		h.publisher.Publish(h.rubeOwidTopicId, pubsub.RubeOwidEventPayload{
			UserId:  req.UserId,
			Link:    link,
			Title:   title,
			TraceID: traceID,
		})
	}

	if h.tracer != nil && traceID != "" {
		h.tracer.Send(tracer.Span{
			TraceID:   traceID,
			SpanID:    uuid.NewString(),
			Service:   "go-job-search",
			Operation: "POST /rube",
			Status:    "ok",
			StartTime: start,
			EndTime:   time.Now(),
			Metadata:  map[string]any{"link": link, "title": title},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ChainResponse{
		Link:  link,
		Title: title,
	})
}
