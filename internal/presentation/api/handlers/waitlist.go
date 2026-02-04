package handlers

// import (
// 	"net/http"
//
// 	"github.com/jackc/pgx/v5/pgxpool"
// 	"github.com/stivo-m/api.kodiflow.com/internal/application/dto"
// 	"github.com/stivo-m/api.kodiflow.com/internal/application/usecase"
// 	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
// 	"github.com/stivo-m/api.kodiflow.com/pkg/api"
// 	"github.com/stivo-m/api.kodiflow.com/pkg/helpers"
// )
//
// type waitlistHandler struct {
// 	usecase *usecase.WaitlistUsecase
// }
//
// // New waitlist handler
// func newWaitlistHandler(queries *database.Queries, pool *pgxpool.Pool) *waitlistHandler {
// 	return &waitlistHandler{
// 		usecase: usecase.NewWaitlistUsecase(queries, pool),
// 	}
// }
//
// // Register routes
// func (h *waitlistHandler) RegisterRoutes(router *http.ServeMux) {
// 	r := http.NewServeMux()
// 	router.Handle("/waitlists/", http.StripPrefix("/waitlists", r))
//
// 	r.Handle("POST /", helpers.ValidateBody(h.handleAddUserToWaitlist))
// }
//
// // Add user to waitlist
// func (h *waitlistHandler) handleAddUserToWaitlist(w http.ResponseWriter, r *http.Request, payload dto.CreateWaitlistDto) {
// 	result := h.usecase.AddUserToWaitlist(r.Context(), &payload)
// 	api.ProcessUsecaseResponse(result, w)
// }
