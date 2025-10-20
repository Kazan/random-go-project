// Package middleware provides thin HTTP middleware components.
//
// Core Components:
//   - Middleware: pass-through wrapper that invokes the next handler without
//     side effects (no logging, metrics, tracing, mutation, or recovery).
//   - RequestIDMiddleware: optional logging middleware that, when configured
//     with a Logger, logs a request ID extracted from a specific header before
//     delegating to the next handler.
//
// RequestIDMiddleware Usage:
//
//	handler, err := middleware.RequestIDMiddleware(mux,
//	    middleware.WithLogger(log.Default()),
//	    middleware.WithHeaderName("X-Request-ID"),
//	)
//	if err != nil { /* handle error */ }
//
// If the header is absent or empty, no log line is emitted. If no Logger is
// provided, the handler behaves like the pass-through Middleware. No context
// values are injected; the behavior is intentionally minimal and additive.
package middleware
