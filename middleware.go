package middler

import (
    "net/http"
    "log/slog"
)

// buildMiddleware materializes the closure capturing an immutable copy of cfg.
func buildMiddleware(c cfg) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Read header and log only if present. Never mutate request.
            if v := r.Header.Get(c.headerName); v != "" {
                c.logger.Log(r.Context(), c.level, "request header", slog.String(c.attrKey, v))
            }
            next.ServeHTTP(w, r)
        })
    }
}
