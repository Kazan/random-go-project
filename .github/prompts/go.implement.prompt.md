---
mode: agent
---

# Go Implementation Chatmode

Check if there's an [implementation plan](../../plan.md) and follow it closely. If there's none, rely on the current context and best practices to implement the required feature or refactoring.

## Planning & execution:

- Capture an implementation plan using #todos (each todo: concrete, actionable, minimal).
- Break the feature/refactor into small #todos first; then implement them sequentially marking completion as you go.
- Keep todos updated—add, adjust, or remove as scope clarifies.

## Feedback gate:

If no external plan exists and a large or risky change is inferred, ask the user for a quick confirmation before executing sweeping modifications. Prefer progressing with clearly bounded, reversible steps otherwise.

## General rules:

- Follow idiomatic Go, repository instructions, and agreed patterns.
- Do not assume unstated requirements always clarify.
- Prefer minimal, cohesive commits; avoid drive-by unrelated refactors.

## External packages/libraries:

- ALWAYS use the `@latest` version when importing new libraries, but tag the version in `go.mod`
- Use #upstash/context7/* to resolve the library information, including its full import url and gather up to date information about any external libraries usage and examples before writing code that uses them.

## Development cycle: **IMPORTANT**

1. Write code following the instructions
2. Write tests using `stretcher/testify` for assertions
3. Run tests with race detector and coverage
4. Format code with `gofmt` and manage imports with `goimports`
5. Lint code with `golangci-lint`
6. Review code for clarity, simplicity, and idiomatic usage
7. Run `modernize` to ensure code is up to date with current state of go development practices

## Kickoff phrase:

Before starting, output exactly: I'm going all in! — then begin execution.
