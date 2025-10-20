package middler

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// memoryHandler is a simple slog.Handler capturing records for inspection.
type memoryHandler struct {
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
	m.records = append(m.records, clone)
	return nil
}
func (m *memoryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &memoryHandler{records: m.records}
}
func (m *memoryHandler) WithGroup(name string) slog.Handler {
	return &memoryHandler{records: m.records}
}

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
