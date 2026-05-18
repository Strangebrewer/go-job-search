package demo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/Strangebrewer/go-job-search/job"
	"github.com/Strangebrewer/go-job-search/recruiter"
)

type Handler struct {
	recruiterStore *recruiter.Store
	jobStore       *job.Store
}

func NewHandler(recruiterStore *recruiter.Store, jobStore *job.Store) *Handler {
	return &Handler{recruiterStore: recruiterStore, jobStore: jobStore}
}

func (h *Handler) HandleDemoRegistered(w http.ResponseWriter, r *http.Request) {
	var msg pubSubMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		slog.Error("demo-registered: decode body", "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	raw, err := base64.StdEncoding.DecodeString(msg.Message.Data)
	if err != nil {
		slog.Error("demo-registered: decode base64", "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	var payload demoRegisteredPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		slog.Error("demo-registered: unmarshal payload", "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		slog.Error("demo-registered: parse userID", "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx := r.Context()
	recruiterIDs, err := h.seedRecruiters(ctx, userID, payload.ExpiresAt)
	if err != nil {
		slog.Error("demo-registered: seed recruiters", "userId", userID, "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := h.seedJobs(ctx, userID, recruiterIDs, payload.ExpiresAt); err != nil {
		slog.Error("demo-registered: seed jobs", "userId", userID, "error", err)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) seedRecruiters(ctx context.Context, userID uuid.UUID, expiresAt time.Time) ([]uuid.UUID, error) {
	type seed struct {
		name, company, phone, email string
	}
	seeds := []seed{
		{"Sarah Chen", "Apex Talent", "555-0101", "sarah.chen@apextalent.io"},
		{"Marcus Webb", "NovaBridge Recruiting", "555-0202", "marcus.webb@novabridge.com"},
	}

	ids := make([]uuid.UUID, 0, len(seeds))
	for _, s := range seeds {
		req := recruiter.CreateRecruiterRequest{
			Name:    s.name,
			Company: s.company,
			Phone:   s.phone,
			Email:   s.email,
		}
		created, err := h.recruiterStore.Create(ctx, userID, req, &expiresAt)
		if err != nil {
			return nil, err
		}
		id, err := uuid.Parse(created.ID)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (h *Handler) seedJobs(ctx context.Context, userID uuid.UUID, recruiterIDs []uuid.UUID, expiresAt time.Time) error {
	type seed struct {
		title, company string
		recruiterIdx   int
	}
	seeds := []seed{
		{"Senior Software Engineer", "Helios Systems", 0},
		{"Backend Developer", "Crestline Tech", 0},
		{"Platform Engineer", "Vortex Labs", 0},
		{"Software Engineer II", "Meridian Software", 1},
		{"Full Stack Developer", "Pinewave Digital", 1},
		{"API Developer", "Stratum IO", 1},
	}

	for _, s := range seeds {
		req := job.CreateJobRequest{
			RecruiterID: recruiterIDs[s.recruiterIdx].String(),
			JobTitle:    s.title,
			CompanyName: s.company,
		}
		if _, err := h.jobStore.Create(ctx, userID, req, &expiresAt); err != nil {
			return err
		}
	}
	return nil
}
