package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper executes a request against h and returns recorder.
func doRequest(t *testing.T, h http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func TestMiddleware_PassThroughStatusBodyHeaders(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom", "abc123")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, "hello")
	})

	wrapped := Middleware(next)
	require.NotNil(t, wrapped)

	rr := doRequest(t, wrapped, http.MethodGet, "/test")

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, "hello", rr.Body.String())
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.Equal(t, "abc123", rr.Header().Get("X-Custom"))
}

func TestMiddleware_NilNextPanics(t *testing.T) {
	assert.Panics(t, func() { _ = Middleware(nil) })
}
