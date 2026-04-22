package job

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrNotFound         = errors.New("job not found")
	ErrInvalidRecruiter = errors.New("invalid recruiter_id")
)

type jobDoc struct {
	ID                string    `bson:"_id"`
	UserID            string    `bson:"userId"`
	RecruiterID       string    `bson:"recruiterId"`
	JobTitle          string    `bson:"jobTitle"`
	WorkFrom          string    `bson:"workFrom"`
	DateApplied       string    `bson:"dateApplied"`
	CompanyName       string    `bson:"companyName"`
	CompanyAddress    string    `bson:"companyAddress"`
	CompanyCity       string    `bson:"companyCity"`
	CompanyState      string    `bson:"companyState"`
	PointOfContact    string    `bson:"pointOfContact"`
	PocTitle          string    `bson:"pocTitle"`
	Interviews        []string  `bson:"interviews"`
	Comments          []string  `bson:"comments"`
	Status            string    `bson:"status"`
	Archived          bool      `bson:"archived"`
	PrimaryLink       string    `bson:"primaryLink"`
	PrimaryLinkText   string    `bson:"primaryLinkText"`
	SecondaryLink     string    `bson:"secondaryLink"`
	SecondaryLinkText string    `bson:"secondaryLinkText"`
	CreatedAt         time.Time `bson:"createdAt"`
	UpdatedAt         time.Time `bson:"updatedAt"`
}

func (d jobDoc) toDomain() Job {
	interviews := d.Interviews
	if interviews == nil {
		interviews = []string{}
	}
	comments := d.Comments
	if comments == nil {
		comments = []string{}
	}
	return Job{
		ID:                d.ID,
		UserID:            d.UserID,
		RecruiterID:       d.RecruiterID,
		JobTitle:          d.JobTitle,
		WorkFrom:          d.WorkFrom,
		DateApplied:       d.DateApplied,
		CompanyName:       d.CompanyName,
		CompanyAddress:    d.CompanyAddress,
		CompanyCity:       d.CompanyCity,
		CompanyState:      d.CompanyState,
		PointOfContact:    d.PointOfContact,
		PocTitle:          d.PocTitle,
		Interviews:        interviews,
		Comments:          comments,
		Status:            d.Status,
		Archived:          d.Archived,
		PrimaryLink:       d.PrimaryLink,
		PrimaryLinkText:   d.PrimaryLinkText,
		SecondaryLink:     d.SecondaryLink,
		SecondaryLinkText: d.SecondaryLinkText,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}
}

type Store struct {
	col        *mongo.Collection
	recruiters *mongo.Collection
}

func NewStore(db *mongo.Database) *Store {
	return &Store{
		col:        db.Collection("jobs"),
		recruiters: db.Collection("recruiters"),
	}
}

func (s *Store) Create(ctx context.Context, userID uuid.UUID, req CreateJobRequest) (Job, error) {
	if req.RecruiterID == "" {
		return Job{}, ErrInvalidRecruiter
	}
	if _, err := uuid.Parse(req.RecruiterID); err != nil {
		return Job{}, ErrInvalidRecruiter
	}

	count, err := s.recruiters.CountDocuments(ctx, bson.D{
		{Key: "_id", Value: req.RecruiterID},
		{Key: "userId", Value: userID.String()},
	})
	if err != nil {
		return Job{}, fmt.Errorf("validate recruiter: %w", err)
	}
	if count == 0 {
		return Job{}, ErrInvalidRecruiter
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Job{}, fmt.Errorf("generate id: %w", err)
	}

	status := req.Status
	if status == "" {
		status = "applied"
	}

	now := time.Now().UTC()
	doc := jobDoc{
		ID:                id.String(),
		UserID:            userID.String(),
		RecruiterID:       req.RecruiterID,
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
	}

	if _, err := s.col.InsertOne(ctx, doc); err != nil {
		return Job{}, fmt.Errorf("create job: %w", err)
	}

	return doc.toDomain(), nil
}

func (s *Store) GetByID(ctx context.Context, id, userID uuid.UUID) (Job, error) {
	var doc jobDoc
	err := s.col.FindOne(ctx, bson.D{
		{Key: "_id", Value: id.String()},
		{Key: "userId", Value: userID.String()},
	}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Job{}, ErrNotFound
		}
		return Job{}, fmt.Errorf("get job: %w", err)
	}
	return doc.toDomain(), nil
}

func (s *Store) List(ctx context.Context, userID uuid.UUID, f JobFilter) ([]Job, error) {
	filter := bson.D{{Key: "userId", Value: userID.String()}}

	if !f.IncludeArchived {
		filter = append(filter, bson.E{Key: "archived", Value: false})
	}
	if !f.IncludeDeclined {
		filter = append(filter, bson.E{Key: "status", Value: bson.D{{Key: "$ne", Value: "declined"}}})
	}
	if f.Company != "" {
		filter = append(filter, bson.E{Key: "companyName", Value: bson.D{
			{Key: "$regex", Value: f.Company},
			{Key: "$options", Value: "i"},
		}})
	}
	if f.RecruiterID != "" {
		if _, err := uuid.Parse(f.RecruiterID); err != nil {
			return nil, ErrInvalidRecruiter
		}
		filter = append(filter, bson.E{Key: "recruiterId", Value: f.RecruiterID})
	}
	if f.Status != "" {
		filter = append(filter, bson.E{Key: "status", Value: f.Status})
	}
	if f.WorkFrom != "" {
		filter = append(filter, bson.E{Key: "workFrom", Value: f.WorkFrom})
	}
	if f.DateMin != "" {
		filter = append(filter, bson.E{Key: "dateApplied", Value: bson.D{{Key: "$gte", Value: f.DateMin}}})
	}
	if f.DateMax != "" {
		filter = append(filter, bson.E{Key: "dateApplied", Value: bson.D{{Key: "$lte", Value: f.DateMax}}})
	}

	opts := options.Find()
	if f.SortBy != "" {
		validSortFields := map[string]bool{
			"jobTitle": true, "companyName": true, "dateApplied": true, "workFrom": true, "status": true,
		}
		if validSortFields[f.SortBy] {
			dir := 1
			if f.SortDir == "desc" {
				dir = -1
			}
			opts.SetSort(bson.D{{Key: f.SortBy, Value: dir}})
		}
	}

	cursor, err := s.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}

	var docs []jobDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode jobs: %w", err)
	}

	jobs := make([]Job, len(docs))
	for i, d := range docs {
		jobs[i] = d.toDomain()
	}
	return jobs, nil
}

func (s *Store) Update(ctx context.Context, id, userID uuid.UUID, req UpdateJobRequest) (Job, error) {
	if req.RecruiterID == "" {
		return Job{}, ErrInvalidRecruiter
	}
	if _, err := uuid.Parse(req.RecruiterID); err != nil {
		return Job{}, ErrInvalidRecruiter
	}

	interviews := req.Interviews
	if interviews == nil {
		interviews = []string{}
	}
	comments := req.Comments
	if comments == nil {
		comments = []string{}
	}

	filter := bson.D{
		{Key: "_id", Value: id.String()},
		{Key: "userId", Value: userID.String()},
	}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "recruiterId", Value: req.RecruiterID},
		{Key: "jobTitle", Value: req.JobTitle},
		{Key: "workFrom", Value: req.WorkFrom},
		{Key: "dateApplied", Value: req.DateApplied},
		{Key: "companyName", Value: req.CompanyName},
		{Key: "companyAddress", Value: req.CompanyAddress},
		{Key: "companyCity", Value: req.CompanyCity},
		{Key: "companyState", Value: req.CompanyState},
		{Key: "pointOfContact", Value: req.PointOfContact},
		{Key: "pocTitle", Value: req.PocTitle},
		{Key: "interviews", Value: interviews},
		{Key: "comments", Value: comments},
		{Key: "status", Value: req.Status},
		{Key: "archived", Value: req.Archived},
		{Key: "primaryLink", Value: req.PrimaryLink},
		{Key: "primaryLinkText", Value: req.PrimaryLinkText},
		{Key: "secondaryLink", Value: req.SecondaryLink},
		{Key: "secondaryLinkText", Value: req.SecondaryLinkText},
		{Key: "updatedAt", Value: time.Now().UTC()},
	}}}

	var doc jobDoc
	err := s.col.FindOneAndUpdate(ctx, filter, update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Job{}, ErrNotFound
		}
		return Job{}, fmt.Errorf("update job: %w", err)
	}

	return doc.toDomain(), nil
}

func (s *Store) Delete(ctx context.Context, id, userID uuid.UUID) error {
	result, err := s.col.DeleteOne(ctx, bson.D{
		{Key: "_id", Value: id.String()},
		{Key: "userId", Value: userID.String()},
	})
	if err != nil {
		return fmt.Errorf("delete job: %w", err)
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}
