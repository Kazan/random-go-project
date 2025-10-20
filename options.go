package middleware

import "fmt"

// Logger is a minimal logging abstraction used by RequestIDMiddleware.
//
// Any logger providing a Printf method (compatible with log.Logger) can be
// supplied via WithLogger. If no logger is configured, request IDs are not
// logged.
type Logger interface {
    Printf(format string, args ...any)
}

// DefaultHeaderName is the default HTTP request header whose value, when
// present and non-empty, is logged by RequestIDMiddleware.
const DefaultHeaderName = "myth-tracer-id"

// Option configures RequestIDMiddleware. Implementations must be side-effect
// free beyond mutating the provided *cfg and should return an error if the
// supplied value is invalid.
type Option func(*cfg) error

// cfg holds middleware configuration. It is intentionally unexported to allow
// forward-compatible changes without breaking callers.
type cfg struct {
    headerName string
    logger     Logger
}

// applyOptions constructs a cfg with defaults and sequentially applies opts.
func applyOptions(opts []Option) (*cfg, error) {
    c := &cfg{headerName: DefaultHeaderName}
    for _, opt := range opts {
        if opt == nil {
            continue // ignore nil option for convenience
        }
        if err := opt(c); err != nil {
            return nil, fmt.Errorf("apply option: %w", err)
        }
    }
    return c, nil
}

// WithHeaderName overrides the header name inspected for the request ID.
//
// Returns an error if name is empty after trimming spaces.
func WithHeaderName(name string) Option {
    return func(c *cfg) error {
        if len(name) == 0 {
            return fmt.Errorf("header name cannot be empty")
        }
        c.headerName = name
        return nil
    }
}

// WithLogger sets the logger used to output request ID lines. Returns an error
// if l is nil.
func WithLogger(l Logger) Option {
    return func(c *cfg) error {
        if l == nil {
            return fmt.Errorf("logger cannot be nil")
        }
        c.logger = l
        return nil
    }
}
