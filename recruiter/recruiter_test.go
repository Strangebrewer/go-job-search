package recruiter_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Strangebrewer/go-job-search/recruiter"
)

var (
	testPool  *pgxpool.Pool
	testStore *recruiter.Store
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		log.Fatalf("failed to start postgres container: %v", err)
	}
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %v", err)
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to create pool: %v", err)
	}
	defer testPool.Close()

	schema, err := os.ReadFile("../db/schema.sql")
	if err != nil {
		log.Fatalf("failed to read schema: %v", err)
	}
	if _, err := testPool.Exec(ctx, string(schema)); err != nil {
		log.Fatalf("failed to apply schema: %v", err)
	}

	testStore = recruiter.NewStore(testPool)

	os.Exit(m.Run())
}

func TestRecruiterStore_Create(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	req := recruiter.CreateRecruiterRequest{
		Name:    "Jane Recruiter",
		Company: "Acme Corp",
		Phone:   "555-0100",
		Email:   "Jane@ACME.COM",
		Rating:  4,
	}

	r, err := testStore.Create(ctx, userID, req)

	require.NoError(t, err)
	assert.NotEmpty(t, r.ID)
	assert.Equal(t, userID, r.UserID)
	assert.Equal(t, "Jane Recruiter", r.Name)
	assert.Equal(t, "jane@acme.com", r.Email) // normalized to lowercase
	assert.Equal(t, int32(4), r.Rating)
	assert.Equal(t, []string{}, r.Comments)
	assert.False(t, r.Archived)
}

func TestRecruiterStore_List(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	otherUserID := uuid.New()

	_, err := testStore.Create(ctx, userID, recruiter.CreateRecruiterRequest{Name: "Alpha"})
	require.NoError(t, err)
	_, err = testStore.Create(ctx, userID, recruiter.CreateRecruiterRequest{Name: "Beta"})
	require.NoError(t, err)
	_, err = testStore.Create(ctx, otherUserID, recruiter.CreateRecruiterRequest{Name: "Other"})
	require.NoError(t, err)

	results, err := testStore.List(ctx, userID)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2)
	for _, r := range results {
		assert.Equal(t, userID, r.UserID)
	}
}

func TestRecruiterStore_GetByID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	created, err := testStore.Create(ctx, userID, recruiter.CreateRecruiterRequest{Name: "Get Me"})
	require.NoError(t, err)

	found, err := testStore.GetByID(ctx, created.ID, userID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Get Me", found.Name)
}

func TestRecruiterStore_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()

	_, err := testStore.GetByID(ctx, uuid.New(), uuid.New())

	assert.ErrorIs(t, err, recruiter.ErrNotFound)
}

func TestRecruiterStore_GetByID_WrongUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	created, err := testStore.Create(ctx, userID, recruiter.CreateRecruiterRequest{Name: "Wrong User"})
	require.NoError(t, err)

	_, err = testStore.GetByID(ctx, created.ID, uuid.New())

	assert.ErrorIs(t, err, recruiter.ErrNotFound)
}

func TestRecruiterStore_Update(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	created, err := testStore.Create(ctx, userID, recruiter.CreateRecruiterRequest{Name: "Before"})
	require.NoError(t, err)

	updated, err := testStore.Update(ctx, created.ID, userID, recruiter.UpdateRecruiterRequest{
		Name:     "After",
		Company:  "New Co",
		Rating:   5,
		Comments: []string{"great"},
		Archived: false,
	})

	require.NoError(t, err)
	assert.Equal(t, "After", updated.Name)
	assert.Equal(t, "New Co", updated.Company)
	assert.Equal(t, int32(5), updated.Rating)
	assert.Equal(t, []string{"great"}, updated.Comments)
}

func TestRecruiterStore_Delete(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	created, err := testStore.Create(ctx, userID, recruiter.CreateRecruiterRequest{Name: "Delete Me"})
	require.NoError(t, err)

	err = testStore.Delete(ctx, created.ID, userID)
	require.NoError(t, err)

	_, err = testStore.GetByID(ctx, created.ID, userID)
	assert.ErrorIs(t, err, recruiter.ErrNotFound)
}

func TestRecruiterStore_Delete_NotFound(t *testing.T) {
	ctx := context.Background()

	err := testStore.Delete(ctx, uuid.New(), uuid.New())

	assert.ErrorIs(t, err, recruiter.ErrNotFound)
}

func TestRecruiterStore_Delete_BlockedByJobs(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	r, err := testStore.Create(ctx, userID, recruiter.CreateRecruiterRequest{Name: "Has Jobs"})
	require.NoError(t, err)

	// Insert a job referencing this recruiter directly via SQL to set up the FK constraint
	jobID := uuid.New()
	_, err = testPool.Exec(ctx,
		`INSERT INTO jobs (id, user_id, recruiter_id, job_title, company_name, created_at, updated_at)
		 VALUES ($1, $2, $3, 'Engineer', 'Test Co', now(), now())`,
		jobID, userID, r.ID,
	)
	require.NoError(t, err)

	err = testStore.Delete(ctx, r.ID, userID)
	assert.ErrorIs(t, err, recruiter.ErrHasJobs)
}
