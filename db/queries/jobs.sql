-- name: CreateJob :one
INSERT INTO jobs (
    id, user_id, recruiter_id, job_title, work_from, date_applied,
    company_name, company_address, company_city, company_state,
    point_of_contact, poc_title, interviews, comments, status, archived,
    primary_link, primary_link_text, secondary_link, secondary_link_text,
    created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
RETURNING *;

-- name: GetJobByID :one
SELECT * FROM jobs WHERE id = $1 AND user_id = $2;

-- name: ListJobs :many
SELECT * FROM jobs
WHERE user_id = sqlc.arg('user_id')
  AND (sqlc.narg('company')::text IS NULL OR company_name ILIKE '%' || sqlc.narg('company') || '%')
  AND (sqlc.narg('recruiter_id')::uuid IS NULL OR recruiter_id = sqlc.narg('recruiter_id'))
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('work_from')::text IS NULL OR work_from ILIKE '%' || sqlc.narg('work_from') || '%')
  AND (sqlc.narg('date_min')::text IS NULL OR date_applied >= sqlc.narg('date_min'))
  AND (sqlc.narg('date_max')::text IS NULL OR date_applied <= sqlc.narg('date_max'))
  AND (sqlc.arg('include_archived')::boolean = true OR archived = false)
  AND (sqlc.arg('include_declined')::boolean = true OR status != 'declined')
ORDER BY created_at DESC;

-- name: UpdateJob :one
UPDATE jobs
SET recruiter_id = $3, job_title = $4, work_from = $5, date_applied = $6,
    company_name = $7, company_address = $8, company_city = $9, company_state = $10,
    point_of_contact = $11, poc_title = $12, interviews = $13, comments = $14,
    status = $15, archived = $16, primary_link = $17, primary_link_text = $18,
    secondary_link = $19, secondary_link_text = $20, updated_at = $21
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteJob :execresult
DELETE FROM jobs WHERE id = $1 AND user_id = $2;
