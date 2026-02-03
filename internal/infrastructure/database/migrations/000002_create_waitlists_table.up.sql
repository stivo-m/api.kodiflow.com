CREATE TYPE WaitlistStatus AS ENUM('pending', 'invited', 'registered');


CREATE TABLE IF NOT EXISTS waitlists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    source TEXT, 
    status WaitlistStatus DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE UNIQUE INDEX IF NOT EXISTS waitlist_email_unique_idx ON waitlists (LOWER(email));
CREATE INDEX IF NOT EXISTS waitlist_created_at_idx ON waitlists (created_at);

