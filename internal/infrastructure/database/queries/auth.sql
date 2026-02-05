-- name: FindUserByEmail :one
select
    id,
    email,
    password_hash,
    full_name,
    phone,
    is_email_verified,
    is_active,
    last_login_at,
    created_at
from users
where email = $1
limit 1
;

-- name: FindUserByPhone :one
select
    id,
    email,
    password_hash,
    full_name,
    phone,
    is_email_verified,
    is_active,
    last_login_at,
    created_at
from users
where phone = $1
limit 1
;

-- name: FindUserById :one
select
    id,
    email,
    password_hash,
    full_name,
    phone,
    is_email_verified,
    is_active,
    last_login_at,
    created_at
from users
where id = $1
limit 1
;

-- name: CreateUserAccount :one
INSERT INTO users(
  email, phone, full_name, password_hash
) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: MarkEmailAsVerified :exec
UPDATE users 
SET is_email_verified = true, is_active = true
WHERE id = $1;

-- name: RecordLastLoginTime :exec
UPDATE users
SET last_login_at = now()
WHERE id = $1;


-- name: DeleteRefreshToken :exec
delete from auth_sessions
where refresh_token_hash = $1 and user_id = $2
;


-- name: CreateRefreshToken :exec
INSERT INTO auth_sessions(
  user_id, 
  refresh_token_hash, 
  ip_address, 
  user_agent,
  expires_at
) VALUES ($1, $2, $3, $4, $5);


-- name: GetTokenForUser :one
select *
from auth_sessions
where user_id = $1
limit 1
;


-- name: CheckIfContactIsTaken :one
select
    exists (
        select 1
        from users
        where
            email = any(sqlc.arg(contacts)::text[])
            or phone = any(sqlc.arg(contacts)::text[])
    ) as contact_exists
;

