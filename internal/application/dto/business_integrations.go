package dto

import (
	"time"

	"github.com/google/uuid"
)

type IntegrationSecretDto struct {
	IntegrationId uuid.UUID `json:"integration_id" validate:"required"`
	KeyName       string    `json:"key_name" validate:"required,oneof=consumer_key consumer_secret"`
}

type UpsertIntegrationSecretDto struct {
	IntegrationId uuid.UUID `json:"integration_id"`
	KeyName       string    `json:"key_name" validate:"required,oneof=consumer_key consumer_secret"`
	KeyValue      string    `json:"key_value" validate:"required,min=4"`
}

type CreateIntegrationDto struct {
	Provider    string `json:"provider" validate:"required"`
	Environment string `json:"environment" validate:"required,oneof=sandbox staging production"`
}

type IntegrationRecord struct {
	ID          string     `json:"id"`
	Provider    string     `json:"provider"`
	Environment string     `json:"environment"`
	CreatedAt   *time.Time `json:"created_at"`
}
