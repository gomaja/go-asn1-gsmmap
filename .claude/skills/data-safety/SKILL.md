---
name: data-safety
description: Production data safety patterns and validation for Go. Use when writing or reviewing code that handles important data.
---

# Data Safety Patterns

## When to Use

- Writing any code that handles prices or financial data
- Implementing data processing logic
- Working with data streams or external APIs
- Reviewing data pipeline code
- Before deploying data services

## Critical Rules

### 1. Never Use Floating Point for Money

```go
// BAD - NEVER DO THIS
price := 1234.567
total := price * quantity

// GOOD - ALWAYS DO THIS
import "github.com/shopspring/decimal"

price, err := decimal.NewFromString("1234.567")
if err != nil {
    return fmt.Errorf("parsing price: %w", err)
}
total := price.Mul(quantity)
```

### 2. No Ignored Errors in Production

```go
// BAD - NEVER
result, _ := doSomething()
json.Unmarshal(data, &obj) // error ignored

// GOOD - ALWAYS
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doing something: %w", err)
}

if err := json.Unmarshal(data, &obj); err != nil {
    return fmt.Errorf("unmarshaling response: %w", err)
}
```

### 3. No Panics in Production Code

```go
// BAD - NEVER in production paths
panic("unexpected state")
log.Fatal("cannot continue") // calls os.Exit

// GOOD - ALWAYS return errors
return fmt.Errorf("unexpected state: %s", state)
return &StateError{State: state, Reason: "invalid transition"}
```

### 4. Validate External Data

```go
func ProcessData(raw []byte) (*Data, error) {
    var data RawData
    if err := json.Unmarshal(raw, &data); err != nil {
        return nil, fmt.Errorf("parsing data: %w", err)
    }

    // Validate before using
    if data.Price.IsNegative() || data.Price.IsZero() {
        return nil, fmt.Errorf("invalid price: %s", data.Price)
    }
    if data.Quantity.IsNegative() {
        return nil, fmt.Errorf("invalid quantity: %s", data.Quantity)
    }
    if data.Timestamp.IsZero() {
        return nil, errors.New("missing timestamp")
    }

    return data.ToValidated(), nil
}
```

### 5. Log State Transitions

```go
func (s *Service) ProcessEvent(ctx context.Context, event *Event) error {
    slog.InfoContext(ctx, "processing event",
        "event_id", event.ID,
        "type", event.Type,
        "timestamp", event.Timestamp,
    )

    result, err := s.writer.Write(ctx, event)

    if err != nil {
        slog.ErrorContext(ctx, "event write failed",
            "event_id", event.ID,
            "error", err,
        )
        return fmt.Errorf("writing event %s: %w", event.ID, err)
    }

    slog.DebugContext(ctx, "event written",
        "event_id", event.ID,
        "rows_affected", result.RowsAffected,
    )
    return nil
}
```

### 6. Context Propagation and Timeouts

```go
// ALWAYS pass context, ALWAYS set timeouts for external calls
func (c *Client) FetchData(ctx context.Context, id string) (*Data, error) {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", c.url+"/data/"+id, nil)
    if err != nil {
        return nil, fmt.Errorf("creating request: %w", err)
    }

    resp, err := c.http.Do(req)
    if err != nil {
        return nil, fmt.Errorf("fetching data %s: %w", id, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status %d for data %s", resp.StatusCode, id)
    }

    var data Data
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return nil, fmt.Errorf("decoding data %s: %w", id, err)
    }
    return &data, nil
}
```

### 7. Reconnection with Exponential Backoff

```go
const maxBackoff = 60 * time.Second

func (s *Service) RunConnectionLoop(ctx context.Context) error {
    backoff := time.Second

    for {
        err := s.connectAndProcess(ctx)
        if ctx.Err() != nil {
            return ctx.Err() // Context cancelled, shutdown
        }

        if err != nil {
            slog.Error("connection error",
                "error", err,
                "backoff", backoff,
            )
        } else {
            slog.Info("connection closed normally")
            backoff = time.Second // Reset on clean close
        }

        select {
        case <-time.After(backoff):
            backoff = min(backoff*2, maxBackoff)
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}
```

### 8. Graceful Shutdown

```go
func Run(ctx context.Context) error {
    g, ctx := errgroup.WithContext(ctx)

    g.Go(func() error {
        return runServer(ctx)
    })

    g.Go(func() error {
        return runWorker(ctx)
    })

    // Wait for signal
    g.Go(func() error {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

        select {
        case sig := <-sigCh:
            slog.Info("shutdown signal received", "signal", sig)
            return fmt.Errorf("received signal: %v", sig)
        case <-ctx.Done():
            return ctx.Err()
        }
    })

    return g.Wait()
}
```

## Pre-Commit Checklist

Before committing code:

- [ ] No `float64`/`float32` for financial calculations
- [ ] No ignored errors (`_` on error returns) in production paths
- [ ] No `panic()` or `log.Fatal()` in production paths
- [ ] All external API errors handled with retry/timeout
- [ ] Data validation at boundaries
- [ ] All state transitions logged with `slog`
- [ ] Graceful shutdown handling via context
- [ ] Reconnection logic with exponential backoff (if applicable)
- [ ] Context propagation to all downstream calls

## Error Types

Define domain-specific errors:

```go
import (
    "errors"
    "fmt"
)

// Sentinel errors
var (
    ErrNotFound      = errors.New("not found")
    ErrAlreadyExists = errors.New("already exists")
    ErrCircuitOpen   = errors.New("circuit breaker open")
)

// Structured errors with context
type ConnectionError struct {
    Host    string
    Port    int
    Cause   error
}

func (e *ConnectionError) Error() string {
    return fmt.Sprintf("connection to %s:%d failed: %v", e.Host, e.Port, e.Cause)
}

func (e *ConnectionError) Unwrap() error {
    return e.Cause
}

// Usage with errors.Is / errors.As
if errors.Is(err, ErrNotFound) { ... }

var connErr *ConnectionError
if errors.As(err, &connErr) {
    slog.Error("connection failed", "host", connErr.Host)
}
```

## Related Skills

- `check-code-quality` - Full quality checklist
- `run-lint` - Linting
