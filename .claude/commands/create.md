---
name: create
description: Implement a feature following all guidelines. Can use a prior plan or work independently for small tasks.
---

# Feature Implementation

Implement a feature following all guidelines, patterns, and best practices. Always validates at the end.

## Input

```
/create <feature description>
/create from plan
```

## Instructions

You are implementing a feature for a Go production system. This command ALWAYS validates the code before completing.

### Phase 1: CHECK FOR PLAN

Determine the approach:

1. **If "from plan" specified**:
   - Look for the most recent `/plan` output in this conversation
   - Use that design as the blueprint
   - If no plan found, ask user to run `/plan` first

2. **If feature description provided**:
   - Assess scope:
     - **Small** (helper function, bug fix, simple addition): Proceed with inline design
     - **Medium** (new package, new type): Do quick inline design, confirm with user
     - **Large** (new service, major feature): Suggest running `/plan` first

3. **Scope indicators**:
   - Small: "add a function", "update", "modify", "fix"
   - Medium: "create a package", "add middleware", "implement X"
   - Large: "new service", "new module", "major refactor"

### Phase 2: REFERENCE (Always)

Before writing ANY code, read these:

1. **Data Safety** - `.claude/skills/data-safety/SKILL.md`
   - CRITICAL patterns that MUST be followed

2. **Design Philosophy** - `.claude/skills/design-philosophy/SKILL.md`
   - Core principles for good design

3. **Existing Patterns** - Check relevant code in `internal/`:
   ```bash
   # Find similar implementations
   ls internal/
   ```

### Phase 3: IMPLEMENT

Write code following these MANDATORY patterns:

#### 3.1 Financial Calculations (CRITICAL)

```go
// NEVER
price := 123.45

// ALWAYS
import "github.com/shopspring/decimal"

price, err := decimal.NewFromString("123.45")
if err != nil {
    return fmt.Errorf("parsing price: %w", err)
}
```

#### 3.2 Error Handling (CRITICAL)

```go
// NEVER
result, _ := doSomething()
panic("unexpected")

// ALWAYS
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doing something: %w", err)
}
```

Define errors:

```go
var ErrNotFound = errors.New("not found")

type ProcessError struct {
    ID  string
    Err error
}

func (e *ProcessError) Error() string {
    return fmt.Sprintf("processing %s: %v", e.ID, e.Err)
}

func (e *ProcessError) Unwrap() error {
    return e.Err
}
```

#### 3.3 State Transition Logging (CRITICAL for production systems)

```go
slog.Info("state changed",
    "old_value", old,
    "new_value", new,
    "reason", reason,
)
```

#### 3.4 Concurrency Patterns

**Graceful Shutdown:**
```go
func Run(ctx context.Context) error {
    g, ctx := errgroup.WithContext(ctx)

    g.Go(func() error {
        return runServer(ctx)
    })

    g.Go(func() error {
        return runWorker(ctx)
    })

    return g.Wait()
}
```

**Reconnection with Backoff:**
```go
backoff := time.Second
for {
    if err := connect(ctx); err != nil {
        slog.Error("connection failed", "error", err)
    } else {
        backoff = time.Second // Reset on success
    }

    select {
    case <-time.After(backoff):
        backoff = min(backoff*2, 60*time.Second)
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

#### 3.5 Package Organization

```go
// internal/peer/peer.go - Diameter peer management
package peer

// internal/routing/routing.go - Message routing logic
package routing

// internal/transaction/transaction.go - Transaction state
package transaction
```

#### 3.6 Documentation

Add doc comments to all exported items:

```go
// ProcessOrder validates and processes a new order.
//
// It returns a [ProcessError] if the order fails validation
// or if the downstream service is unavailable.
func ProcessOrder(ctx context.Context, order *Order) (*Result, error) { ... }
```

#### 3.7 Tests

Write tests alongside implementation:

```go
func TestProcessOrder(t *testing.T) {
    tests := []struct {
        name    string
        order   *Order
        want    *Result
        wantErr bool
    }{
        {
            name:  "valid order",
            order: &Order{ID: "123", Price: decimal.NewFromInt(100)},
            want:  &Result{Status: "accepted"},
        },
        {
            name:    "empty order ID",
            order:   &Order{ID: "", Price: decimal.NewFromInt(100)},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ProcessOrder(context.Background(), tt.order)
            if (err != nil) != tt.wantErr {
                t.Errorf("ProcessOrder() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr {
                assert.Equal(t, tt.want.Status, got.Status)
            }
        })
    }
}
```

### Phase 4: VALIDATE (Always - MANDATORY)

After implementation, run ALL of these:

```bash
# 1. Format
goimports -w .

# 2. Vet
go vet ./...

# 3. Lint (must pass with zero issues)
golangci-lint run ./...

# 4. Test with race detector
go test -race ./...

# 5. Build
go build -o go-dsr .
```

**If any check fails:**
1. Fix the issue
2. Re-run ALL checks
3. Repeat until all pass

**DO NOT complete this command until all checks pass.**

### Phase 5: REPORT

After successful validation, report:

```markdown
## Created: <Feature Name>

### Files Added/Modified
- `path/to/file.go` - Description

### Summary
Brief description of what was implemented.

### Tests Added
- `TestName` - What it tests

### Validation
- goimports - passed
- go vet - passed
- golangci-lint - passed (0 issues)
- go test -race - passed (N tests)
- go build - passed

### Usage Example
```go
// How to use the new feature
```
```

## Checklist Before Completion

- [ ] No `float64`/`float32` for prices or quantities
- [ ] No ignored errors in production code
- [ ] No `panic()` in production code
- [ ] Error types use sentinel values or structured types with `Unwrap()`
- [ ] State transitions are logged with `slog`
- [ ] Goroutines use `errgroup` for lifecycle management
- [ ] Context propagation to all downstream calls
- [ ] All exported items have doc comments
- [ ] Tests cover happy path and error cases (table-driven)
- [ ] All validation checks pass
