---
name: run-lint
description: Run golangci-lint with strict settings and format code. Use after code changes and before commits.
---

# Linting and Formatting

## When to Use

- After making code changes
- Fixing lint warnings
- Before creating commits
- When user says "run lint", "fix lints", "format code"

## Instructions

### Step 1: Format Code

```bash
goimports -w .
```

### Step 2: Run Vet

```bash
go vet ./...
```

### Step 3: Run golangci-lint

```bash
golangci-lint run ./...
```

**All issues must be resolved.** Fix all problems.

For auto-fixable issues:

```bash
golangci-lint run --fix ./...
```

### Common Lint Categories

| Category | Priority | Action |
|----------|----------|--------|
| errcheck | Must fix | Unchecked error returns |
| staticcheck | Must fix | Static analysis issues |
| govet | Must fix | Go vet issues |
| gosec | Must fix | Security issues |
| ineffassign | Should fix | Ineffective assignments |
| misspell | Should fix | Spelling errors |
| revive | Should fix | Style issues |
| gocyclo | Review | High cyclomatic complexity |

### Step 4: Production-Specific Checks

Check for these anti-patterns:

```go
// BAD: Floating point for money
price := 123.45

// BAD: Ignored error
result, _ := doSomething()

// BAD: Panic in production
panic("unexpected state")

// BAD: Missing context
func Process(data []byte) error { ... }
// GOOD: Context propagation
func Process(ctx context.Context, data []byte) error { ... }
```

### Step 5: Run Tests

If lint fixes modified behavior:

```bash
go test -race ./...
```

## Reporting Results

- All lints passed -> "Lint checks passed!"
- Issues found -> List issue categories and counts
- Auto-fixed -> Report what was corrected

## Recommended golangci-lint Configuration

Create `.golangci.yml` in the workspace root:

```yaml
run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gosec
    - revive
    - misspell
    - gocyclo
    - unconvert
    - unparam
    - goimports
    - errname
    - errorlint
    - exhaustive
    - noctx
    - prealloc

linters-settings:
  gocyclo:
    min-complexity: 15
  revive:
    rules:
      - name: exported
        severity: warning
  errcheck:
    check-type-assertions: true
    check-blank: true
  govet:
    enable-all: true
  noctx:
    # Flags HTTP requests without context
```

## Related Commands

- `/check` - Invokes full quality check
