CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    aggregate_type TEXT NOT NULL,
    aggregate_id UUID NOT NULL,

    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,

    status TEXT NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    last_error TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ
);


CREATE INDEX IF NOT EXISTS outbox_pending_idx
    ON outbox_events (status, created_at);

CREATE INDEX IF NOT EXISTS outbox_aggregate_idx
    ON outbox_events (aggregate_type, aggregate_id);

