CREATE TABLE IF NOT EXISTS auth_sessions(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) NOT NULL,
  refresh_token_hash TEXT,
  ip_address INET,
  user_agent TEXT,
  expires_at TIMESTAMPTZ,
   
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()


);


CREATE INDEX IF NOT EXISTS auth_sessions_refresh_token_idx
ON auth_sessions(refresh_token_hash);

