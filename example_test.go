package middler_test

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"

	middler "github.com/ivan-nunez/random-go-project"
)

// ExampleNew demonstrates constructing and applying the middleware. It logs a header
// when present and remains silent when absent. The output is not asserted here, but
// the example shows typical usage with ServeMux.
func ExampleNew() {
	// Use a logger that discards output to keep example output deterministic.
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mw, err := middler.New(
		middler.WithHeader("myth-trace-id"),
		middler.WithLogger(logger),
	)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})))

	// Simulate a request with the header.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("myth-trace-id", "abc123")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	fmt.Println("middleware invoked")

	// Output:
	// middleware invoked
}
