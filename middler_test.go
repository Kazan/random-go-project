package middler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// memoryHandler is a simple slog.Handler capturing records for inspection.
type memoryHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (m *memoryHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (m *memoryHandler) Handle(_ context.Context, r slog.Record) error {
	// Copy record to avoid mutation issues (attributes iterator is consumed on use).
	clone := slog.Record{Time: r.Time, Level: r.Level, Message: r.Message}
	r.Attrs(func(a slog.Attr) bool {
		clone.AddAttrs(a)
		return true
	})
	m.mu.Lock()
	m.records = append(m.records, clone)
	m.mu.Unlock()
	return nil
}
func (m *memoryHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return m }
func (m *memoryHandler) WithGroup(name string) slog.Handler       { return m }

func TestNew_ErrorMissingHeader(t *testing.T) {
	mw, err := New() // no header option
	require.Error(t, err)
	assert.Nil(t, mw)
}

func TestNew_SuccessFallbackLogger(t *testing.T) {
	mw, err := New(WithHeader("myth-trace-id")) // nil logger allowed
	require.NoError(t, err)
	require.NotNil(t, mw)
}

func TestNew_ErrorDuplicateHeaderOption(t *testing.T) {
	mw, err := New(WithHeader("trace-id"), WithHeader("other"))
	require.Error(t, err)
	assert.Nil(t, mw)
}

func TestNew_HeaderTrimming(t *testing.T) {
	mh := &memoryHandler{}
	logger := slog.New(mh)
	mw, err := New(WithHeader("  X-Trace-ID  "), WithLogger(logger))
	require.NoError(t, err)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Trace-ID", "val")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Len(t, mh.records, 1)
	logged := false
	mh.records[0].Attrs(func(a slog.Attr) bool {
		if a.Key == "header.x_trace_id" && a.Value.String() == "val" {
			logged = true
		}
		return true
	})
	assert.True(t, logged, "trimmed header name should be normalized and logged")
}

func TestMiddleware_MultiValueHeader_LogsFirst(t *testing.T) {
	mh := &memoryHandler{}
	logger := slog.New(mh)
	mw, err := New(WithHeader("multi-value"), WithLogger(logger))
	require.NoError(t, err)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("multi-value", "first")
	req.Header.Add("multi-value", "second")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Len(t, mh.records, 1)
	logged := false
	mh.records[0].Attrs(func(a slog.Attr) bool {
		if a.Key == "header.multi_value" && a.Value.String() == "first" {
			logged = true
		}
		return true
	})
	assert.True(t, logged, "should log first value returned by Header.Get")
}

func TestMiddleware_EmptyValue_NoLog(t *testing.T) {
	mh := &memoryHandler{}
	logger := slog.New(mh)
	mw, err := New(WithHeader("maybe-empty"), WithLogger(logger))
	require.NoError(t, err)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("maybe-empty", "") // present but empty
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	assert.Len(t, mh.records, 0)
}

func TestMiddleware_LongValue(t *testing.T) {
	mh := &memoryHandler{}
	logger := slog.New(mh)
	mw, err := New(WithHeader("long"), WithLogger(logger))
	require.NoError(t, err)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	longVal := strings.Repeat("x", 1024)
	req.Header.Set("long", longVal)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Len(t, mh.records, 1)
	found := false
	mh.records[0].Attrs(func(a slog.Attr) bool {
		if a.Key == "header.long" && a.Value.String() == longVal {
			found = true
		}
		return true
	})
	assert.True(t, found)
}

func TestMiddleware_Concurrency(t *testing.T) {
	mh := &memoryHandler{}
	logger := slog.New(mh)
	mw, err := New(WithHeader("cid"), WithLogger(logger))
	require.NoError(t, err)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	const workers = 50
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("cid", fmt.Sprintf("val-%d", i))
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
		}(i)
	}
	wg.Wait()
	mh.mu.Lock()
	count := len(mh.records)
	mh.mu.Unlock()
	assert.Equal(t, workers, count, "each request should log once")
}

func TestMiddleware_HeaderPresent(t *testing.T) {
	mh := &memoryHandler{}
	logger := slog.New(mh)
	mw, err := New(WithHeader("myth-trace-id"), WithLogger(logger))
	require.NoError(t, err)

	// build handler chain
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("myth-trace-id", "abc123")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Len(t, mh.records, 1)
	rec := mh.records[0]
	assert.Equal(t, slog.LevelInfo, rec.Level)
	assert.Equal(t, "request header", rec.Message)
	// find attribute
	found := false
	rec.Attrs(func(a slog.Attr) bool {
		if a.Key == "header.myth_trace_id" && a.Value.String() == "abc123" {
			found = true
		}
		return true
	})
	assert.True(t, found, "expected header attribute logged")
}

func TestMiddleware_HeaderAbsent_NoLog(t *testing.T) {
	mh := &memoryHandler{}
	logger := slog.New(mh)
	mw, err := New(WithHeader("myth-trace-id"), WithLogger(logger))
	require.NoError(t, err)

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Code)
	assert.Len(t, mh.records, 0, "should not log when header absent")
}

func TestMiddleware_CustomLevel(t *testing.T) {
	mh := &memoryHandler{}
	logger := slog.New(mh)
	mw, err := New(WithHeader("myth-trace-id"), WithLogger(logger), WithLevel(slog.LevelWarn))
	require.NoError(t, err)

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("myth-trace-id", "xyz")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Len(t, mh.records, 1)
	rec := mh.records[0]
	assert.Equal(t, slog.LevelWarn, rec.Level)
}

func TestNew_ErrorWhitespaceHeader(t *testing.T) {
	mw, err := New(WithHeader("   \t  \n"))
	require.Error(t, err)
	assert.Nil(t, mw)
}

func TestNew_DuplicateSameHeaderAllowed(t *testing.T) {
	// Using the exact same trimmed header twice should not error (current logic errors only if different second value)
	mh := &memoryHandler{}
	logger := slog.New(mh)
	mw, err := New(WithHeader("trace-id"), WithHeader("trace-id"), WithLogger(logger))
	require.NoError(t, err)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("trace-id", "v")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Len(t, mh.records, 1)
}

func TestMiddleware_NilLoggerFallback(t *testing.T) {
	// Temporarily set a custom default logger to verify fallback usage.
	mh := &memoryHandler{}
	defaultLogger := slog.New(mh)
	old := slog.Default()
	slog.SetDefault(defaultLogger)
	defer slog.SetDefault(old)

	mw, err := New(WithHeader("fallback"), WithLogger(nil))
	require.NoError(t, err)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("fallback", "x")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	mh.mu.Lock()
	count := len(mh.records)
	mh.mu.Unlock()
	assert.Equal(t, 1, count, "fallback default logger should log header")
}

func TestMiddleware_NegativeLogLevel(t *testing.T) {
	mh := &memoryHandler{}
	logger := slog.New(mh)
	// Use a negative level to ensure we don't restrict arbitrary slog.Level values.
	customLevel := slog.Level(-4)
	mw, err := New(WithHeader("neg"), WithLogger(logger), WithLevel(customLevel))
	require.NoError(t, err)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("neg", "val")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Len(t, mh.records, 1)
	assert.Equal(t, customLevel, mh.records[0].Level)
}
