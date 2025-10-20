package middleware

import (
    "net/http"
)

// RequestIDMiddleware returns a handler that logs a request ID extracted from
// the configured header if present and non-empty, then delegates to next.
//
// The header name defaults to DefaultHeaderName and can be changed with
// WithHeaderName. Logging occurs only if a Logger is configured with
// WithLogger. Missing or empty header values are ignored silently.
//
// Panics if next is nil.
func RequestIDMiddleware(next http.Handler, opts ...Option) (http.Handler, error) {
    if next == nil {
        panic("middleware: next handler is nil")
    }
    c, err := applyOptions(opts)
    if err != nil {
        return nil, err
    }
    // If no logger configured we can early return pass-through wrapper.
    if c.logger == nil {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            next.ServeHTTP(w, r)
        }), nil
    }
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := r.Header.Get(c.headerName)
        if id != "" { // treat empty as absent
            logRequestID(c, r, id)
        }
        next.ServeHTTP(w, r)
    }), nil
}

// logRequestID formats and writes the log line. Separated for clarity and to
// allow future extension (e.g., metrics) without cluttering the handler body.
func logRequestID(c *cfg, r *http.Request, id string) {
    if c.logger == nil || id == "" { // defensive; should be filtered earlier
        return
    }
    c.logger.Printf("request_id=%s method=%s path=%s", id, r.Method, r.URL.Path)
}
