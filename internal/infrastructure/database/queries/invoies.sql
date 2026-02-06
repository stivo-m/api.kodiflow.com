-- name: CreateInvoice :one
INSERT INTO invoices(
  business_id, invoice_number, seller_pin, buyer_pin, subtotal, vat_total, total
) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: UpdateInvoiceWithKraDetails :exec
UPDATE invoices
SET kra_invoice_id = $1, kra_receipt_url = $1, updated_at = now()
WHERE id = $1;

-- name: UpdateInvoiceTotals :exec
UPDATE invoices
SET subtotal = $1, vat_total = $2, total = $3, updated_at = now()
WHERE business_id = $4 AND id = $5;


-- name: UpdateInvoiceStatus :exec
UPDATE invoices
SET status = $1, updated_at = now()
WHERE business_id = $2 AND id = $3;

-- name: AddLineInvoiceLineItem :exec
INSERT INTO invoice_items (
  invoice_id, description, quantity, unit_price, vat_rate, vat_amount, line_total
) VALUES($1, $2, $3, $4, $5, $6, $7);

