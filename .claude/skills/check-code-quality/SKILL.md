---
name: check-code-quality
description: Run comprehensive Go code quality checks. Use after completing code changes and before creating commits.
---

# Go Code Quality Checks

## When to Use

- After completing significant code changes
- Before creating commits
- Before creating pull requests
- When user says "check code quality", "run quality checks", etc.

## Instructions

Run this quality checklist in order from the project root:

### 1. Format Code

```bash
goimports -w .
```

### 2. Vet Check

```bash
go vet ./...
```

### 3. Run golangci-lint (Strict)

```bash
golangci-lint run ./...
```

All issues must be resolved. Common auto-fixable issues:

```bash
golangci-lint run --fix ./...
```

### 4. Run Tests with Race Detector

```bash
go test -race ./...
```

If tests fail, analyze and fix before continuing.

### 5. Build All

```bash
go build ./...
```

Catches compilation issues across all modules.

### 6. Production-Specific Checks

Review changed files for:

- [ ] No `float64`/`float32` for prices or quantities
- [ ] No ignored errors (`_` on error returns) in production paths
- [ ] No `panic()` or `log.Fatal()` in production paths
- [ ] All external calls use `context.Context` with timeouts
- [ ] Goroutines have proper lifecycle management (errgroup)
- [ ] State changes are logged with `slog`

## Reporting Results

After running all checks:

- All checks passed -> "All quality checks passed! Ready to commit."
- Some checks failed -> List which steps failed and what needs fixing
- Auto-fixed -> Report what was automatically fixed

## Related Skills

- `run-lint` - Focused linting
- `data-safety` - Production data safety validation

## Related Commands

- `/check` - Invokes this skill
