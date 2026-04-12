package job

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/Strangebrewer/go-job-search/db/generated"
)

var (
	ErrNotFound        = errors.New("job not found")
	ErrInvalidRecruiter = errors.New("invalid recruiter_id")
)

type Store struct {
	q *db.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: db.New(pool)}
}

func (s *Store) Create(ctx context.Context, userID uuid.UUID, req CreateJobRequest) (db.Job, error) {
	recruiterID, err := uuid.Parse(req.RecruiterID)
	if err != nil {
		return db.Job{}, ErrInvalidRecruiter
	}

	id, err := uuid.NewV7()
	if err != nil {
		return db.Job{}, fmt.Errorf("generate id: %w", err)
	}

	status := req.Status
	if status == "" {
		status = "applied"
	}

	now := time.Now().UTC()
	j, err := s.q.CreateJob(ctx, db.CreateJobParams{
		ID:                id,
		UserID:            userID,
		RecruiterID:       recruiterID,
		JobTitle:          req.JobTitle,
		WorkFrom:          req.WorkFrom,
		DateApplied:       req.DateApplied,
		CompanyName:       req.CompanyName,
		CompanyAddress:    req.CompanyAddress,
		CompanyCity:       req.CompanyCity,
		CompanyState:      req.CompanyState,
		PointOfContact:    req.PointOfContact,
		PocTitle:          req.PocTitle,
		Interviews:        []string{},
		Comments:          []string{},
		Status:            status,
		Archived:          false,
		PrimaryLink:       req.PrimaryLink,
		PrimaryLinkText:   req.PrimaryLinkText,
		SecondaryLink:     req.SecondaryLink,
		SecondaryLinkText: req.SecondaryLinkText,
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return db.Job{}, ErrInvalidRecruiter
		}
		return db.Job{}, fmt.Errorf("create job: %w", err)
	}

	return j, nil
}

func (s *Store) GetByID(ctx context.Context, id, userID uuid.UUID) (db.Job, error) {
	j, err := s.q.GetJobByID(ctx, db.GetJobByIDParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Job{}, ErrNotFound
		}
		return db.Job{}, fmt.Errorf("get job: %w", err)
	}
	return j, nil
}

func (s *Store) List(ctx context.Context, userID uuid.UUID, f JobFilter) ([]db.Job, error) {
	params := db.ListJobsParams{
		UserID:          userID,
		IncludeArchived: f.IncludeArchived,
		IncludeDeclined: f.IncludeDeclined,
	}

	if f.Company != "" {
		params.Company = pgtype.Text{String: f.Company, Valid: true}
	}
	if f.RecruiterID != "" {
		rid, err := uuid.Parse(f.RecruiterID)
		if err != nil {
			return nil, ErrInvalidRecruiter
		}
		params.RecruiterID = pgtype.UUID{Bytes: [16]byte(rid), Valid: true}
	}
	if f.Status != "" {
		params.Status = pgtype.Text{String: f.Status, Valid: true}
	}
	if f.WorkFrom != "" {
		params.WorkFrom = pgtype.Text{String: f.WorkFrom, Valid: true}
	}
	if f.DateMin != "" {
		params.DateMin = pgtype.Text{String: f.DateMin, Valid: true}
	}
	if f.DateMax != "" {
		params.DateMax = pgtype.Text{String: f.DateMax, Valid: true}
	}

	jobs, err := s.q.ListJobs(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	if jobs == nil {
		return []db.Job{}, nil
	}

	sortFields := map[string]func(a, b db.Job) bool{
		"jobTitle":    func(a, b db.Job) bool { return a.JobTitle < b.JobTitle },
		"companyName": func(a, b db.Job) bool { return a.CompanyName < b.CompanyName },
		"dateApplied": func(a, b db.Job) bool { return a.DateApplied < b.DateApplied },
		"workFrom":    func(a, b db.Job) bool { return a.WorkFrom < b.WorkFrom },
		"status":      func(a, b db.Job) bool { return a.Status < b.Status },
	}
	if less, ok := sortFields[f.SortBy]; ok {
		sort.Slice(jobs, func(i, j int) bool {
			if f.SortDir == "desc" {
				return less(jobs[j], jobs[i])
			}
			return less(jobs[i], jobs[j])
		})
	}

	return jobs, nil
}

func (s *Store) Update(ctx context.Context, id, userID uuid.UUID, req UpdateJobRequest) (db.Job, error) {
	recruiterID, err := uuid.Parse(req.RecruiterID)
	if err != nil {
		return db.Job{}, ErrInvalidRecruiter
	}

	interviews := req.Interviews
	if interviews == nil {
		interviews = []string{}
	}
	comments := req.Comments
	if comments == nil {
		comments = []string{}
	}

	j, err := s.q.UpdateJob(ctx, db.UpdateJobParams{
		ID:                id,
		UserID:            userID,
		RecruiterID:       recruiterID,
		JobTitle:          req.JobTitle,
		WorkFrom:          req.WorkFrom,
		DateApplied:       req.DateApplied,
		CompanyName:       req.CompanyName,
		CompanyAddress:    req.CompanyAddress,
		CompanyCity:       req.CompanyCity,
		CompanyState:      req.CompanyState,
		PointOfContact:    req.PointOfContact,
		PocTitle:          req.PocTitle,
		Interviews:        interviews,
		Comments:          comments,
		Status:            req.Status,
		Archived:          req.Archived,
		PrimaryLink:       req.PrimaryLink,
		PrimaryLinkText:   req.PrimaryLinkText,
		SecondaryLink:     req.SecondaryLink,
		SecondaryLinkText: req.SecondaryLinkText,
		UpdatedAt:         time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Job{}, ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return db.Job{}, ErrInvalidRecruiter
		}
		return db.Job{}, fmt.Errorf("update job: %w", err)
	}
	return j, nil
}

func (s *Store) Delete(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := s.q.DeleteJob(ctx, db.DeleteJobParams{ID: id, UserID: userID})
	if err != nil {
		return fmt.Errorf("delete job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
