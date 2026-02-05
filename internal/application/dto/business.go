package dto

import (
	"time"

	"github.com/google/uuid"
)

type BusinessKycRecord struct {
	ID           string     `json:"id"`
	DocumentType string     `json:"document_type"`
	DocumentUrl  string     `json:"document_url"`
	Status       string     `json:"status"`
	StatusText   string     `json:"status_text,omitzero"`
	VerifiedAt   *time.Time `json:"verified_at,omitzero"`
	CreatedAt    *time.Time `json:"created_at,omitzero"`
}

type BusinessKycResponse struct {
	Records []BusinessKycRecord `json:"records"`
}

type CreateBusinessKycDto struct {
	DocumentType string `json:"document_type" validate:"required,min=4"`
	DocumentUrl  string `json:"document_url" validate:"required,min=4"`
}

type BusinessRecord struct {
	ID           string     `json:"id"`
	LegalName    string     `json:"legal_name"`
	KraPin       string     `json:"kra_pin"`
	BusinessType string     `json:"business_type"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	Country      string     `json:"country"`
	Status       string     `json:"status"`
	CreatedAt    *time.Time `json:"created_at"`
}

type BusinessesResponse struct {
	Records []BusinessRecord `json:"records"`
}

type AddUserToBusinessDto struct {
	UserId     uuid.UUID `json:"user_id" validate:"required,min=4"`
	BusinessId uuid.UUID `json:"business_id" validate:"required,min=4"`
	Role       string    `json:"role" validate:"required,oneof=admin member"`
}

type CreateBusinessDto struct {
	LegalName    string `json:"legal_name" validate:"required,min=4"`
	TradingName  string `json:"trading_name" validate:"required,min=4"`
	KraPin       string `json:"kra_pin" validate:"required,min=4"`
	BusinessType string `json:"business_type" validate:"required,min=4"`
	Email        string `json:"email" validate:"required,email,min=4"`
	Phone        string `json:"phone" validate:"required,min=6"`
	Country      string `json:"country" validate:"required,min=2,max=4"`
}
