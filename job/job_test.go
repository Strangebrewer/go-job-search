package job_test

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
	testStore       *job.Store
	seedRecruiterID string
	seedUserID      uuid.UUID
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

	testStore = job.NewStore(db)

	seedUserID = uuid.New()
	recruiterStore := recruiter.NewStore(db)
	r, err := recruiterStore.Create(ctx, seedUserID, recruiter.CreateRecruiterRequest{
		Name:    "Seed Recruiter",
		Company: "Seed Co",
	})
	if err != nil {
		log.Fatalf("failed to create seed recruiter: %v", err)
	}
	seedRecruiterID = r.ID

	os.Exit(m.Run())
}

func seedJob(t *testing.T, overrides ...func(*job.CreateJobRequest)) job.CreateJobRequest {
	t.Helper()
	req := job.CreateJobRequest{
		RecruiterID: seedRecruiterID,
		JobTitle:    "Software Engineer",
		CompanyName: "Acme Corp",
		Status:      "applied",
		WorkFrom:    "remote",
		DateApplied: "2026-01-15",
	}
	for _, fn := range overrides {
		fn(&req)
	}
	return req
}

func mustParseUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	require.NoError(t, err)
	return id
}

func TestJobStore_Create(t *testing.T) {
	ctx := context.Background()

	req := seedJob(t)
	j, err := testStore.Create(ctx, seedUserID, req)

	require.NoError(t, err)
	assert.NotEmpty(t, j.ID)
	assert.Equal(t, seedUserID.String(), j.UserID)
	assert.Equal(t, seedRecruiterID, j.RecruiterID)
	assert.Equal(t, "Software Engineer", j.JobTitle)
	assert.Equal(t, "applied", j.Status)
	assert.Equal(t, []string{}, j.Interviews)
	assert.Equal(t, []string{}, j.Comments)
	assert.False(t, j.Archived)
}

func TestJobStore_Create_RequiresRecruiter(t *testing.T) {
	ctx := context.Background()

	req := seedJob(t)
	req.RecruiterID = ""

	_, err := testStore.Create(ctx, seedUserID, req)

	assert.ErrorIs(t, err, job.ErrInvalidRecruiter)
}

func TestJobStore_Create_InvalidRecruiterID(t *testing.T) {
	ctx := context.Background()

	req := seedJob(t)
	req.RecruiterID = uuid.New().String() // valid UUID but doesn't exist in DB

	_, err := testStore.Create(ctx, seedUserID, req)

	assert.ErrorIs(t, err, job.ErrInvalidRecruiter)
}

func TestJobStore_GetByID(t *testing.T) {
	ctx := context.Background()

	created, err := testStore.Create(ctx, seedUserID, seedJob(t))
	require.NoError(t, err)

	found, err := testStore.GetByID(ctx, mustParseUUID(t, created.ID), seedUserID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestJobStore_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()

	_, err := testStore.GetByID(ctx, uuid.New(), seedUserID)

	assert.ErrorIs(t, err, job.ErrNotFound)
}

func TestJobStore_GetByID_WrongUser(t *testing.T) {
	ctx := context.Background()

	created, err := testStore.Create(ctx, seedUserID, seedJob(t))
	require.NoError(t, err)

	_, err = testStore.GetByID(ctx, mustParseUUID(t, created.ID), uuid.New())

	assert.ErrorIs(t, err, job.ErrNotFound)
}

func TestJobStore_List(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	for i := 0; i < 3; i++ {
		_, err := testStore.Create(ctx, userID, job.CreateJobRequest{
			RecruiterID: seedRecruiterID,
			JobTitle:    "Engineer",
			CompanyName: "Co",
		})
		require.NoError(t, err)
	}

	results, err := testStore.List(ctx, userID, job.JobFilter{})

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 3)
	for _, j := range results {
		assert.Equal(t, userID.String(), j.UserID)
	}
}

func TestJobStore_List_FilterByStatus(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	_, err := testStore.Create(ctx, userID, job.CreateJobRequest{
		RecruiterID: seedRecruiterID,
		JobTitle:    "Dev",
		CompanyName: "Co",
		Status:      "interviewing",
	})
	require.NoError(t, err)
	_, err = testStore.Create(ctx, userID, job.CreateJobRequest{
		RecruiterID: seedRecruiterID,
		JobTitle:    "Dev",
		CompanyName: "Co",
		Status:      "applied",
	})
	require.NoError(t, err)

	results, err := testStore.List(ctx, userID, job.JobFilter{Status: "interviewing"})

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "interviewing", results[0].Status)
}

func TestJobStore_List_ExcludesArchivedByDefault(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	active, err := testStore.Create(ctx, userID, job.CreateJobRequest{
		RecruiterID: seedRecruiterID,
		JobTitle:    "Active",
		CompanyName: "Co",
	})
	require.NoError(t, err)

	archived, err := testStore.Create(ctx, userID, job.CreateJobRequest{
		RecruiterID: seedRecruiterID,
		JobTitle:    "Archived",
		CompanyName: "Co",
	})
	require.NoError(t, err)

	_, err = testStore.Update(ctx, mustParseUUID(t, archived.ID), userID, job.UpdateJobRequest{
		RecruiterID: seedRecruiterID,
		JobTitle:    "Archived",
		CompanyName: "Co",
		Archived:    true,
	})
	require.NoError(t, err)

	results, err := testStore.List(ctx, userID, job.JobFilter{})
	require.NoError(t, err)

	ids := make([]string, len(results))
	for i, r := range results {
		ids[i] = r.ID
	}
	assert.Contains(t, ids, active.ID)
	assert.NotContains(t, ids, archived.ID)
}

func TestJobStore_Update(t *testing.T) {
	ctx := context.Background()

	created, err := testStore.Create(ctx, seedUserID, seedJob(t))
	require.NoError(t, err)

	updated, err := testStore.Update(ctx, mustParseUUID(t, created.ID), seedUserID, job.UpdateJobRequest{
		RecruiterID: seedRecruiterID,
		JobTitle:    "Senior Engineer",
		CompanyName: "New Co",
		Status:      "interviewing",
		Interviews:  []string{"phone screen"},
		Comments:    []string{"went well"},
	})

	require.NoError(t, err)
	assert.Equal(t, "Senior Engineer", updated.JobTitle)
	assert.Equal(t, "New Co", updated.CompanyName)
	assert.Equal(t, "interviewing", updated.Status)
	assert.Equal(t, []string{"phone screen"}, updated.Interviews)
	assert.Equal(t, []string{"went well"}, updated.Comments)
}

func TestJobStore_Update_NotFound(t *testing.T) {
	ctx := context.Background()

	_, err := testStore.Update(ctx, uuid.New(), seedUserID, job.UpdateJobRequest{
		RecruiterID: seedRecruiterID,
		JobTitle:    "X",
		CompanyName: "X",
	})

	assert.ErrorIs(t, err, job.ErrNotFound)
}

func TestJobStore_Delete(t *testing.T) {
	ctx := context.Background()

	created, err := testStore.Create(ctx, seedUserID, seedJob(t))
	require.NoError(t, err)

	err = testStore.Delete(ctx, mustParseUUID(t, created.ID), seedUserID)
	require.NoError(t, err)

	_, err = testStore.GetByID(ctx, mustParseUUID(t, created.ID), seedUserID)
	assert.ErrorIs(t, err, job.ErrNotFound)
}

func TestJobStore_Delete_NotFound(t *testing.T) {
	ctx := context.Background()

	err := testStore.Delete(ctx, uuid.New(), seedUserID)

	assert.ErrorIs(t, err, job.ErrNotFound)
}
