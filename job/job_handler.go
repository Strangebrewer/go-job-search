package job

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Strangebrewer/go-job-search/middleware"
	"github.com/Strangebrewer/go-job-search/tracer"
)

type Handler struct {
	store  *Store
	tracer *tracer.Client
}

func NewHandler(store *Store, tc *tracer.Client) *Handler {
	return &Handler{store: store, tracer: tc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	traceID := r.Header.Get("X-Trace-ID")

	f := JobFilter{
		Company:         r.URL.Query().Get("company"),
		RecruiterID:     r.URL.Query().Get("recruiter"),
		Status:          r.URL.Query().Get("status"),
		WorkFrom:        r.URL.Query().Get("workFrom"),
		DateMin:         r.URL.Query().Get("dateMin"),
		DateMax:         r.URL.Query().Get("dateMax"),
		IncludeArchived: r.URL.Query().Get("archived") == "true",
		IncludeDeclined: r.URL.Query().Get("includeDeclined") == "true",
		SortBy:          r.URL.Query().Get("sortBy"),
		SortDir:         r.URL.Query().Get("sortDir"),
	}

	start := time.Now()
	jobs, err := h.store.List(r.Context(), userID, f)
	end := time.Now()

	if err != nil {
		if errors.Is(err, ErrInvalidRecruiter) {
			if h.tracer != nil && traceID != "" {
				errMsg := "invalid recruiter id"
				h.tracer.Send(tracer.Span{
					TraceID:   traceID,
					SpanID:    uuid.NewString(),
					Service:   "go-job-search",
					Operation: "list_jobs",
					Status:    "error",
					Error:     &errMsg,
					StartTime: start,
					EndTime:   end,
				})
			}
			http.Error(w, "invalid recruiter id", http.StatusBadRequest)
			return
		}
		slog.Error("list jobs", "error", err)
		if h.tracer != nil && traceID != "" {
			errMsg := "internal server error"
			h.tracer.Send(tracer.Span{
				TraceID:   traceID,
				SpanID:    uuid.NewString(),
				Service:   "go-job-search",
				Operation: "list_jobs",
				Status:    "error",
				Error:     &errMsg,
				StartTime: start,
				EndTime:   end,
			})
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	if h.tracer != nil && traceID != "" {
		h.tracer.Send(tracer.Span{
			TraceID:   traceID,
			SpanID:    uuid.NewString(),
			Service:   "go-job-search",
			Operation: "list_jobs",
			Status:    "ok",
			StartTime: start,
			EndTime:   end,
			Metadata:  map[string]any{"count": len(jobs)},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(jobs)
}

func (h *Handler) GetOne(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	j, err := h.store.GetByID(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("get job", "id", id, "error", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(j)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.RecruiterID == "" {
		http.Error(w, "recruiter_id is required", http.StatusBadRequest)
		return
	}
	if req.JobTitle == "" {
		http.Error(w, "job_title is required", http.StatusBadRequest)
		return
	}
	if req.CompanyName == "" {
		http.Error(w, "company_name is required", http.StatusBadRequest)
		return
	}

	created, err := h.store.Create(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, ErrInvalidRecruiter) {
			http.Error(w, "invalid recruiter_id", http.StatusBadRequest)
			return
		}
		slog.Error("create job", "error", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req UpdateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	updated, err := h.store.Update(r.Context(), id, userID, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrInvalidRecruiter) {
			http.Error(w, "invalid recruiter_id", http.StatusBadRequest)
			return
		}
		slog.Error("update job", "id", id, "error", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(updated)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.store.Delete(r.Context(), id, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("delete job", "id", id, "error", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func userIDFromContext(r *http.Request) (uuid.UUID, error) {
	idStr, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return uuid.UUID{}, errors.New("no user id in context")
	}
	return uuid.Parse(idStr)
}
