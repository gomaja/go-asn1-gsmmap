# Go Development Commands - Quick Reference

## Workflow Overview

```
/plan -> /create -> /test -> /check -> /audit -> /fix (if needed)
```

---

## Command Summary

| Command | Phase | Speed | When to Use |
|---------|-------|-------|-------------|
| `/plan` | Design | Fast | Before implementing any non-trivial feature |
| `/create` | Implement | Medium | When ready to write code |
| `/test` | Verify | Medium | After making changes |
| `/check` | Validate | Fast | During development, before commits |
| `/audit` | Review | Slow | Before releases, after major changes |
| `/fix` | Remediate | Varies | When something is broken |
| `/discuss` | Explore | Fast | Before committing to a design |

---

## /plan - Design Phase

**Purpose:** Design features before coding. Get approval first.

**Usage:**
```
/plan <feature description>
```

**What it does:**
1. Analyzes requirements
2. References all guidelines (data-safety, design-philosophy)
3. Creates implementation plan with packages, data flow, error handling
4. Presents plan for approval before any code is written

**Use when:** Starting any feature that creates new packages, affects multiple modules, or involves complex logic.

---

## /create - Implementation Phase

**Purpose:** Build features following all guidelines.

**Usage:**
```
/create <feature>           # Small features (inline planning)
/create from plan           # After /plan approval
/create <module>/<feature>  # Specify target module
```

**What it does:**
1. References guidelines before writing code
2. Uses Decimal (not float64) for money
3. Proper error handling (no ignored errors, no panics)
4. Logs state transitions with slog
5. Validates with goimports -> vet -> lint -> test -> build

**Critical rule:** For complex features, run `/plan` first!

---

## /test - Verification Phase

**Purpose:** Run tests, analyze failures, fix them.

**Usage:**
```
/test                 # All tests
/test <module>        # Specific module
/test <test_name>     # Specific test
```

**What it does:**
1. Runs tests with race detector
2. Analyzes failures (test bug vs code bug)
3. Fixes issues
4. Repeats until all pass
5. Runs lint check after fixes

**Failure categories:**
- Test Bug -> Fix the test
- Code Bug -> Fix the code
- Missing Implementation -> Implement it
- Race Condition -> Fix synchronization
- Flaky Test -> Fix timing dependency

---

## /check - Validation Phase (Quick)

**Purpose:** Fast feedback during development.

**Usage:**
```
/check              # Entire workspace
/check <module>     # Specific module
/check --fix        # Auto-fix issues
```

**What it does:**
1. Format check (`goimports -l .`)
2. Vet check (`go vet ./...`)
3. Lint check (`golangci-lint run ./...`)
4. Build check (`go build ./...`)

**Does NOT:** Run tests (use `/test` for that)

**Use when:** After making changes, before committing, for quick feedback.

---

## /audit - Review Phase (Full)

**Purpose:** Comprehensive production readiness review.

**Usage:**
```
/audit              # Full audit, report only
/audit --fix        # Full audit with fixes
/audit <module>     # Specific module
```

**What it does:**
1. Reads all guidelines
2. Runs complete quality suite
3. Production-critical checks:
   - float64/float32 for money (CRITICAL)
   - Ignored errors (CRITICAL)
   - Panics in production (CRITICAL)
   - Missing context propagation
   - Goroutine lifecycle
   - State transition logging
4. Generates detailed report with severity levels

**Severity levels:**
- Critical - Must fix before deploy
- Warning - Should fix soon
- Info - Nice to have

---

## /fix - Remediation Phase

**Purpose:** Fix specific issues.

**Usage:**
```
/fix <issue description>  # Fix a specific bug
/fix tests                # Fix failing tests
/fix lint                 # Fix lint warnings
/fix from audit           # Fix audit findings
/fix from check           # Fix check failures
```

**What it does:**
1. Identifies issue type
2. References guidelines for correct pattern
3. Applies minimal, targeted fix
4. Validates fix doesn't break anything
5. Reports what was changed

**Scope rule:** Fix only the issue, don't refactor surrounding code.

---

## /discuss - Exploration Phase

**Purpose:** Explore concepts and trade-offs before committing to a design.

**Usage:**
```
/discuss <topic or question>
```

**What it does:**
1. Analyzes current codebase patterns
2. Researches industry best practices
3. Presents 2-3 options with pros/cons
4. Recommends approach (if clear winner)

**Rules:** No code output, no files written. Conceptual only.

---

## Data Safety Rules (Always Apply)

| Rule | Bad | Good |
|------|-----|------|
| Money types | `float64` | `decimal.Decimal` |
| Error handling | `result, _ :=` | `if err != nil { return }` |
| Panics | `panic("bad")` | `return fmt.Errorf(...)` |
| State changes | Silent | Logged with `slog` |
| External calls | No context | `context.Context` with timeout |
| Goroutines | Bare `go func()` | `errgroup` lifecycle |
| Shutdown | Abrupt | `signal.NotifyContext` |

---

## Quick Decision Tree

```
Starting new work?
├── Complex feature -> /plan -> /create from plan
├── Exploring options -> /discuss -> /plan
└── Simple change -> /create

Made changes?
├── Just wrote code -> /check
├── Ready to verify -> /test
└── Ready to release -> /audit

Something broken?
├── Tests failing -> /fix tests
├── Lint warnings -> /fix lint
├── Audit findings -> /fix from audit
└── Specific bug -> /fix <description>
```

---

## Files Reference

| File | Purpose |
|------|---------|
| `CLAUDE.md` | Main project instructions |
| `AGENTS.md` | Universal AI agent guidelines |
| `.claude/commands/` | Command definitions |
| `.claude/skills/` | Coding patterns and checklists |
| `.reference/` | Cloned guideline repos |
| `.golangci.yml` | Linter configuration |
