package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/stivo-m/api.kodiflow.com/pkg/helpers"
)

func BusinessContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		businessIDStr := r.Header.Get("X-Business-ID")
		if businessIDStr == "" {
			businessIDStr = r.URL.Query().Get("business_id")
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

		ctx := context.WithValue(r.Context(), helpers.BusinessContextKey, businessID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
