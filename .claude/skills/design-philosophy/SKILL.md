---
name: design-philosophy
description: Core design principles for Go production systems. Apply when designing APIs, packages, or data structures.
---

# Design Philosophy

## When to Use

- Designing new APIs, packages, or data structures
- Refactoring existing code
- Reviewing code for maintainability
- Making architectural decisions

## Core Principles

### 1. Minimize Cognitive Load

Code should be easy to understand without loading too much into working memory.

**Guidelines:**
- Each package has a single, clear responsibility
- Limit the number of concepts a reader must hold simultaneously
- Good separation of concerns means fewer "things" to keep in mind

```go
// BAD: Too many parameters, too much to track
func Process(data []byte, offset, length int, flags uint32, mode byte, retries int) ([]byte, error)

// GOOD: Group related concepts
type ProcessOptions struct {
    Flags   uint32
    Mode    byte
    Retries int
}

func Process(input []byte, opts ProcessOptions) ([]byte, error)
```

### 2. Accept Interfaces, Return Structs

Keep dependencies flexible and outputs concrete.

**Guidelines:**
- Function parameters should accept interfaces for flexibility
- Return concrete types so callers know exactly what they get
- Define interfaces where they are used, not where they are implemented

```go
// BAD: Accepting concrete type, returning interface
func NewProcessor(db *PostgresDB) DataProcessor { ... }

// GOOD: Accept interface, return struct
type Store interface {
    Get(ctx context.Context, id string) ([]byte, error)
}

func NewProcessor(store Store) *Processor { ... }
```

### 3. Make Zero Values Useful

Design types so their zero value is a valid, usable state.

**Guidelines:**
- A zero-value struct should be ready to use (or clearly require construction)
- Use pointer fields only when nil is a meaningful distinct state
- Prefer `sync.Mutex{}` over `*sync.Mutex` (zero value works)

```go
// GOOD: Zero value is useful
type Buffer struct {
    mu   sync.Mutex
    data []byte
}

// buf := Buffer{} is ready to use

// GOOD: When construction is required, enforce it
type Client struct {
    baseURL string // unexported, must use constructor
}

func NewClient(baseURL string) (*Client, error) {
    if baseURL == "" {
        return nil, errors.New("base URL required")
    }
    return &Client{baseURL: baseURL}, nil
}
```

### 4. Abstractions Must Earn Their Keep

An abstraction should reduce cognitive load, not add to it.

**Guidelines:**
- If understanding the abstraction requires more effort than concrete code, don't abstract
- Good abstractions match mental models developers already have
- Three similar lines of code is often better than a premature abstraction
- Define small interfaces (1-3 methods) close to where they're consumed

```go
// BAD: Abstraction adds complexity, used exactly once
type Processor[T Input, U Output] interface {
    Process(ctx context.Context, in T) (U, error)
}

// GOOD: Concrete until proven otherwise
func ProcessMarketData(ctx context.Context, raw []byte) (*MarketData, error) { ... }
// Abstract only when you have 2+ concrete use cases
```

### 5. Keep Packages Focused

**Guidelines:**
- Package names are single lowercase words (no underscores, no mixedCase)
- Package name should describe what it provides, not what it contains
- Avoid `util`, `common`, `helpers` - they are junk drawers
- A package with more than ~500 lines of production code should be evaluated for splitting

```go
// BAD: Grab-bag package
package utils

func FormatPrice(d decimal.Decimal) string { ... }
func ValidateEmail(s string) error { ... }
func RetryWithBackoff(fn func() error) error { ... }

// GOOD: Focused packages
package pricing  // FormatPrice lives here
package validate // ValidateEmail lives here
package retry    // RetryWithBackoff lives here
```

## Production-Specific Principles

### 6. Correctness Over Speed

A slow correct system beats a fast incorrect one.

```go
// BAD: Fast but potentially wrong
total := price * qty // float64 precision loss

// GOOD: Correct, performance is secondary
total, err := price.Mul(qty) // decimal.Decimal
if err != nil {
    return fmt.Errorf("calculating total: %w", err)
}
```

### 7. Explicit Error Handling

Every failure mode should be handled explicitly.

```go
// BAD: Silent failure
_ = sendNotification(ctx, msg)

// GOOD: Explicit handling
if err := sendNotification(ctx, msg); err != nil {
    slog.Error("notification failed",
        "error", err,
        "msg_id", msg.ID,
    )
    return fmt.Errorf("sending notification: %w", err)
}
```

### 8. Audit Everything

Log all state transitions for debugging and compliance.

```go
// BAD: Silent state change
s.count += delta

// GOOD: Auditable state change
oldCount := s.count
s.count += delta
slog.Info("counter updated",
    "old_value", oldCount,
    "new_value", s.count,
    "delta", delta,
    "reason", reason,
)
```

## Quick Reference

| Principle | Ask Yourself |
|-----------|--------------|
| Cognitive Load | "How many concepts must a reader hold to understand this?" |
| Interfaces | "Am I accepting interfaces and returning structs?" |
| Zero Values | "Is the zero value of this type useful?" |
| Abstraction Worth | "Does this make the code easier or harder to understand?" |
| Package Focus | "Does this package do one thing well?" |
| Correctness | "Is this correct? Speed is secondary." |
| Error Handling | "What happens when this fails?" |
| Audit Trail | "Can we trace what happened after the fact?" |
