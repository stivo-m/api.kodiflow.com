package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		processed := fmt.Sprintf("%s %s took: %v", r.Method, r.URL.Path, time.Since(start))
		slog.Info(processed)
	})
}
