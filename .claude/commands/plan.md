---
name: plan
description: Design a feature before implementation. Outputs a structured plan for approval.
---

# Feature Planning

Design a feature thoroughly before writing any code. This ensures proper architecture and alignment with guidelines.

## Input

```
/plan <feature description>
```

## Output

Write the plan to: `plans/<feature-name>.md`

- Derive `<feature-name>` from the feature description (lowercase, hyphens, concise)
- Examples:
  - `/plan WebSocket reconnection logic` -> `plans/websocket-reconnection.md`
  - `/plan order validation middleware` -> `plans/order-validation-middleware.md`
- Create the `plans/` directory if it doesn't exist

## Instructions

You are designing a feature for a Go production system. Follow this workflow:

### Phase 1: UNDERSTAND

1. **Parse the request** - What is the user asking for?
2. **Ask clarifying questions** if anything is ambiguous:
   - What's the scope? (single function vs new package vs new service)
   - What modules are affected?
   - Are there performance requirements?
   - Are there specific edge cases to handle?
3. **Identify scope**:
   - Small: Helper function, simple change
   - Medium: New package, new type, new API
   - Large: New service module, major feature, cross-cutting concern

### Phase 2: REFERENCE

Read these guidelines before designing:

1. **Design Philosophy** - Read `.claude/skills/design-philosophy/SKILL.md`
   - Minimize cognitive load
   - Accept interfaces, return structs
   - Make zero values useful
   - Abstractions must earn their keep

2. **Data Safety** - Read `.claude/skills/data-safety/SKILL.md`
   - No float64 for money -> Use decimal
   - No ignored errors -> Handle every error
   - Log all state transitions
   - Handle reconnection and shutdown

3. **Existing Patterns** - Analyze relevant code in `internal/`:
   - How are similar features implemented?
   - What patterns does this codebase use?
   - What error types exist?

4. **Reference Guidelines** (for complex features):
   - `.reference/uber-go-guide/` - Uber's patterns
   - Google Go Style Guide (online)

### Phase 3: DESIGN

Create a structured design covering:

#### 3.1 Types

Define new types with proper validation:

```go
// Use dedicated types for domain concepts
type OrderID string

func NewOrderID(id string) (OrderID, error) {
    if id == "" {
        return "", errors.New("order ID cannot be empty")
    }
    return OrderID(id), nil
}
```

#### 3.2 Error Types

Define errors with sentinel values and structured types:

```go
var ErrNotFound = errors.New("not found")

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s: %s", e.Field, e.Message)
}
```

#### 3.3 Package Structure

Plan where code will live:

```
go-dsr/
├── main.go                 # Entry point
└── internal/
    ├── config/             # Configuration
    ├── peer/               # Peer connection management
    ├── routing/            # Routing table and logic
    ├── transaction/        # Transaction state
    └── ...                 # Additional packages as needed
```

#### 3.4 API Design

Define the public interface:

```go
// Constructor
func New(cfg Config, deps Dependencies) (*Service, error)

// Core methods
func (s *Service) Process(ctx context.Context, input Input) (*Output, error)

// Lifecycle
func (s *Service) Run(ctx context.Context) error
func (s *Service) Close() error
```

#### 3.5 Tests Needed

List tests to write:
- `TestFeature_HappyPath` - Normal operation
- `TestFeature_ErrorCase` - Error handling
- `TestFeature_EdgeCase` - Edge cases
- `TestFeature_Concurrent` - Race condition safety

#### 3.6 Integration Points

How does this feature connect to existing code?
- What existing packages does it use?
- What existing packages need to call it?
- Are there breaking changes?

### Phase 4: PRESENT

Output the plan in this format:

```markdown
## Plan: <Feature Name>

### Summary
Brief description of what this feature does.

### Scope
- [ ] Small / [x] Medium / [ ] Large
- Affected modules: `module1`, `module2`

### Requirements
1. Requirement 1
2. Requirement 2

### Design

#### Types
\```go
// Type definitions
\```

#### Errors
\```go
// Error definitions
\```

#### Package Structure
- `path/to/package/` - Description

#### Public API
\```go
// API signatures
\```

### Tests
- `TestName` - What it tests

### Questions/Decisions
- Any open questions for the user?

### Next Steps
Run `/create` to implement this design (references `plans/<feature-name>.md`).
```

### Phase 5: WRITE AND GET APPROVAL

1. **Write the plan** to `plans/<feature-name>.md`
2. **Ask the user**:
   - "I've written the plan to `plans/<feature-name>.md`. Does this design look good?"
   - "Any changes needed before implementation?"
   - "Ready to proceed with `/create`?"

## Notes

- This command produces a PLAN only, no code is written
- The plan is written to `plans/<feature-name>.md`
- The plan can be referenced by `/create` which will read from the plans directory
- For small tasks, user may skip `/plan` and go directly to `/create`
- Complex designs may need multiple iterations
