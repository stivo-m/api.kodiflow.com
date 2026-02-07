package dto

import "time"

type CreateInvocieItemDto struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" validate:"required,min=1"`
	VatRate     float64 `json:"vat_rate" validate:"required,min=1"`
}

type CreateInvoiceDto struct {
	InvoiceNumber string                 `json:"invoice_number" validate:"required,min=4"`
	LineItems     []CreateInvocieItemDto `json:"line_items" validate:"required,min=1,dive"`
}

type InvoiceRecord struct {
	ID            string     `json:"id,omitzero"`
	InvoiceNumber string     `json:"invoice_number,omitzero"`
	Currency      string     `json:"currency,omitzero"`
	Subtotal      float64    `json:"subtotal,omitzero"`
	VatTotal      float64    `json:"vat_total,omitzero"`
	Total         float64    `json:"total,omitzero"`
	Status        string     `json:"status,omitzero"`
	IssuedAt      *time.Time `json:"issued_at,omitzero"`
	CreatedAt     *time.Time `json:"created_at,omitzero"`
	UpdatedAt     *time.Time `json:"updated_at,omitzero"`
}
