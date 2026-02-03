-- name: AddToWaitlist :one
INSERT INTO waitlists (
  email,
  source
) VALUES ($1, $2) RETURNING *;


-- name: GetWaitlistUserByEmail :one
select *
from waitlists
where email = $1
limit 1
;


-- name: ListWaitlistUsersByStatus :many
select *
from waitlists
where status = $1
;

