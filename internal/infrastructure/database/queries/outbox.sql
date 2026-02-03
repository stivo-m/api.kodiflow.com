-- name: CreateOutboxEvent :exec
INSERT INTO outbox_events (
    aggregate_type,
    aggregate_id,
    event_type,
    payload
) VALUES (
    $1, $2, $3, $4
);


-- name: FetchPendingOutboxEvents :many
select *
from outbox_events
where status = 'pending'
order by created_at
limit $1
for update skip locked
;


-- name: MarkOutboxEventProcessing :exec
UPDATE outbox_events
SET status = 'processing'
WHERE id = $1;


-- name: MarkOutboxEventProcessed :exec
UPDATE outbox_events
SET
    status = 'processed',
    processed_at = now()
WHERE id = $1;


-- name: MarkOutboxEventFailed :exec
UPDATE outbox_events
SET
    status = 'pending',
    retry_count = retry_count + 1,
    last_error = $2
WHERE id = $1;


-- name: MarkOutboxEventDead :exec
UPDATE outbox_events
SET
    status = 'failed',
    last_error = $2
WHERE id = $1;


-- name: FetchStuckOutboxEvents :many
select *
from outbox_events
where status = 'processing' and created_at < now() - interval '10 minutes'
limit $1
;

