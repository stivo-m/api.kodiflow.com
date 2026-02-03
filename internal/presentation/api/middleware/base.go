package middleware

import "net/http"

type middlewareFunc func(http.Handler) http.Handler

// Wraps middlewares in a stack in sequence
func CreateMiddlewareStack(xs ...middlewareFunc) middlewareFunc {
	return func(next http.Handler) http.Handler {
		for i := len(xs) - 1; i >= 0; i-- {
			x := xs[i]
			next = x(next)
		}

		return next
	}
}
