# requestmiddleware

Minimal pass-through HTTP middleware for Go.

## Overview

`requestmiddleware` provides a single function `middleware.Middleware` that wraps an `http.Handler` and forwards requests unchanged. It performs **no** logging, metrics, tracing, header mutation, context injection, panic recovery, or timing. This can be used as:

- A placeholder in a chain where conditional behavior may be added later.
- A reference implementation for creating custom middleware with zero overhead.

## Installation

Once published:

```
go get github.com/mytheresa/requestmiddleware
```

## Usage

```go
package main

import (
    "log"
    "net/http"

    "github.com/mytheresa/requestmiddleware"
    "github.com/mytheresa/requestmiddleware/middleware"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("ok"))
    })

    // Wrap with pass-through middleware.
    wrapped := middleware.Middleware(mux)

    log.Println("listening on :8080")
    if err := http.ListenAndServe(":8080", wrapped); err != nil {
        log.Fatal(err)
    }
}
```

## API

```go
func Middleware(next http.Handler) http.Handler
```

Behavior:

- Panics if `next` is `nil` (fail-fast programmer error).
- Simply calls `next.ServeHTTP(w, r)` with the original `ResponseWriter` and `*Request`.
- Performs zero additional allocations beyond what the standard library would already do.

## Design Decisions

| Decision               | Rationale                                                            |
| ---------------------- | -------------------------------------------------------------------- |
| Panic on nil `next`    | Surfaces configuration mistakes early rather than silently ignoring. |
| Single exported symbol | Keeps surface minimal and focused.                                   |
| No LICENSE initially   | Deferred until legal decision; easier to add later.                  |

## Testing

Unit tests verify preservation of:

- Status code
- Body
- Headers
- Panic on nil `next`

Run tests with race detector and coverage:

```
go test -race -count=1 -cover ./...
```

## Future (Out of Scope Now)

- Benchmarks comparing direct vs wrapped handler.
- Additional helper constructors (not currently needed).

## Rationale

Middleware stacks often grow; starting from the simplest possible implementation clarifies responsibility boundaries and avoids premature abstraction.

## No License Yet

The project currently omits a LICENSE file per initial requirements. Until a license is added, usage terms are implicit and may restrict adoption. Add one (e.g., MIT or Apache 2.0) when ready.
