package recruiter_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcmongo "github.com/testcontainers/testcontainers-go/modules/mongodb"

	"github.com/Strangebrewer/go-job-search/db_connection"
	"github.com/Strangebrewer/go-job-search/job"
	"github.com/Strangebrewer/go-job-search/recruiter"
)

var (
	testStore    *recruiter.Store
	testJobStore *job.Store
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	mongoContainer, err := tcmongo.Run(ctx, "mongo:6")
	if err != nil {
		log.Fatalf("failed to start mongo container: %v", err)
	}
	defer func() {
		if err := mongoContainer.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %v", err)
		}
	}()

	uri, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	_, db, err := db_connection.Connect(ctx, uri)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	testStore = recruiter.NewStore(db)
	testJobStore = job.NewStore(db)

	os.Exit(m.Run())
}

func mustParseUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	require.NoError(t, err)
	return id
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
	assert.Equal(t, userID.String(), r.UserID)
	assert.Equal(t, "Jane Recruiter", r.Name)
	assert.Equal(t, "jane@acme.com", r.Email)
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
		assert.Equal(t, userID.String(), r.UserID)
	}
}

func TestRecruiterStore_GetByID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	created, err := testStore.Create(ctx, userID, recruiter.CreateRecruiterRequest{Name: "Get Me"})
	require.NoError(t, err)

	found, err := testStore.GetByID(ctx, mustParseUUID(t, created.ID), userID)

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

	_, err = testStore.GetByID(ctx, mustParseUUID(t, created.ID), uuid.New())

	assert.ErrorIs(t, err, recruiter.ErrNotFound)
}

func TestRecruiterStore_Update(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	created, err := testStore.Create(ctx, userID, recruiter.CreateRecruiterRequest{Name: "Before"})
	require.NoError(t, err)

	updated, err := testStore.Update(ctx, mustParseUUID(t, created.ID), userID, recruiter.UpdateRecruiterRequest{
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

	err = testStore.Delete(ctx, mustParseUUID(t, created.ID), userID)
	require.NoError(t, err)

	_, err = testStore.GetByID(ctx, mustParseUUID(t, created.ID), userID)
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

	_, err = testJobStore.Create(ctx, userID, job.CreateJobRequest{
		RecruiterID: r.ID,
		JobTitle:    "Engineer",
		CompanyName: "Test Co",
	})
	require.NoError(t, err)

	err = testStore.Delete(ctx, mustParseUUID(t, r.ID), userID)
	assert.ErrorIs(t, err, recruiter.ErrHasJobs)
}
