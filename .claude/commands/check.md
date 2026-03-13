---
name: check
description: Quick code quality validation. Runs format, lint, and build checks. Use during development.
---

# Quick Quality Check

Fast validation of code quality during development. Does NOT run tests (use `/test` for that).

## Input

```
/check              # Check entire workspace
/check <module>     # Check specific module
```

## Instructions

Run quick quality checks on the codebase. This is meant for fast feedback during development.

### Step 1: FORMAT CHECK

```bash
# Check formatting (list files that need formatting)
goimports -l .
```

**If formatting issues found:**
- Report files that need formatting
- Run `goimports -w .` to fix
- Re-check

### Step 2: VET CHECK

```bash
go vet ./...
```

### Step 3: LINT CHECK (golangci-lint)

```bash
# Full workspace
golangci-lint run ./...

# Specific package
golangci-lint run ./internal/routing/...
```

**If lint issues found:**
- List all issues by category
- For auto-fixable issues: `golangci-lint run --fix ./...`
- For manual fixes: Describe what needs to change

**Lint categories (by severity):**

| Category | Action | Example |
|----------|--------|---------|
| **errcheck** | Must fix | Unchecked error returns |
| **staticcheck** | Must fix | Static analysis issues |
| **govet** | Must fix | Vet issues |
| **gosec** | Must fix | Security issues |
| **revive** | Should fix | Style violations |
| **misspell** | Should fix | Spelling errors |
| **gocyclo** | Review | High complexity |

### Step 4: BUILD CHECK

```bash
# Quick build check
go build ./...
```

**If build fails:**
- Report compilation errors
- Identify root cause
- Suggest fix

### Step 5: REPORT

Report results in this format:

```markdown
## Check Results

### Format
Passed / N files need formatting

### Vet
Passed / N issues found

### Lint (golangci-lint)
Passed (0 issues) / N issues found

<details>
<summary>Issues (if any)</summary>

- `file.go:42` - errcheck: error return value not checked
- `file.go:87` - revive: exported function should have comment

</details>

### Build
Passed / Failed

### Summary
All checks passed - Ready for `/test` or commit
Issues found - Fix before proceeding
```

## What This Command Does NOT Do

- Does NOT run tests (use `/test`)
- Does NOT do full audit (use `/audit`)
- Does NOT fix issues automatically (reports only, unless asked)

## When to Use

- During development, after making changes
- Before running tests
- Before committing (quick sanity check)
- When you want fast feedback

## Quick Fix Mode

If you want to fix issues immediately, say:

```
/check --fix
```

This will:
1. Run `goimports -w .`
2. Run `golangci-lint run --fix ./...`
3. Re-run checks to verify

## Comparison with Other Commands

| Command | Speed | Scope | Fixes |
|---------|-------|-------|-------|
| `/check` | Fast | Format, lint, build | Reports only |
| `/test` | Medium | Tests only | Analyzes and fixes |
| `/audit` | Slow | Everything | Full review and fixes |
