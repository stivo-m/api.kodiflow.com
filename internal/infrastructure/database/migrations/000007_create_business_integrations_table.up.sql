CREATE TABLE IF NOT EXISTS business_integrations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  business_id UUID NOT NULL
    REFERENCES businesses(id) ON DELETE CASCADE,

  provider VARCHAR(50) NOT NULL,
  environment VARCHAR(20) NOT NULL DEFAULT 'sandbox',
  status VARCHAR(30) NOT NULL DEFAULT 'pending',

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS business_integrations_unique_provider
ON business_integrations (business_id, provider, environment);

CREATE INDEX IF NOT EXISTS business_integrations_business_id_idx
ON business_integrations (business_id);


CREATE TABLE IF NOT EXISTS integration_secrets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  integration_id UUID NOT NULL
    REFERENCES business_integrations(id) ON DELETE CASCADE,

  key_name VARCHAR(50) NOT NULL,
  encrypted_value TEXT NOT NULL,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (integration_id, key_name)
);

CREATE INDEX IF NOT EXISTS integration_secrets_integration_id_idx
ON integration_secrets (integration_id);

