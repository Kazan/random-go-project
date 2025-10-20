package middler

import (
	"errors"
	"log/slog"
	"strings"
)

// Option represents a functional option that mutates the middleware configuration.
// It must return an error if the provided value is invalid so construction can fail fast.
type Option func(*cfg) error

// WithHeader configures the header name that will be read and logged if present.
// The name must be non-empty. Leading/trailing whitespace is trimmed before use.
// Calling WithHeader more than once returns an error (avoid silent override).
// WARNING: Be cautious not to log sensitive headers (e.g. Authorization, Cookie).
func WithHeader(name string) Option {
	return func(c *cfg) error {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return errors.New("header name cannot be empty")
		}
		if c.headerName != "" && c.headerName != trimmed {
			return errors.New("header already configured; multiple WithHeader calls not allowed")
		}
		c.headerName = trimmed
		c.attrKey = "header." + toSnakeLower(trimmed)
		return nil
	}
}

// WithLevel sets the log level used when emitting the header log record.
func WithLevel(level slog.Level) Option {
	return func(c *cfg) error {
		c.level = level
		return nil
	}
}

// WithLogger supplies the slog.Logger to use. If nil is provided the constructor
// falls back to slog.Default(). This option never errors.
func WithLogger(logger *slog.Logger) Option {
	return func(c *cfg) error {
		c.logger = logger
		return nil
	}
}

// toSnakeLower converts a header name (potentially containing hyphens) into
// lower-case snake_case. Only hyphens are replaced; other characters are left as-is.
func toSnakeLower(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	return strings.ReplaceAll(s, "-", "_")
}
