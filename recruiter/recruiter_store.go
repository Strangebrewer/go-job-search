package recruiter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/Strangebrewer/go-job-search/db/generated"
)

var (
	ErrNotFound = errors.New("recruiter not found")
	ErrHasJobs  = errors.New("recruiter has associated jobs")
)

type Store struct {
	q *db.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: db.New(pool)}
}

func (s *Store) Create(ctx context.Context, userID uuid.UUID, req CreateRecruiterRequest) (db.Recruiter, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return db.Recruiter{}, fmt.Errorf("generate id: %w", err)
	}

	now := time.Now().UTC()
	r, err := s.q.CreateRecruiter(ctx, db.CreateRecruiterParams{
		ID:        id,
		UserID:    userID,
		Name:      req.Name,
		Company:   req.Company,
		Phone:     req.Phone,
		Email:     strings.ToLower(strings.TrimSpace(req.Email)),
		Rating:    req.Rating,
		Comments:  []string{},
		Archived:  false,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return db.Recruiter{}, fmt.Errorf("create recruiter: %w", err)
	}

	return r, nil
}

func (s *Store) GetByID(ctx context.Context, id, userID uuid.UUID) (db.Recruiter, error) {
	r, err := s.q.GetRecruiterByID(ctx, db.GetRecruiterByIDParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Recruiter{}, ErrNotFound
		}
		return db.Recruiter{}, fmt.Errorf("get recruiter: %w", err)
	}
	return r, nil
}

func (s *Store) List(ctx context.Context, userID uuid.UUID) ([]db.Recruiter, error) {
	recruiters, err := s.q.ListRecruiters(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list recruiters: %w", err)
	}
	if recruiters == nil {
		return []db.Recruiter{}, nil
	}
	return recruiters, nil
}

func (s *Store) Update(ctx context.Context, id, userID uuid.UUID, req UpdateRecruiterRequest) (db.Recruiter, error) {
	comments := req.Comments
	if comments == nil {
		comments = []string{}
	}

	r, err := s.q.UpdateRecruiter(ctx, db.UpdateRecruiterParams{
		ID:        id,
		UserID:    userID,
		Name:      req.Name,
		Company:   req.Company,
		Phone:     req.Phone,
		Email:     strings.ToLower(strings.TrimSpace(req.Email)),
		Rating:    req.Rating,
		Comments:  comments,
		Archived:  req.Archived,
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Recruiter{}, ErrNotFound
		}
		return db.Recruiter{}, fmt.Errorf("update recruiter: %w", err)
	}
	return r, nil
}

func (s *Store) Delete(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := s.q.DeleteRecruiter(ctx, db.DeleteRecruiterParams{ID: id, UserID: userID})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrHasJobs
		}
		return fmt.Errorf("delete recruiter: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
