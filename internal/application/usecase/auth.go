package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stivo-m/api.kodiflow.com/internal/application/dto"
	"github.com/stivo-m/api.kodiflow.com/internal/domain/events"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
	"github.com/stivo-m/api.kodiflow.com/pkg/api"
	"github.com/stivo-m/api.kodiflow.com/pkg/hashing"
	"github.com/stivo-m/api.kodiflow.com/pkg/helpers"
)

type AuthUsecase struct {
	queries *database.Queries
	pool    *pgxpool.Pool
	logger  *slog.Logger
}

// Instantiates a new auth usecase
func NewAuthUsecase(queries *database.Queries, pool *pgxpool.Pool) *AuthUsecase {
	logger := slog.Default().WithGroup("metadata").With(slog.String("type", "usecase"), slog.String("name", "auth"))
	return &AuthUsecase{queries: queries, pool: pool, logger: logger}
}

// Handle creating a new user account
func (u *AuthUsecase) CreateUserUsecase(ctx context.Context, payload *dto.CreateUserDto) *api.ApiResponse {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		u.logger.Error("failed to start transaction when creating a user")
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Unable to handle logout",
		}
	}

	defer tx.Rollback(ctx)
	qtx := u.queries.WithTx(tx)

	contacts := []string{
		payload.Email,
		payload.Phone,
	}
	exists, err := qtx.CheckIfContactIsTaken(ctx, contacts)
	if err != nil || exists {
		u.logger.Error("failed to check if a contact is taken when creating a user", "error", err)
		return &api.ApiResponse{
			Code:    400,
			Errors:  err,
			Message: "One or more of your contacts are already taken",
		}
	}

	passwordHash, err := hashing.HashPassword(payload.Password)
	if err != nil {
		u.logger.Error("failed to hash user password when creating a user", "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to hash user password",
		}
	}

	userParams := database.CreateUserAccountParams{
		Email:        payload.Email,
		Phone:        pgtype.Text{String: payload.Phone, Valid: true},
		FullName:     pgtype.Text{String: payload.FullName, Valid: true},
		PasswordHash: passwordHash,
	}

	user, err := qtx.CreateUserAccount(ctx, userParams)
	if err != nil {
		u.logger.Error("failed to create a user account on the database", "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to register user account",
		}
	}

	p := events.UserEventPayload{
		ID:    user.ID,
		Email: user.Email,
		Phone: user.Phone.String,
	}

	jsonPayload, _ := json.Marshal(p)

	evtPayload := database.CreateOutboxEventParams{
		AggregateType: string(events.UserAggregateType),
		AggregateID:   user.ID,
		EventType:     string(events.UserCreatedEvent),
		Payload:       jsonPayload,
	}

	err = qtx.CreateOutboxEvent(ctx, evtPayload)
	if err != nil {
		u.logger.Error("failed to create outbox event for registered user", "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to create outbox event for user registration",
		}
	}

	tx.Commit(ctx)

	return &api.ApiResponse{
		Code:    201,
		Message: "User created successfully. Follow prompts on your email on how to login",
	}
}

// Handle phone login to the application
func (u *AuthUsecase) PhoneLoginUsecase(ctx context.Context, payload *dto.PhoneLoginDto) *api.ApiResponse {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		u.logger.Error("failed to start transaction while using phone login", "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Unable to handle logout",
		}
	}

	defer tx.Rollback(ctx)
	qtx := u.queries.WithTx(tx)

	findParam := pgtype.Text{Valid: true, String: payload.Phone}
	user, err := qtx.FindUserByPhone(ctx, findParam)
	if err != nil {
		u.logger.Error("failed to find user by phone using phone login", "error", err)
		if err == sql.ErrNoRows {
			return &api.ApiResponse{
				Code:    404,
				Message: "Either email or password is invalid",
			}
		}

		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to login user",
		}
	}

	if !user.IsActive.Bool || !user.IsEmailVerified.Bool {

		return &api.ApiResponse{
			Code:    403,
			Message: "User is either inactive or has not verified their email address",
		}
	}

	accessToken, err := hashing.CreateJwtToken(user.ID)
	if err != nil {
		u.logger.Error("failed to generate user access token when using phone login", "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to generate access token",
		}
	}

	refreshToken, err := hashing.GenerateRefreshToken(hashing.RefreshTokenTtl)
	if err != nil {
		u.logger.Error("failed to generate refresh token when using phone login", "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to generate refresh token",
		}
	}

	ipAddres, _ := helpers.ParseIPAddress(payload.IpAddress)
	params := database.CreateRefreshTokenParams{
		RefreshTokenHash: pgtype.Text{String: refreshToken.HashedToken, Valid: true},
		ExpiresAt:        pgtype.Timestamptz{Time: refreshToken.ExpiresAt, Valid: true},
		UserAgent:        pgtype.Text{String: payload.UserAgent, Valid: true},
		IpAddress:        ipAddres,
		UserID:           user.ID,
	}

	err = qtx.CreateRefreshToken(ctx, params)
	if err != nil {
		u.logger.Error("failed to persist refresh token when using phone login", "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to store refresh token",
		}
	}

	tx.Commit(ctx)

	output := dto.LoginResponse{
		User: dto.UserRecord{
			ID:       user.ID.String(),
			FullName: user.FullName.String,
			Email:    user.Email,
			Phone:    user.Phone.String,
		},

		Tokens: dto.RefreshTokenResponse{
			AccessToken:           accessToken,
			AccessTokenExpiresAt:  &hashing.AccessTokenExpiry,
			RefreshToken:          refreshToken.RawToken,
			RefreshTokenExpiresAt: &refreshToken.ExpiresAt,
		},
	}

	return &api.ApiResponse{
		Code:    200,
		Data:    output,
		Message: "User logged in successfully",
	}
}

// Handle email login to the application
func (u *AuthUsecase) EmailLoginUsecase(ctx context.Context, payload *dto.EmailLoginDto) *api.ApiResponse {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		u.logger.Error("failed to start db transaction when using email login", "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Unable to handle logout",
		}
	}

	defer tx.Rollback(ctx)
	qtx := u.queries.WithTx(tx)

	user, err := qtx.FindUserByEmail(ctx, payload.Email)
	if err != nil {
		u.logger.Error("failed to find user by email when using email login", "error", err)
		if err == sql.ErrNoRows {
			return &api.ApiResponse{
				Code:    404,
				Message: "Either email or password is invalid",
			}
		}

		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to login user",
		}
	}

	if !user.IsActive.Bool || !user.IsEmailVerified.Bool {
		return &api.ApiResponse{
			Code:    403,
			Message: "User is either inactive or has not verified their email address",
		}
	}

	accessToken, err := hashing.CreateJwtToken(user.ID)
	if err != nil {
		u.logger.Error("failed to generate access token when using email login", "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to generate access token",
		}
	}

	refreshToken, err := hashing.GenerateRefreshToken(hashing.RefreshTokenTtl)
	if err != nil {
		u.logger.Error("failed to generate refresh token when using email login", "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to generate refresh token",
		}
	}

	ipAddres, _ := helpers.ParseIPAddress(payload.IpAddress)
	params := database.CreateRefreshTokenParams{
		RefreshTokenHash: pgtype.Text{String: refreshToken.HashedToken, Valid: true},
		ExpiresAt:        pgtype.Timestamptz{Time: refreshToken.ExpiresAt, Valid: true},
		UserAgent:        pgtype.Text{String: payload.UserAgent, Valid: true},
		IpAddress:        ipAddres,
		UserID:           user.ID,
	}

	err = qtx.CreateRefreshToken(ctx, params)
	if err != nil {
		u.logger.Error("failed to persist refresh token when using email login", "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to store refresh token",
		}
	}

	tx.Commit(ctx)

	output := dto.LoginResponse{
		User: dto.UserRecord{
			ID:       user.ID.String(),
			FullName: user.FullName.String,
			Email:    user.Email,
			Phone:    user.Phone.String,
		},

		Tokens: dto.RefreshTokenResponse{
			AccessToken:           accessToken,
			AccessTokenExpiresAt:  &hashing.AccessTokenExpiry,
			RefreshToken:          refreshToken.RawToken,
			RefreshTokenExpiresAt: &refreshToken.ExpiresAt,
		},
	}

	return &api.ApiResponse{
		Code:    200,
		Data:    output,
		Message: "User logged in successfully",
	}
}

// Handle verify email address
func (u *AuthUsecase) VerifyEmailUsecase(ctx context.Context, payload *dto.VerifyEmailDto) *api.ApiResponse {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		u.logger.Error("failed to start db transaction when verifying email", "error", err)
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Unable to handle logout",
		}
	}

	defer tx.Rollback(ctx)
	qtx := u.queries.WithTx(tx)

	user, err := qtx.FindUserByEmail(ctx, payload.Email)
	if err != nil {
		u.logger.Error("failed to find user by email when verifying email", "error", err)
		return &api.ApiResponse{
			Code:    400,
			Errors:  err,
			Message: "Either user is not found or code is expired",
		}
	}

	params := database.GetVerifiationByCodeForUserParams{
		UserID:           user.ID,
		VerificationCode: payload.Code,
	}
	v, err := qtx.GetVerifiationByCodeForUser(ctx, params)
	if err != nil {
		u.logger.Error("failed to find verificaiton for user when verifying email", "user_id", user.ID, "error", err)
		return &api.ApiResponse{
			Code:    400,
			Errors:  err,
			Message: "Either user is not found or code is expired",
		}
	}

	err = qtx.MarkEmailAsVerified(ctx, user.ID)
	if err != nil {
		u.logger.Error("failed to mark email as verified user when verifying email", "user_id", user.ID, "error", err)
		return &api.ApiResponse{
			Code:    400,
			Errors:  err,
			Message: "Either user is not found or code is expired",
		}
	}

	err = qtx.MarkVerificationCodeAsConsumed(ctx, v.ID)
	if err != nil {
		u.logger.Error("failed to mark verification record as consumed user when verifying email", "user_id", user.ID, "error", err)
		return &api.ApiResponse{
			Code:    400,
			Errors:  err,
			Message: "Either user is not found or code is expired",
		}
	}

	tx.Commit(ctx)

	return &api.ApiResponse{
		Code:    200,
		Message: "Email has been verified successfully",
	}
}

// Handle forgot password initiation

// Handle reset password flow

// Handle refresh token flow
func (u *AuthUsecase) RefreshTokenUsecase(ctx context.Context, payload *dto.RefreshTokenDto) *api.ApiResponse {
	userId, err := helpers.GetUserFromContext(ctx)
	if err != nil {
		return &api.ApiResponse{
			Code:    401,
			Errors:  err,
			Message: "Unauthorized",
		}
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Unable to handle logout",
		}
	}

	defer tx.Rollback(ctx)
	qtx := u.queries.WithTx(tx)

	session, err := qtx.GetTokenForUser(ctx, userId)
	if err != nil {
		return &api.ApiResponse{
			Code:    403,
			Errors:  err,
			Message: "Failed to obtain refresh token for the user",
		}
	}

	hashedToken := hashing.HashRefreshToken(payload.RefreshToken)
	if hashedToken != session.RefreshTokenHash.String || time.Now().Before(session.CreatedAt.Time) {
		return &api.ApiResponse{
			Code:    403,
			Errors:  err,
			Message: "Referesh token is invalid or has expired",
		}
	}

	deleteParams := database.DeleteRefreshTokenParams{
		UserID:           userId,
		RefreshTokenHash: session.RefreshTokenHash,
	}
	defer qtx.DeleteRefreshToken(ctx, deleteParams)

	newToken, err := hashing.GenerateRefreshToken(hashing.RefreshTokenTtl)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to generate a new refresh token",
		}
	}

	ipAddres, _ := helpers.ParseIPAddress(payload.IpAddress)

	params := database.CreateRefreshTokenParams{
		RefreshTokenHash: pgtype.Text{String: newToken.HashedToken, Valid: true},
		ExpiresAt:        pgtype.Timestamptz{Time: newToken.ExpiresAt, Valid: true},
		UserAgent:        pgtype.Text{String: payload.UserAgent, Valid: true},
		IpAddress:        ipAddres,
		UserID:           userId,
	}

	err = qtx.CreateRefreshToken(ctx, params)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to store newly created refresh token",
		}
	}

	access, err := hashing.CreateJwtToken(userId)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to generate a new access token",
		}
	}

	tx.Commit(ctx)

	output := dto.RefreshTokenResponse{
		AccessToken:           access,
		AccessTokenExpiresAt:  &hashing.AccessTokenExpiry,
		RefreshToken:          newToken.RawToken,
		RefreshTokenExpiresAt: &newToken.ExpiresAt,
	}

	return &api.ApiResponse{
		Code:    200,
		Data:    output,
		Message: "Token refreshed successfully",
	}
}

// Handle logout flow
func (u *AuthUsecase) LogoutUsecase(ctx context.Context, refreshToken string) *api.ApiResponse {
	userId, err := helpers.GetUserFromContext(ctx)
	if err != nil {
		return &api.ApiResponse{
			Code:    401,
			Errors:  err,
			Message: "Unauthorized",
		}
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Unable to handle logout",
		}
	}

	defer tx.Rollback(ctx)
	qtx := u.queries.WithTx(tx)

	session, err := qtx.GetTokenForUser(ctx, userId)
	if err != nil {
		return &api.ApiResponse{
			Code:    403,
			Errors:  err,
			Message: "Failed to obtain refresh token for the user",
		}
	}

	hashedToken := hashing.HashRefreshToken(refreshToken)
	if hashedToken != session.RefreshTokenHash.String || time.Now().Before(session.CreatedAt.Time) {
		return &api.ApiResponse{
			Code:    403,
			Errors:  err,
			Message: "referesh token is invalid or has expired",
		}
	}

	deleteParams := database.DeleteRefreshTokenParams{
		UserID:           userId,
		RefreshTokenHash: session.RefreshTokenHash,
	}
	err = qtx.DeleteRefreshToken(ctx, deleteParams)
	if err != nil {
		return &api.ApiResponse{
			Code:    500,
			Errors:  err,
			Message: "Failed to handle logout",
		}
	}

	tx.Commit(ctx)
	return &api.ApiResponse{
		Code:    200,
		Message: "Logout completed successfully ",
	}
}
