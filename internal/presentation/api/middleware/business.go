package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
	"github.com/stivo-m/api.kodiflow.com/pkg/helpers"
)

/*
*	Checks for the business id on params and header and
* validates that the logged in user is part of the business users
 */
func BusinessContextMiddleware(q *database.Queries) middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			businessIDStr := r.Header.Get("X-Business-ID")
			if businessIDStr == "" {
				businessIDStr = r.URL.Query().Get("business_id")
				if businessIDStr == "" {
					businessIDStr = r.PathValue("businessId")
					slog.Info(businessIDStr)
				}
			}

			if businessIDStr == "" {
				http.Error(w, "business_id is required", http.StatusBadRequest)
				return
			}

			businessID, err := uuid.Parse(businessIDStr)
			if err != nil {
				http.Error(w, "invalid business_id", http.StatusBadRequest)
				return
			}

			userID, err := helpers.GetUserFromContext(r.Context())
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			params := database.CheckIfUserIsPartOfBusinessParams{
				UserID:     userID,
				BusinessID: businessID,
			}

			ok, err := q.CheckIfUserIsPartOfBusiness(r.Context(), params)
			if err != nil {
				http.Error(w, "Business or user not found", http.StatusNotFound)
				return
			}

			if !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), helpers.BusinessContextKey, businessID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
