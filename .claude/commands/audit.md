---
name: audit
description: Full production readiness review. Comprehensive check against all guidelines with detailed findings report.
---

# Production Readiness Audit

Comprehensive code quality audit against all guidelines. The most thorough review available.

## Input

```
/audit                  # Full audit, report only
/audit --fix            # Full audit with automatic fixes
/audit <module>         # Audit specific module
```

## Instructions

You are performing a complete production readiness audit of a Go system.

### Phase 1: READ ALL GUIDELINES

Before auditing, load the complete context:

1. **Data Safety** - `.claude/skills/data-safety/SKILL.md`
   - Financial calculation patterns
   - Error handling requirements
   - Logging requirements
   - Concurrency patterns

2. **Design Philosophy** - `.claude/skills/design-philosophy/SKILL.md`
   - Cognitive load principles
   - Interface design
   - Package design

3. **Production Audit Skill** - `.claude/skills/production-audit/SKILL.md`
   - Complete checklist
   - Fix patterns

4. **Reference Guidelines**:
   - `.reference/uber-go-guide/` - Uber patterns
   - Google Go Style Guide (online reference)

### Phase 2: RUN ALL QUALITY CHECKS

Execute the complete quality suite:

```bash
# 1. Format check
goimports -l .

# 2. Vet check
go vet ./...

# 3. Lint check
golangci-lint run ./...

# 4. Test with race detector
go test -race ./...

# 5. Build check
go build -o go-dsr .

# 6. Security audit
gosec ./...
```

Record all outputs for analysis.

### Phase 3: CRITICAL AUDIT

Search for critical issues:

#### 3.1 Floating Point for Money (CRITICAL)

```bash
grep -rn "float64\|float32" --include="*.go" . | grep -v _test.go | grep -v vendor
```

**Each occurrence must be evaluated:**
- Is it used for prices/quantities? -> CRITICAL
- Is it used for non-financial calculations (e.g., percentages, metrics)? -> OK

#### 3.2 Ignored Errors (CRITICAL)

```bash
grep -rn ", _\|_ =" --include="*.go" . | grep -v _test.go | grep -v vendor
```

**Each occurrence must be evaluated:**
- Is an error being discarded? -> CRITICAL
- Is it a non-error blank identifier (e.g., range index)? -> OK

#### 3.3 Panics in Production (CRITICAL)

```bash
grep -rn "panic(\|log.Fatal" --include="*.go" . | grep -v _test.go | grep -v vendor
```

**Each occurrence must be evaluated:**
- Is it in a production code path? -> CRITICAL
- Is it in `init()` for truly unrecoverable setup? -> Review
- Is it in test code only? -> OK

#### 3.4 Missing Context Propagation

For each function doing I/O:
- [ ] Has `context.Context` as first parameter
- [ ] Passes context to downstream calls
- [ ] Sets timeouts for external calls

#### 3.5 Goroutine Lifecycle

For each goroutine:
- [ ] Has context cancellation
- [ ] Uses errgroup for lifecycle management
- [ ] Handles panics (recover in goroutine if needed)
- [ ] Has proper cleanup

#### 3.6 State Transition Logging

For each significant state change:
- [ ] Logged with `slog`
- [ ] Includes old and new values
- [ ] Includes reason/context

### Phase 4: CODE QUALITY AUDIT

#### 4.1 Documentation

- [ ] All packages have doc comments
- [ ] All exported types and functions have doc comments
- [ ] Complex logic has explanatory comments

#### 4.2 Test Coverage

- [ ] Happy path tests exist (table-driven)
- [ ] Error case tests exist
- [ ] Edge cases are tested
- [ ] Race detector passes (`go test -race`)

#### 4.3 API Design (per Uber/Google Guides)

- [ ] Functions accept interfaces, return structs
- [ ] Interfaces are small (1-3 methods) and defined at usage site
- [ ] Error types implement `Unwrap()` where appropriate
- [ ] `context.Context` is first parameter
- [ ] Functional options for complex constructors

### Phase 5: GENERATE REPORT

Create comprehensive audit report:

```markdown
## Audit Report: [Project Name]

**Date**: YYYY-MM-DD
**Scope**: [workspace / specific module]

---

### Executive Summary

| Category | Status | Issues |
|----------|--------|--------|
| Format | Pass/Warn/Fail | N |
| Vet | Pass/Warn/Fail | N |
| Lint | Pass/Warn/Fail | N |
| Tests | Pass/Warn/Fail | N |
| Security | Pass/Warn/Fail | N |
| Production Safety | Pass/Warn/Fail | N |
| Documentation | Pass/Warn/Fail | N |

**Overall**: Production Ready / Issues Found / Critical Issues

---

### Critical Issues (Must Fix)

#### 1. [Issue Title]
- **Location**: `file.go:42`
- **Severity**: CRITICAL
- **Issue**: Description
- **Fix**: How to fix

---

### Warnings (Should Fix)

#### 1. [Issue Title]
- **Location**: `file.go:87`
- **Severity**: WARNING
- **Issue**: Description
- **Fix**: How to fix

---

### Recommendations (Nice to Have)

- Recommendation 1
- Recommendation 2

---

### Files Reviewed

| Module | Files | Status |
|--------|-------|--------|
| shared | 12 | Pass |
| service_a | 8 | Warn |
| ... | ... | ... |

---

### Next Steps

1. Run `/fix from audit` to address critical issues
2. Run `/fix from audit` again for warnings
3. Run `/audit` to verify all issues resolved
```

### Phase 6: APPLY FIXES (if --fix)

If `--fix` flag is provided:

1. Fix all CRITICAL issues first
2. Re-run quality checks
3. Fix WARNINGS
4. Re-run quality checks
5. Update report with "Fixed" status

### Phase 7: FINAL VERIFICATION

After fixes (or at end of report):

```bash
goimports -l .
go vet ./...
golangci-lint run ./...
go test -race ./...
go build -o go-dsr .
```

All must pass for "Production Ready" status.

## Severity Levels

| Level | Meaning | Action |
|-------|---------|--------|
| Critical | Can cause incorrect behavior or system failure | Must fix before deploy |
| Warning | Code smell or potential issue | Should fix soon |
| Info | Improvement opportunity | Nice to have |
| Pass | Meets standards | No action needed |

## Comparison with Other Commands

| Command | Depth | Speed | Use Case |
|---------|-------|-------|----------|
| `/check` | Surface | Fast | During development |
| `/test` | Tests only | Medium | After changes |
| `/audit` | Complete | Slow | Before release |
