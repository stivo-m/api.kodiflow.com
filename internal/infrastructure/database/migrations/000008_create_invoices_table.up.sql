CREATE TABLE IF NOT EXISTS invoices(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
  invoice_number VARCHAR(100) NOT NULL,
  seller_pin VARCHAR(20) NOT NULL,
  buyer_pin VARCHAR(20),
  currency VARCHAR(3) DEFAULT 'KES',
  subtotal NUMERIC(12,2) NOT NULL,
  vat_total NUMERIC(12,2) NOT NULL,
  total NUMERIC(12,2) NOT NULL,
  status VARCHAR(30) NOT NULL DEFAULT 'draft', -- draft, issued, paid, voided
  kra_invoice_id VARCHAR(100),
  kra_receipt_url TEXT,
  issued_at TIMESTAMPTZ,
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  UNIQUE (business_id, invoice_number)
);

CREATE TABLE IF NOT EXISTS invoice_items(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
  description TEXT,
  quantity NUMERIC(10,2) NOT NULL,
  unit_price NUMERIC(12,2) NOT NULL,
  vat_rate NUMERIC(5,2) NOT NULL,
  vat_amount NUMERIC(12,2) NOT NULL,
  line_total NUMERIC(12,2) NOT NULL
);

