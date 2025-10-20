---
mode: agent
---

Check if there's an [implementation plan](../../plan.md) and follow it closely. If there's none, rely on the current context and best practices to implement the required feature or refactoring.

Planning & execution:
• Capture an implementation plan using #todos (each todo: concrete, actionable, minimal).
• Break the feature/refactor into small #todos first; then implement them sequentially marking completion as you go.
• Keep todos updated—add, adjust, or remove as scope clarifies.

Feedback gate:
If no external plan exists and a large or risky change is inferred, ask the user for a quick confirmation before executing sweeping modifications. Prefer progressing with clearly bounded, reversible steps otherwise.

General rules:
• Follow idiomatic Go, repository instructions, and agreed patterns.
• Do not assume unstated requirements—clarify only when absolutely blocking.
• Prefer minimal, cohesive commits; avoid drive-by unrelated refactors.

External packages/libraries:
• Use #upstash/context7/* to gather up to date information about any external libraries usage before writing code that uses them.
• Include relevant usage examples in your implementation plan.

Kickoff phrase:
Before starting, output exactly: I'm going all in! — then begin execution.
