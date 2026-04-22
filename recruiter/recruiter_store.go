package recruiter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrNotFound = errors.New("recruiter not found")
	ErrHasJobs  = errors.New("recruiter has associated jobs")
)

type recruiterDoc struct {
	ID        string    `bson:"_id"`
	UserID    string    `bson:"userId"`
	Name      string    `bson:"name"`
	Company   string    `bson:"company"`
	Phone     string    `bson:"phone"`
	Email     string    `bson:"email"`
	Rating    int32     `bson:"rating"`
	Comments  []string  `bson:"comments"`
	Archived  bool      `bson:"archived"`
	CreatedAt time.Time `bson:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

func (d recruiterDoc) toDomain() Recruiter {
	comments := d.Comments
	if comments == nil {
		comments = []string{}
	}
	return Recruiter{
		ID:        d.ID,
		UserID:    d.UserID,
		Name:      d.Name,
		Company:   d.Company,
		Phone:     d.Phone,
		Email:     d.Email,
		Rating:    d.Rating,
		Comments:  comments,
		Archived:  d.Archived,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

type Store struct {
	col  *mongo.Collection
	jobs *mongo.Collection
}

func NewStore(db *mongo.Database) *Store {
	return &Store{
		col:  db.Collection("recruiters"),
		jobs: db.Collection("jobs"),
	}
}

func (s *Store) Create(ctx context.Context, userID uuid.UUID, req CreateRecruiterRequest) (Recruiter, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Recruiter{}, fmt.Errorf("generate id: %w", err)
	}

	now := time.Now().UTC()
	doc := recruiterDoc{
		ID:        id.String(),
		UserID:    userID.String(),
		Name:      req.Name,
		Company:   req.Company,
		Phone:     req.Phone,
		Email:     strings.ToLower(strings.TrimSpace(req.Email)),
		Rating:    req.Rating,
		Comments:  []string{},
		Archived:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if _, err := s.col.InsertOne(ctx, doc); err != nil {
		return Recruiter{}, fmt.Errorf("create recruiter: %w", err)
	}

	return doc.toDomain(), nil
}

func (s *Store) GetByID(ctx context.Context, id, userID uuid.UUID) (Recruiter, error) {
	var doc recruiterDoc
	err := s.col.FindOne(ctx, bson.D{
		{Key: "_id", Value: id.String()},
		{Key: "userId", Value: userID.String()},
	}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Recruiter{}, ErrNotFound
		}
		return Recruiter{}, fmt.Errorf("get recruiter: %w", err)
	}
	return doc.toDomain(), nil
}

func (s *Store) List(ctx context.Context, userID uuid.UUID) ([]Recruiter, error) {
	cursor, err := s.col.Find(ctx,
		bson.D{{Key: "userId", Value: userID.String()}},
		options.Find().SetSort(bson.D{{Key: "name", Value: 1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("list recruiters: %w", err)
	}

	var docs []recruiterDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode recruiters: %w", err)
	}

	recruiters := make([]Recruiter, len(docs))
	for i, d := range docs {
		recruiters[i] = d.toDomain()
	}
	return recruiters, nil
}

func (s *Store) Update(ctx context.Context, id, userID uuid.UUID, req UpdateRecruiterRequest) (Recruiter, error) {
	comments := req.Comments
	if comments == nil {
		comments = []string{}
	}

	filter := bson.D{
		{Key: "_id", Value: id.String()},
		{Key: "userId", Value: userID.String()},
	}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "name", Value: req.Name},
		{Key: "company", Value: req.Company},
		{Key: "phone", Value: req.Phone},
		{Key: "email", Value: strings.ToLower(strings.TrimSpace(req.Email))},
		{Key: "rating", Value: req.Rating},
		{Key: "comments", Value: comments},
		{Key: "archived", Value: req.Archived},
		{Key: "updatedAt", Value: time.Now().UTC()},
	}}}

	var doc recruiterDoc
	err := s.col.FindOneAndUpdate(ctx, filter, update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Recruiter{}, ErrNotFound
		}
		return Recruiter{}, fmt.Errorf("update recruiter: %w", err)
	}

	return doc.toDomain(), nil
}

func (s *Store) Delete(ctx context.Context, id, userID uuid.UUID) error {
	count, err := s.jobs.CountDocuments(ctx, bson.D{{Key: "recruiterId", Value: id.String()}})
	if err != nil {
		return fmt.Errorf("check jobs: %w", err)
	}
	if count > 0 {
		return ErrHasJobs
	}

	result, err := s.col.DeleteOne(ctx, bson.D{
		{Key: "_id", Value: id.String()},
		{Key: "userId", Value: userID.String()},
	})
	if err != nil {
		return fmt.Errorf("delete recruiter: %w", err)
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}
