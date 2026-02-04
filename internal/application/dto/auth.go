package dto

import (
	"time"
)

type LogoutDto struct {
	RefreshToken string `json:"refresh_token" validate:"required,min=6"`
}

type RefreshTokenDto struct {
	RefreshToken string `json:"refresh_token" validate:"required,min=6"`
	IpAddress    string `json:"-"`
	UserAgent    string `json:"-"`
}

type RefreshTokenResponse struct {
	AccessToken           string     `json:"access_tokenk"`
	AccessTokenExpiresAt  *time.Time `json:"access_token_expires_at"`
	RefreshToken          string     `json:"refresh_token"`
	RefreshTokenExpiresAt *time.Time `json:"refresh_token_expires_at"`
}

type CreateUserDto struct {
	FullName string `json:"full_name" validate:"required,min=5"`
	Email    string `json:"email" validate:"required,email,min=5"`
	Phone    string `json:"phone" validate:"required,min=5"`
	Password string `json:"password" validate:"required,min=4"`
}

type EmailLoginDto struct {
	Email     string `json:"email" validate:"required,min=5"`
	Password  string `json:"password" validate:"required,min=4"`
	IpAddress string `json:"-"`
	UserAgent string `json:"-"`
}

type PhoneLoginDto struct {
	Phone     string `json:"phone" validate:"required,min=5"`
	Password  string `json:"password" validate:"required,min=4"`
	IpAddress string `json:"-"`
	UserAgent string `json:"-"`
}

type UserRecord struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

type LoginResponse struct {
	User   UserRecord           `json:"user"`
	Tokens RefreshTokenResponse `json:"tokens"`
}

type VerifyEmailDto struct {
	Email string `json:"email" validate:"required,email,min=5"`
	Code  string `json:"code" validate:"required,min=6"`
}
