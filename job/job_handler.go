package job

import (
	"encoding/json"
	"errors"
	"fmt"
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
			errMsg := "invalid recruiter id"
			h.tracer.SendErrorSpan(traceID, "list_jobs", errMsg, start, end)
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}
		slog.Error("list jobs", "error", err)
		errMsg := "internal server error"
		h.tracer.SendErrorSpan(traceID, "list_jobs", errMsg, start, end)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	h.tracer.SendSpan(traceID, "list_jobs", start, end, len(jobs))

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

	traceID := r.Header.Get("X-Trace-ID")
	op := fmt.Sprintf("get_job by id: %s", id)

	start := time.Now()
	j, err := h.store.GetByID(r.Context(), id, userID)
	end := time.Now()
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			errMsg := "not found"
			h.tracer.SendErrorSpan(traceID, op, errMsg, start, end)
			http.Error(w, errMsg, http.StatusNotFound)
			return
		}
		slog.Error("get job", "id", id, "error", err)
		errMsg := "internal server error"
		h.tracer.SendErrorSpan(traceID, op, errMsg, start, end)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	h.tracer.SendSpan(traceID, op, start, end)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(j)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	traceID := r.Header.Get("X-Trace-ID")

	start := time.Now()
	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errMsg := "invalid json"
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "create_job", errMsg, start, end)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if req.RecruiterID == "" {
		errMsg := "recruiter_id is required"
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "create_job", errMsg, start, end)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	if req.JobTitle == "" {
		errMsg := "job_title is required"
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "create_job", errMsg, start, end)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	if req.CompanyName == "" {
		errMsg := "company_name is required"
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "create_job", errMsg, start, end)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	created, err := h.store.Create(r.Context(), userID, req)
	end := time.Now()
	if err != nil {
		if errors.Is(err, ErrInvalidRecruiter) {
			errMsg := "invalid recruiter id"
			h.tracer.SendErrorSpan(traceID, "create_job", errMsg, start, end)
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}
		slog.Error("create job", "error", err)
		errMsg := "internal server error"
		h.tracer.SendErrorSpan(traceID, "create_job", errMsg, start, end)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	h.tracer.SendSpan(traceID, "create_job", start, end)

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

	traceID := r.Header.Get("X-Trace-ID")

	start := time.Now()
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	op := fmt.Sprintf("update_job by id: %s", id)

	if err != nil {
		errMsg := "invalid id"
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, op, errMsg, start, end)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	var req UpdateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errMsg := "invalid json"
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, op, errMsg, start, end)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	updated, err := h.store.Update(r.Context(), id, userID, req)
	end := time.Now()

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			errMsg := "not found"
			h.tracer.SendErrorSpan(traceID, op, errMsg, start, end)
			http.Error(w, errMsg, http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrInvalidRecruiter) {
			errMsg := "invalid recruiter_id"
			h.tracer.SendErrorSpan(traceID, op, errMsg, start, end)
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}
		slog.Error("update job", "id", id, "error", err)
		errMsg := "internal server error"
		h.tracer.SendErrorSpan(traceID, op, errMsg, start, end)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	h.tracer.SendSpan(traceID, op, start, end)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(updated)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	traceID := r.Header.Get("X-Trace-ID")

	start := time.Now()
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	op := fmt.Sprintf("delete_job by id: %s", id)
	if err != nil {
		errMsg := "invalid id"
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, op, errMsg, start, end)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if err := h.store.Delete(r.Context(), id, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			errMsg := "not found"
			end := time.Now()
			h.tracer.SendErrorSpan(traceID, op, errMsg, start, end)
			http.Error(w, errMsg, http.StatusNotFound)
			return
		}
		slog.Error("delete job", "id", id, "error", err)
		errMsg := "internal server error"
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, op, errMsg, start, end)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	end := time.Now()
	h.tracer.SendSpan(traceID, op, start, end)

	w.WriteHeader(http.StatusNoContent)
}

func userIDFromContext(r *http.Request) (uuid.UUID, error) {
	idStr, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return uuid.UUID{}, errors.New("no user id in context")
	}
	return uuid.Parse(idStr)
}
