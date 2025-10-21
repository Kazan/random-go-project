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
- ALWAYS use #upstash/context7/\* to resolve the library information, including its full import url and gather up to date information about any external libraries usage and examples before writing code that uses them.

## IMPLEMENT ACTIONS (Deterministic Execution)

When the user asks to "implement" a feature/refactor, perform the following exact ordered steps before concluding the task. Treat each as mandatory.

1. Plan
   - If a plan (`plan.md`) exists: load and convert it into a todo list. Otherwise create a concise initial todo list.
   - Mark only one todo in-progress at a time.
2. Context gathering
   - Read all directly related files BEFORE editing.
   - Read instruction files relevant to Go if not already read in session.
3. Code implementation
   - Apply minimal cohesive patches; avoid unrelated changes.
   - Keep public surface minimal and idiomatic.
4. Dependencies
   - For any new dependency: `go get <module>@latest`, then `go mod tidy`, then `go mod vendor` (if vendor directory is used).
5. Formatting
   - Run `gofmt` and `goimports` (or ensure editor tooling did so) after each substantial change batch.
6. Tests
   - Add/modify table-driven tests using `stretchr/testify`.
   - Run: `go test -race -cover ./...` and capture coverage.
   - Fix failures immediately; iterate until green.
7. Lint
   - Run `golangci-lint run` with repo config.
   - Address all errors; may ignore documented false positives (state rationale).
8. Modernize
   - Run: `modernize` including the directories being modified (e.g. `./...`, `pkg/<package_name>`, etc) and capture suggestions.
   - Review for usage of latest Go features applicable (e.g., `sync.WaitGroup.Go` if Go >=1.25, `errors.Join`, new stdlib APIs).
9. Diagnostics
   - Re-run `golangci-lint run`, `go vet ./...` and `modernize` against the modified code for final pass.
10. Repetition
    - Continue iterating until all todos are complete, don't skip the running of the advised tools.
11. Summary output
    - Provide PASS/FAIL for: Build, Lint, Tests.
    - List changed files with one-line purpose.
    - Note any skipped/omitted steps and why.
    - Provide next-step suggestions (CI, docs, examples) when value-add is low risk.

Kickoff Reminder: Always start implementation responses with exactly: `I'm going all in!` (verbatim) before performing actions.
