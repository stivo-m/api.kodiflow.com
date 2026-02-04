package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stivo-m/api.kodiflow.com/internal/application/dto"
	"github.com/stivo-m/api.kodiflow.com/internal/application/usecase"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
	"github.com/stivo-m/api.kodiflow.com/internal/presentation/api/middleware"
	"github.com/stivo-m/api.kodiflow.com/pkg/api"
	"github.com/stivo-m/api.kodiflow.com/pkg/helpers"
)

type authhandler struct {
	usecase *usecase.AuthUsecase
}

// New auth handlers
func newAuthHandler(queries *database.Queries, pool *pgxpool.Pool) *authhandler {
	return &authhandler{
		usecase: usecase.NewAuthUsecase(queries, pool),
	}
}

// Register routes
func (h *authhandler) RegisterRoutes(router *http.ServeMux) {
	r := http.NewServeMux()
	router.Handle("/auth/", http.StripPrefix("/auth", r))

	r.Handle("POST /email-login", helpers.ValidateBody(h.emailLoginHandler))
	r.Handle("POST /phone-login", helpers.ValidateBody(h.phoneLoginHandler))
	r.Handle("POST /register", helpers.ValidateBody(h.createUserHandler))

	r.Handle("POST /logout",
		middleware.AuthMiddleware(
			helpers.ValidateBody(h.logoutHandler),
		),
	)
	r.Handle("POST /refresh-token",
		middleware.AuthMiddleware(
			helpers.ValidateBody(h.refreshTokenHandler),
		),
	)

	r.Handle("POST /verify/email", helpers.ValidateBody(h.verifyEmailHandler))
}

// Email login route
func (h *authhandler) emailLoginHandler(w http.ResponseWriter, r *http.Request, payload *dto.EmailLoginDto) {
	data := payload
	data.UserAgent = r.UserAgent()
	data.IpAddress = r.RemoteAddr

	result := h.usecase.EmailLoginUsecase(r.Context(), data)
	api.ProcessUsecaseResponse(result, w)
}

// Phone login route
func (h *authhandler) phoneLoginHandler(w http.ResponseWriter, r *http.Request, payload *dto.PhoneLoginDto) {
	data := payload
	data.UserAgent = r.UserAgent()
	data.IpAddress = r.RemoteAddr

	result := h.usecase.PhoneLoginUsecase(r.Context(), data)
	api.ProcessUsecaseResponse(result, w)
}

// Create user route
func (h *authhandler) createUserHandler(w http.ResponseWriter, r *http.Request, payload *dto.CreateUserDto) {
	data := payload

	result := h.usecase.CreateUserUsecase(r.Context(), data)
	api.ProcessUsecaseResponse(result, w)
}

// Logout route
func (h *authhandler) logoutHandler(w http.ResponseWriter, r *http.Request, payload *dto.LogoutDto) {
	result := h.usecase.LogoutUsecase(r.Context(), payload.RefreshToken)
	api.ProcessUsecaseResponse(result, w)
}

// Refresh token route
func (h *authhandler) refreshTokenHandler(w http.ResponseWriter, r *http.Request, payload *dto.RefreshTokenDto) {
	result := h.usecase.RefreshTokenUsecase(r.Context(), payload)
	api.ProcessUsecaseResponse(result, w)
}

// Verify email address
func (h *authhandler) verifyEmailHandler(w http.ResponseWriter, r *http.Request, payload *dto.VerifyEmailDto) {
	result := h.usecase.VerifyEmailUsecase(r.Context(), payload)
	api.ProcessUsecaseResponse(result, w)
}
