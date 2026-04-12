-- name: CreateRecruiter :one
INSERT INTO recruiters (id, user_id, name, company, phone, email, rating, comments, archived, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetRecruiterByID :one
SELECT * FROM recruiters WHERE id = $1 AND user_id = $2;

-- name: ListRecruiters :many
SELECT * FROM recruiters WHERE user_id = $1 ORDER BY name ASC;

-- name: UpdateRecruiter :one
UPDATE recruiters
SET name = $3, company = $4, phone = $5, email = $6, rating = $7, comments = $8, archived = $9, updated_at = $10
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteRecruiter :execresult
DELETE FROM recruiters WHERE id = $1 AND user_id = $2;
