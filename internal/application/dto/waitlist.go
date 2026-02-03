package dto

type CreateWaitlistDto struct {
	Email  string `json:"email" validate:"required,min=4"`
	Source string `json:"source"`
}
