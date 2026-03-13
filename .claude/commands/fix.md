---
name: fix
description: Fix specific issues - bugs, test failures, lint warnings, or audit findings. Targeted remediation.
---

# Fix Issues

Targeted remediation for specific issues. Use when something is broken and needs fixing.

## Input

```
/fix <issue description>       # Fix a specific issue
/fix tests                     # Fix failing tests (alias for /test)
/fix lint                      # Fix lint warnings
/fix from audit                # Fix issues found by /audit
/fix from check                # Fix issues found by /check
```

## Instructions

You are fixing specific issues in a Go production system. This is targeted remediation, not general improvement.

### Phase 1: UNDERSTAND THE ISSUE

#### 1.1 Parse the Request

Determine issue type:

| Input | Type | Action |
|-------|------|--------|
| Description of bug | Bug fix | Locate and fix code |
| "tests" | Test failures | Run `/test` workflow |
| "lint" | Lint warnings | Fix lint issues |
| "from audit" | Audit findings | Fix issues from last audit |
| "from check" | Check failures | Fix issues from last check |

#### 1.2 Locate the Issue

For bug descriptions:
- Search codebase for relevant code
- Identify the root cause
- Understand the expected behavior

For test/lint/audit:
- Reference the output from the previous command
- List specific failures to address

### Phase 2: REFERENCE GUIDELINES

Before fixing, check the correct pattern:

1. **Data Safety** - `.claude/skills/data-safety/SKILL.md`
   - Ensure fix uses decimal, not float64
   - Ensure fix handles all errors
   - Add logging if state changes

2. **Design Philosophy** - `.claude/skills/design-philosophy/SKILL.md`
   - Fix should not increase complexity
   - Fix should be minimal and focused

3. **Existing Patterns**
   - How is similar code handled elsewhere?
   - Follow existing conventions

### Phase 3: FIX

Apply targeted fix based on issue type:

#### 3.1 Bug Fixes

```go
// 1. Identify the bug
// Division by zero at line 42

// 2. Apply minimal fix
// Before
result := a / b

// After
if b == 0 {
    return fmt.Errorf("division by zero: denominator is %v", b)
}
result := a / b
```

**Bug fix principles:**
- Fix only the bug, don't refactor surrounding code
- Add test to prevent regression
- Log the fix if it affects state

#### 3.2 Lint Fixes

Common lint fixes:

```go
// errcheck -> handle errors
// Before
json.Unmarshal(data, &obj)
// After
if err := json.Unmarshal(data, &obj); err != nil {
    return fmt.Errorf("unmarshaling: %w", err)
}

// noctx -> add context to HTTP requests
// Before
resp, err := http.Get(url)
// After
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
if err != nil {
    return fmt.Errorf("creating request: %w", err)
}
resp, err := http.DefaultClient.Do(req)

// revive/exported -> add doc comment
// Before
func ProcessOrder(o *Order) error { ... }
// After
// ProcessOrder validates and submits the given order.
func ProcessOrder(o *Order) error { ... }
```

#### 3.3 Test Failure Fixes

See `/test` command for detailed workflow. Key points:
- Determine if test or code is wrong
- Fix the correct one
- Re-run with race detector

#### 3.4 Audit Finding Fixes

Common audit fixes:

| Finding | Fix |
|---------|-----|
| float64 for money | Replace with `decimal.Decimal` |
| Ignored error | Add error checking with `if err != nil` |
| Panic in production | Replace with error return |
| Missing context | Add `context.Context` as first parameter |
| Missing logging | Add `slog.Info` for state changes |
| Goroutine leak | Add errgroup lifecycle management |

### Phase 4: VALIDATE

After applying fix:

```bash
# 1. Ensure fix compiles
go build -o go-dsr .

# 2. Ensure fix doesn't break vet
go vet ./...

# 3. Ensure fix doesn't break lint
golangci-lint run ./...

# 4. Ensure fix doesn't break tests
go test -race ./...
```

**If new issues arise from fix:**
1. Fix the new issues
2. Re-validate
3. Repeat until clean

### Phase 5: REPORT

```markdown
## Fixed: <Issue Summary>

### Issue
<Description of what was wrong>

### Root Cause
<Why it was wrong>

### Fix Applied
- `file.go:42` - <what changed>

### Code Change
```go
// Before
<old code>

// After
<new code>
```

### Validation
- go build - passed
- go vet - passed
- golangci-lint - passed
- go test -race - passed

### Regression Prevention
- Added test: `TestName` to prevent recurrence
```

## Scope Limits

**Do:**
- Fix the specific issue
- Add test to prevent regression
- Follow existing patterns

**Don't:**
- Refactor surrounding code (unless necessary for fix)
- Add unrelated improvements
- Change public API (without approval)

## When to Ask User

- **Unclear cause**: If multiple potential causes, ask which to investigate
- **API change needed**: If fix requires changing public API, get approval
- **Trade-off decision**: If fix has performance/complexity trade-offs, present options
- **Test vs code**: If unclear whether test or code is wrong, ask
