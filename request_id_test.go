package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubLogger is a concurrency-safe in-memory logger.
type stubLogger struct {
	mu  sync.Mutex
	log []string
}

func (s *stubLogger) Printf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	s.mu.Lock()
	s.log = append(s.log, msg)
	s.mu.Unlock()
}

func (s *stubLogger) entries() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.log))
	copy(out, s.log)
	return out
}

// helper similar to existing doRequest but allows header injection.
func doRequestWithHeader(t *testing.T, h http.Handler, method, path, header, value string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	if header != "" {
		req.Header.Set(header, value)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func TestRequestID_DefaultHeaderPresent(t *testing.T) {
	slog := &stubLogger{}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	h, err := RequestIDMiddleware(next, WithLogger(slog))
	require.NoError(t, err)
	require.NotNil(t, h)

	rr := doRequestWithHeader(t, h, http.MethodGet, "/x", DefaultHeaderName, "abc123")
	assert.Equal(t, http.StatusOK, rr.Code)
	entries := slog.entries()
	require.Len(t, entries, 1)
	assert.Contains(t, entries[0], "request_id=abc123")
	assert.Contains(t, entries[0], "method=GET")
	assert.Contains(t, entries[0], "path=/x")
}

func TestRequestID_CustomHeader(t *testing.T) {
	slog := &stubLogger{}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	h, err := RequestIDMiddleware(next, WithLogger(slog), WithHeaderName("X-Trace"))
	require.NoError(t, err)
	require.NotNil(t, h)

	_ = doRequestWithHeader(t, h, http.MethodPost, "/y", "X-Trace", "xyz789")
	entries := slog.entries()
	require.Len(t, entries, 1)
	assert.Contains(t, entries[0], "request_id=xyz789")
	assert.Contains(t, entries[0], "method=POST")
	assert.Contains(t, entries[0], "path=/y")
}

func TestRequestID_MissingHeaderSilent(t *testing.T) {
	slog := &stubLogger{}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	h, err := RequestIDMiddleware(next, WithLogger(slog))
	require.NoError(t, err)

	_ = doRequestWithHeader(t, h, http.MethodGet, "/z", "", "")
	assert.Len(t, slog.entries(), 0)
}

func TestRequestID_InvalidHeaderOption(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h, err := RequestIDMiddleware(next, WithHeaderName(""))
	require.Error(t, err)
	require.Nil(t, h)
}

func TestRequestID_NilLoggerOption(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h, err := RequestIDMiddleware(next, WithLogger(nil))
	require.Error(t, err)
	require.Nil(t, h)
}

func TestRequestID_NilNextPanics(t *testing.T) {
	assert.Panics(t, func() { _, _ = RequestIDMiddleware(nil) })
}

func TestRequestID_NoLoggerProvided(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusAccepted) })
	h, err := RequestIDMiddleware(next) // no logger option
	require.NoError(t, err)
	rr := doRequestWithHeader(t, h, http.MethodGet, "/n", DefaultHeaderName, "abc123")
	assert.Equal(t, http.StatusAccepted, rr.Code)
	// No logger so nothing to assert beyond status; ensure no panic.
}

func TestRequestID_Concurrency(t *testing.T) {
	slog := &stubLogger{}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h, err := RequestIDMiddleware(next, WithLogger(slog))
	require.NoError(t, err)

	const n = 10
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			doRequestWithHeader(t, h, http.MethodGet, "/c", DefaultHeaderName, "id-"+strconv.Itoa(i))
		}(i)
	}
	wg.Wait()
	entries := slog.entries()
	require.Len(t, entries, n)
	// Spot check a couple of entries.
	assert.Contains(t, entries[0], "request_id=id-")
}
