-- name: CreateBusinessIntegration :one
INSERT INTO business_integrations (
  business_id,
  provider,
  environment
) VALUES (
  $1, $2, $3
)
RETURNING *;


-- name: GetBusinessIntegrationByID :one
select *
from business_integrations
where id = $1 and business_id = $2
;


-- name: ListBusinessIntegrations :many
select *
from business_integrations
where business_id = $1
order by created_at desc
;


-- name: GetBusinessIntegrationByProvider :one
select *
from business_integrations
where business_id = $1 and provider = $2 and environment = $3
;


-- name: UpdateBusinessIntegrationStatus :one
UPDATE business_integrations
SET status = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;


-- name: UpdateBusinessIntegrationEnvironment :one
UPDATE business_integrations
SET environment = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;


-- name: DeleteBusinessIntegration :exec
delete from business_integrations
where id = $1 and business_id = $2
;


-- ---------------------
-- Integration secrets
-- ---------------------
-- name: UpsertIntegrationSecret :one
INSERT INTO integration_secrets (
  integration_id,
  key_name,
  encrypted_value
) VALUES (
  $1, $2, $3
)
ON CONFLICT (integration_id, key_name)
DO UPDATE SET
  encrypted_value = EXCLUDED.encrypted_value,
  updated_at = now()
RETURNING *;


-- name: ListIntegrationSecrets :many
select key_name, encrypted_value
from integration_secrets
where integration_id = $1
;


-- name: GetIntegrationSecret :one
select encrypted_value
from integration_secrets
where integration_id = $1 and key_name = $2
;


-- name: DeleteIntegrationSecret :exec
delete from integration_secrets
where integration_id = $1 and key_name = $2
;

