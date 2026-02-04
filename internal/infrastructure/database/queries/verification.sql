-- name: StoreVerificationRecord :exec
INSERT INTO verifications(
  user_id, verification_type, verification_code
) VALUES ($1, $2, $3);

-- name: GetVerifiationByCodeForUser :one
select *
from verifications
where user_id = $1 and verification_code = $2
limit 1
;


-- name: MarkVerificationCodeAsConsumed :exec
delete from verifications
where id = $1
;

