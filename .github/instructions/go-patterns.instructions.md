---
description: "Instructions for writing agreed Go patterns and best practices"
applyTo: "**/*.go,**/go.mod,**/go.sum"
---

# Go Patterns Instructions

Follow the [Go instructions](./go.instructions.md) and these additional Go patterns and best practices when writing Go code.

## General Instructions

### Package initialization and configuration

Applies to reusable packages in a project, or when writing libraries.

**IMPORTANT**: Keep in mind forward and backward compatibility when designing and adding new configuration options.

Always ask the user the preferred way to configure the package, avoiding global state. The common patterns are:

1. Passing each configuration option as a parameter to the constructor function.

- This is only recommended for packages with few configuration options (2-3).
- Consider using any other approach to improve compatibility when adding new options in the future.

e.g.

```go
func New(name string, cli *http.Client) (*Type, error) {
    // ...
}
```

2. Via a config struct

- Create a `Config` struct with all configuration options
- Provide a `New(config Config) (*Type, error)` constructor function
- Assume we will use environment variables to populate the config struct via gotdotenv `env` annotations
- If an application has a centralized configuration management, use this struct but never define it that application package.

e.g.

```go
type Config struct {
    Name   string `env:"NAME"`
    Debug  bool   `env:"DEBUG,default=false"`
}

func New(config Config) (*Type, error) {
    // ...
}
```

3. Via functional options pattern (with `Option` type and `WithX` functions)

- prefer this approach for packages with many configuration options or when options are optional
- Define an `Option` type as a function that modifies the configuration
- Provide `WithX` functions for each configuration option
- Create a `New(opts ...Option) (*Type, error)` constructor function that applies the options
- Attempt to split the options into their own file in the package, e.g. `options.go` in order to improve readability

e.g.

```go

cfg struct {
    Name   string
    Debug  bool
    cli   *http.Client
}

type Option func(*cfg) error

func WithName(name string) Option {
    return func(c *cfg) error {
        if name == "" {
            return errors.New("name cannot be empty")
        }
        c.Name = name
        return nil
    }
}

func WithHTTPClient(cli *http.Client) Option {
    return func(c *cfg) error {
        c.cli = cli
        return nil
    }
}

func New(opts ...Option) (*Type, error) {
    config := &cfg{
        cli: http.DefaultClient,
    }

    for _, opt := range opts {
        if err := opt(config); err != nil {
            return nil, fmt.Errorf("applying option: %w", err)
        }
    }

    // ...
}
```
