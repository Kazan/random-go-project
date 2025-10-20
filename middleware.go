// Package middleware provides a transparent HTTP middleware that forwards the
// incoming request to the next handler without any modification or side effect.
//
// The Middleware function is intentionally minimal: it performs no logging,
// metrics, tracing, header mutation, context changes, panic recovery, or other
// behaviors. It simply invokes the supplied next handler. This serves as a
// lightweight building block or placeholder within a middleware chain when no
// additional processing is required.
//
// If next is nil the function panics to fail fast, signaling a programmer error.
package middleware

import "net/http"

// Middleware returns an http.Handler that invokes next without altering the
// ResponseWriter or Request.
//
// Panics if next is nil.
func Middleware(next http.Handler) http.Handler {
	if next == nil {
		panic("middleware: next handler is nil")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
