CREATE TABLE IF NOT EXISTS users(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
 
  full_name VARCHAR(255),
  phone VARCHAR(30),
 
  is_email_verified BOOLEAN DEFAULT FALSE,
  is_active BOOLEAN DEFAULT TRUE,
  
  last_login_at TIMESTAMPTZ,

  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_idx
ON users (LOWER(email));

CREATE UNIQUE INDEX IF NOT EXISTS users_phone_idx 
ON users (phone);

