---
name: test
description: Run tests, analyze failures, and fix them. Repeats until all tests pass.
---

# Test Workflow

Run tests, analyze any failures, fix them, and verify all tests pass.

## Input

```
/test                    # Run all tests in workspace
/test <module>           # Run tests for specific module
/test <test_name>        # Run specific test
```

## Instructions

You are running and fixing tests for a Go production system.

### Phase 1: RUN TESTS

Execute tests based on input:

```bash
# All tests with race detector
go test -race ./...

# Specific package
go test -race ./internal/routing/...

# Specific test function
go test -race ./internal/routing/... -run TestRealmRouting

# With verbose output (for debugging)
go test -race -v ./...

# With coverage
go test -race -cover ./...
```

Capture the output for analysis.

### Phase 2: ANALYZE FAILURES

If tests fail, analyze each failure:

#### 2.1 Parse Test Output

Look for:
- Test name that failed
- Assertion that failed
- Expected vs actual values
- Panic message (if any)
- File and line number

#### 2.2 Categorize Failure

| Type | Indicator | Action |
|------|-----------|--------|
| **Test Bug** | Test logic is wrong | Fix the test |
| **Code Bug** | Code doesn't match spec | Fix the code |
| **Missing Implementation** | Not yet implemented | Implement it |
| **Race Condition** | `-race` flag catches data race | Fix synchronization |
| **Flaky Test** | Passes sometimes | Fix timing dependency |
| **Environment Issue** | Missing setup | Fix test setup |

#### 2.3 Locate Root Cause

1. Read the failing test code
2. Read the code being tested
3. Understand what the test expects
4. Identify the discrepancy

### Phase 3: FIX

Apply fixes based on failure type:

#### 3.1 Fixing Test Bugs

```go
// Before: Wrong assertion
assert.Equal(t, 100, result) // But result should be 99

// After: Correct assertion
assert.Equal(t, 99, result)
```

#### 3.2 Fixing Code Bugs

Reference guidelines when fixing:
- Read `.claude/skills/data-safety/SKILL.md` for data safety patterns
- Ensure fix follows project conventions
- Don't introduce new issues

#### 3.3 Fixing Race Conditions

```go
// Before: Data race on shared state
type Counter struct {
    count int
}

func (c *Counter) Inc() { c.count++ }

// After: Thread-safe with mutex
type Counter struct {
    mu    sync.Mutex
    count int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}
```

#### 3.4 Fixing Flaky Tests

Common causes and fixes:

```go
// Timing dependency - use channels for synchronization
done := make(chan struct{})
go func() {
    defer close(done)
    // work
}()
<-done // Wait for completion instead of time.Sleep

// Order dependency - make tests independent
func TestIndependent(t *testing.T) {
    svc := newTestService(t) // Fresh state per test
}
```

### Phase 4: VERIFY

After fixing, re-run tests:

```bash
# Re-run the specific failing test first
go test -race ./... -run TestFailingName

# Then run all tests
go test -race ./...
```

**Repeat Phase 2-4 until all tests pass.**

### Phase 5: ENSURE CODE QUALITY

After tests pass, verify fixes didn't break anything:

```bash
# Quick lint check
golangci-lint run ./...

# Build check
go build ./...
```

Fix any new issues introduced.

### Phase 6: REPORT

After all tests pass:

```markdown
## Test Results: All Passing

### Initial Failures
- `TestName1` - Description of failure
- `TestName2` - Description of failure

### Fixes Applied
1. **TestName1**: Fixed [test/code] - Description of fix
2. **TestName2**: Fixed [test/code] - Description of fix

### Final Results
ok      module/package1    0.042s
ok      module/package2    0.031s
...

### Files Modified
- `path/to/file.go` - What was changed
```

## Common Test Patterns

### Table-Driven Tests

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   Input
        want    Output
        wantErr bool
    }{
        {
            name:  "happy path",
            input: validInput,
            want:  expectedOutput,
        },
        {
            name:    "error case",
            input:   invalidInput,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr {
                assert.Equal(t, tt.want, got)
            }
        })
    }
}
```

### Test with Context

```go
func TestWithContext(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    result, err := Function(ctx)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Test Helpers

```go
func newTestService(t *testing.T) *Service {
    t.Helper()
    svc, err := New(testConfig())
    if err != nil {
        t.Fatalf("creating test service: %v", err)
    }
    t.Cleanup(func() {
        svc.Close()
    })
    return svc
}
```

## When to Ask User

- **Test vs Code**: If unclear whether to fix test or code, ask user
- **Missing Spec**: If test intent is unclear, ask for clarification
- **Breaking Change**: If fix requires changing public API, get approval
