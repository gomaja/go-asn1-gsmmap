---
name: commit
description: Create conventional commits with Go code quality validation
category: git
tags: [commit, conventional, go, golang]
version: 1.0.0
allowed-tools:
  - Task
  - Read
  - Grep
  - Bash
  - Glob
  # Git operations
  - Bash(git:*)
  # Go operations
  - Bash(go:*)
  - Bash(goimports:*)
  - Bash(golangci-lint:*)
  # File operations
  - Bash(ls:*)
  - Bash(cat:*)
  - Bash(head:*)
  - Bash(tail:*)
  # Text processing
  - Bash(rg:*)
  - Bash(grep:*)
  - Bash(sed:*)
  - Bash(awk:*)
  - Bash(sort:*)
  - Bash(uniq:*)
  # Date operations
  - Bash(date:*)
---

# Conventional Commit for Go Workspace

Create atomic commits following conventional commit specification with Go code quality
validation. Refer to CLAUDE.md and AGENTS.md for Go-specific guidelines.

## Process

1. **Analyze Changes**: Review staged/unstaged changes
2. **Run Quality Checks**: Format, lint, test (from CLAUDE.md)
3. **Classify Changes**: Determine type and scope from affected modules
4. **Generate Message**: Create conventional commit message
5. **Execute Commit**: Run git commit

## Quality Checks (from CLAUDE.md)

Before committing, run these checks:

```bash
# 1. Format code
goimports -w .

# 2. Run vet
go vet ./...

# 3. Run linter with strict settings
golangci-lint run ./...

# 4. Run tests with race detector
go test -race ./...

# 5. Build
go build -o go-dsr .
```

If any check fails, fix the issues before committing.

## Commit Message Format

```text
<type>(<scope>): <description>

[optional body with details]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `refactor` | Code refactoring (no functional change) |
| `test` | Adding or updating tests |
| `docs` | Documentation only |
| `chore` | Maintenance (deps, configs, scripts) |
| `perf` | Performance improvements |

### Scope Detection

Scope is the affected package name:

| Changed Files | Scope |
|---------------|-------|
| `internal/peer/**` | `peer` |
| `internal/routing/**` | `routing` |
| `internal/transaction/**` | `transaction` |
| `internal/config/**` | `config` |
| `main.go` | `dra` |
| `configs/**` | `config` |
| Multiple packages | use primary affected package |

### Examples

```text
feat(routing): add realm-based routing table

fix(peer): handle context cancellation in reconnection loop

refactor(transaction): extract hop-by-hop ID mapping

chore(config): update default configuration

test(routing): add table-driven tests for route selection
```

## Change Type Classification

**File Pattern Analysis:**

- `**/*.go` (new exported functions/types) -> `feat`
- `**/*.go` (bug fixes) -> `fix`
- `**/*.go` (restructuring) -> `refactor`
- `**/*_test.go` -> `test`
- `*.md`, `docs/**` -> `docs`
- `go.mod`, `go.sum` -> `chore`
- `configs/**` -> `chore`

## Important Rules

From CLAUDE.md:
- **No AI attribution** in commit messages
- **Never commit unless explicitly asked**
- Use conventional commits: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`

## Arguments

- `<message>`: Custom commit message (bypasses auto-generation)
- `--all`: Stage all changes before committing
- `--skip-checks`: Skip quality checks (use sparingly)
- `--scope <scope>`: Override automatic scope detection
- `--type <type>`: Override automatic type detection
