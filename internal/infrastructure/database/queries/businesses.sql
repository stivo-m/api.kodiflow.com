-- name: CreateBusiness :one
INSERT INTO businesses (
  legal_name,
  trading_name,
  kra_pin,
  business_type,
  email,
  phone,
  country
)
VALUES (
  $1, $2, $3, $4, $5, $6, COALESCE($7, 'KE')
)
RETURNING *;


-- name: GetBusinessByID :one
select *
from businesses
where id = $1
;


-- name: GetBusinessByKRAPIN :one
select *
from businesses
where kra_pin = $1
;


-- name: ListBusinessesByStatus :many
select *
from businesses
where status = $1
order by created_at desc
;


-- name: UpdateBusinessStatus :exec
UPDATE businesses
SET status = $2,
    updated_at = now()
WHERE id = $1;


-- name: UpdateBusinessContact :exec
UPDATE businesses
SET email = $2,
    phone = $3,
    updated_at = now()
WHERE id = $1;


-- name: UpdateBusinessDetails :exec
UPDATE businesses
SET legal_name   = $2,
    trading_name = $3,
    business_type = $4,
    updated_at   = now()
WHERE id = $1;


-- name: DeleteBusiness :exec
delete from businesses
where id = $1
;


-- name: AddUserToBusiness :exec
INSERT INTO business_users (
  business_id,
  user_id,
  role
)
VALUES ($1, $2, $3);


-- name: GetUserBusinessRole :one
select role
from business_users
where business_id = $1 and user_id = $2 and is_active = true
;


-- name: ListBusinessesForUser :many
select b.*
from businesses b
join business_users bu on bu.business_id = b.id
where bu.user_id = $1 and bu.is_active = true
order by b.created_at desc
;


-- name: CreateBusinessKYC :one
INSERT INTO business_kyc (
  business_id,
  document_type,
  document_url,
  status
)
VALUES ($1, $2, $3, 'pending')
RETURNING *;


-- name: GetBusinessKYC :many
select *
from business_kyc
where business_id = $1
order by created_at desc
;


-- name: UpdateKYCStatus :exec
UPDATE business_kyc
SET status = $2,
    verified_by = $3,
    verified_at = now()
WHERE id = $1;


-- name: HasVerifiedKYC :one
select
    exists (
        select 1 from business_kyc where business_id = $1 and status = 'verified'
    ) as has_verified
;

