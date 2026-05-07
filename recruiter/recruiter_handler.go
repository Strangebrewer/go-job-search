package recruiter

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

	start := time.Now()
	recruiters, err := h.store.List(r.Context(), userID)
	end := time.Now()
	if err != nil {
		slog.Error("list recruiters", "error", err)
		errMsg := "internal server error"
		h.tracer.SendErrorSpan(traceID, "list_recruiters", errMsg, start, end)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	h.tracer.SendSpan(traceID, "list_recruiters", start, end, len(recruiters))

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(recruiters)
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
	op := fmt.Sprintf("get_recruiter by id: %s", id)

	start := time.Now()
	recruiter, err := h.store.GetByID(r.Context(), id, userID)
	end := time.Now()
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			errMsg := "not found"
			h.tracer.SendErrorSpan(traceID, op, errMsg, start, end)
			http.Error(w, errMsg, http.StatusNotFound)
			return
		}
		slog.Error("get recruiter", "id", id, "error", err)
		errMsg := "internal server error"
		h.tracer.SendErrorSpan(traceID, op, errMsg, start, end)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	h.tracer.SendSpan(traceID, op, start, end)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(recruiter)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	traceID := r.Header.Get("X-Trace-ID")

	start := time.Now()
	var req CreateRecruiterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errMsg := "invalid json"
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "create_recruiter", errMsg, start, end)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		errMsg := "name is required"
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "create_recruiter", errMsg, start, end)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	created, err := h.store.Create(r.Context(), userID, req)
	end := time.Now()
	if err != nil {
		slog.Error("create recruiter", "error", err)
		errMsg := "internal server error"
		h.tracer.SendErrorSpan(traceID, "create_recruiter", errMsg, start, end)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	h.tracer.SendSpan(traceID, "create_recruiter", start, end)

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
	op := fmt.Sprintf("update_recruiter by id: %s", id)

	if err != nil {
		errMsg := "invalid id"
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, op, errMsg, start, end)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	var req UpdateRecruiterRequest
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
		slog.Error("update recruiter", "id", id, "error", err)
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
	op := fmt.Sprintf("delete_recruiter by id: %s", id)
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
		if errors.Is(err, ErrHasJobs) {
			errMsg := "recruiter has associated jobs"
			end := time.Now()
			h.tracer.SendErrorSpan(traceID, op, errMsg, start, end)
			http.Error(w, errMsg, http.StatusConflict)
			return
		}
		slog.Error("delete recruiter", "id", id, "error", err)
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
