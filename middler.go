// Package middler provides an HTTP middleware that logs a configured request header
// using slog when the header is present. It never mutates the request nor generates
// values when the header is absent.
package middler

import (
	"fmt"
	"log/slog"
	"net/http"
)

// Middleware is the type returned by New. It wraps an http.Handler producing another one.
// Usage:
//
//	mw, _ := middler.New(middler.WithHeader("myth-trace-id"))
//	mux.Handle("/", mw(finalHandler))
type Middleware func(next http.Handler) http.Handler

// cfg holds internal configuration produced by functional options.
type cfg struct {
	headerName string
	attrKey    string // precomputed attribute key: header.<snake_case>
	level      slog.Level
	logger     *slog.Logger
}

// New constructs a Middleware using provided functional options. It validates
// required parameters and returns an error if misconfigured. When the logger
// is not supplied it falls back to slog.Default().
func New(opts ...Option) (Middleware, error) {
	c := &cfg{
		level: slog.LevelInfo,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("applying option: %w", err)
		}
	}

	if c.headerName == "" {
		return nil, fmt.Errorf("missing header name: use WithHeader")
	}
	if c.logger == nil {
		c.logger = slog.Default()
	}

	return buildMiddleware(*c), nil
}
