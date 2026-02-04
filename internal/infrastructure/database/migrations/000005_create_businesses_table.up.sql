DO $$
BEGIN
  CREATE TYPE BusinessStatus AS ENUM ('pending', 'active', 'suspended');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END
$$;

CREATE TABLE IF NOT EXISTS businesses(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  legal_name VARCHAR(255) NOT NULL,
  trading_name VARCHAR(255),
  kra_pin VARCHAR(20) UNIQUE NOT NULL,
  business_type VARCHAR(50), 
  email VARCHAR(255),
  phone VARCHAR(30),
  country VARCHAR(2) DEFAULT 'KE',
  status BusinessStatus DEFAULT 'pending', 
  
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);


CREATE UNIQUE INDEX businesses_email_unique_idx
ON businesses(email)
WHERE email IS NOT NULL;

CREATE UNIQUE INDEX businesses_phone_unique_idx
ON businesses(phone)
WHERE phone IS NOT NULL;


-- Business Users
CREATE TABLE IF NOT EXISTS business_users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  business_id UUID REFERENCES businesses(id),
  user_id UUID REFERENCES users(id),
  role VARCHAR(30), 
  is_active BOOLEAN DEFAULT TRUE,
   
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()


);


-- Ensure that a user must have only one entry here per business
CREATE UNIQUE INDEX business_users_idx
ON business_users(business_id, user_id);


-- KYC related
DO $$
BEGIN
  CREATE TYPE DocumentStatus AS ENUM ('pending', 'verified', 'rejected');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END
$$;




CREATE TABLE IF NOT EXISTS business_kyc(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  business_id UUID REFERENCES businesses(id),
  document_type VARCHAR(50), 
  document_url TEXT,
  status DocumentStatus DEFAULT 'pending', 
  status_text TEXT,
  verified_by UUID REFERENCES users(id),
  verified_at TIMESTAMPTZ,
   
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()

);

