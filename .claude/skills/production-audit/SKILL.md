---
name: production-audit
description: Full production readiness audit referencing Uber Go Style Guide, Google Go Style Guide, and production safety patterns. Use for comprehensive code review and fixes.
---

# Production Readiness Audit

## When to Use

- Before major releases or deployments
- When onboarding to a new codebase
- After significant refactoring
- When user says "audit", "production ready", "full check", "review code quality"
- Periodically for codebase health checks

## Reference Materials

This audit references these guidelines (in `.reference/`):

| Guideline | Location | Focus Areas |
|-----------|----------|-------------|
| Uber Go Style Guide | `.reference/uber-go-guide/` | Error handling, concurrency, performance |
| Google Go Style Guide | Online reference | Naming, documentation, best practices |
| Data Safety | `.claude/skills/data-safety/` | No floats, no panics, logging |
| Design Philosophy | `.claude/skills/design-philosophy/` | Cognitive load, interfaces, packages |

## Audit Checklist

### 1. Critical Issues (MUST FIX)

| Issue | Search Pattern | Severity |
|-------|---------------|----------|
| Floating point for money | `grep -rn "float64\|float32" --include="*.go"` in financial paths | CRITICAL |
| Ignored errors | `grep -rn "_, _\|, _" --include="*.go" \| grep -v test` | CRITICAL |
| Panic in production | `grep -rn "panic(" --include="*.go" \| grep -v test` | CRITICAL |
| log.Fatal in production | `grep -rn "log.Fatal" --include="*.go" \| grep -v test\|main.go` | HIGH |
| Missing context param | Functions doing I/O without `context.Context` | HIGH |

### 2. Code Quality Issues

| Issue | How to Check | Severity |
|-------|--------------|----------|
| golangci-lint warnings | `golangci-lint run ./...` | HIGH |
| go vet issues | `go vet ./...` | HIGH |
| Race conditions | `go test -race ./...` | HIGH |
| Formatting issues | `goimports -l .` | MEDIUM |
| Large functions | Functions > 60 lines | MEDIUM |

### 3. Concurrency Issues

| Issue | What to Check | Severity |
|-------|---------------|----------|
| No context propagation | Goroutines without ctx parameter | CRITICAL |
| Goroutine leak | Goroutines without lifecycle management | CRITICAL |
| Missing errgroup | Multiple goroutines without error collection | HIGH |
| Shared state without mutex | Struct fields accessed from multiple goroutines | CRITICAL |
| Unbounded goroutines | `go func()` without limit/pool | HIGH |

### 4. API Design (per Uber/Google Go Style Guides)

| Guideline | Check |
|-----------|-------|
| Accept interfaces, return structs | Function signatures follow this pattern |
| Small interfaces | Interfaces have 1-3 methods |
| Error wrapping | Errors use `fmt.Errorf("context: %w", err)` |
| Context first | `context.Context` is first parameter |
| Options pattern | Complex constructors use functional options |

### 5. Documentation

| Requirement | Check |
|-------------|-------|
| Package docs | Each package has a doc comment on the `package` line |
| Exported docs | All exported types, functions, methods have comments |
| Examples | Complex APIs have `Example` test functions |
| Error docs | Error conditions are documented |

## Fix Patterns

### Replacing float64 with Decimal

```go
// Before
price := msg.Price // float64
total := price * quantity

// After
import "github.com/shopspring/decimal"

price, err := decimal.NewFromString(msg.Price)
if err != nil {
    return fmt.Errorf("parsing price: %w", err)
}
total := price.Mul(quantity)
```

### Replacing Ignored Errors

```go
// Before
data, _ := json.Marshal(obj)
conn.Write(data) // error ignored

// After
data, err := json.Marshal(obj)
if err != nil {
    return fmt.Errorf("marshaling object: %w", err)
}
if _, err := conn.Write(data); err != nil {
    return fmt.Errorf("writing to connection: %w", err)
}
```

### Adding State Transition Logging

```go
// Before
s.status = StatusActive

// After
oldStatus := s.status
s.status = StatusActive
slog.Info("status changed",
    "old_status", oldStatus,
    "new_status", s.status,
    "reason", reason,
)
```

### Adding Context to Functions

```go
// Before
func (s *Service) Process(data []byte) error { ... }

// After
func (s *Service) Process(ctx context.Context, data []byte) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    // ... process data
}
```

## Verification Commands

After all fixes:

```bash
# Must all pass with zero warnings/errors
goimports -l .              # Should produce no output
go vet ./...
golangci-lint run ./...
go test -race ./...
go build -o go-dsr .
```

## Related Commands

- `/audit` - Invokes this full audit
- `/check` - Quick quality check (subset of this)
- `/lint` - Just linting
