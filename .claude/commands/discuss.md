---
name: discuss
description: Explore concepts, approaches, and trade-offs before committing to a design. Research-oriented, no code output.
---

# Conceptual Discussion

Explore ideas, research industry practices, and analyze trade-offs before deciding what to build. This command is for thinking through problems conceptually without writing any code.

## Input

```
/discuss <topic or question>
```

## Output

**Conversational response only** - no files are written, no code is shown.

## When to Use

- You're unsure what approach to take
- You want to understand industry best practices first
- You need to explore naming conventions
- You want to compare multiple architectural options
- You're deciding whether a feature is even needed

## Instructions

You are helping explore a concept for a Go production system. Follow this workflow:

### Phase 1: UNDERSTAND THE QUESTION

1. **Parse the topic** - What concept or decision is being explored?
2. **Identify the type of discussion**:
   - **Architecture**: How should something be structured?
   - **Naming**: What should things be called?
   - **Approach**: Which technique or pattern to use?
   - **Trade-offs**: What are the pros/cons of options?
   - **Feasibility**: Should we even do this?
3. **Ask clarifying questions** if the topic is too broad or ambiguous

### Phase 2: ANALYZE CURRENT CODEBASE

Before researching externally, understand internal context:

1. **Existing patterns** - How does this repo currently handle similar concerns?
2. **Naming conventions** - What naming style is already in use?
3. **Dependencies** - What libraries are already available?
4. **Constraints** - What are the project's established principles? (Reference CLAUDE.md)

Summarize findings in plain language, not code.

### Phase 3: RESEARCH INDUSTRY STANDARDS

Use web search and references to understand:

1. **How mature Go systems solve this** - What do production systems do?
2. **Go ecosystem conventions** - What's idiomatic in the Go community?
3. **Authoritative sources** - What do Uber/Google style guides recommend?
4. **Common pitfalls** - What mistakes do others make?

Focus on:
- Naming conventions used by well-known projects (Kubernetes, Prometheus, etc.)
- Architectural patterns in similar domains
- Best practices from authoritative sources (Effective Go, Uber guide, Google guide)

### Phase 4: PRESENT OPTIONS

Present 2-3 distinct approaches:

For each option:
- **Name** - A short label for the approach
- **Description** - What it involves (in plain language, no code)
- **Pros** - Benefits and strengths
- **Cons** - Drawbacks and risks
- **When to use** - Scenarios where this option shines

**Important**: Describe approaches conceptually. Do NOT include code snippets, type definitions, or implementation details.

### Phase 5: HIGHLIGHT DECISION POINTS

Identify key questions that need answers before proceeding:

- What assumptions need validation?
- What trade-offs need explicit decision?
- What information is still missing?
- What constraints might change the answer?

### Phase 6: RECOMMEND (IF APPROPRIATE)

If one option is clearly better given the context:
- State the recommendation clearly
- Explain why it fits this project

If it depends:
- Say "It depends on..." and list the deciding factors
- Don't force a recommendation when trade-offs are genuinely balanced

## Output Format

Structure your response as:

```
## Discussion: <Topic>

### Context
What the current codebase does related to this topic. What constraints exist.

### Industry Practices
What authoritative sources and mature projects recommend.

### Options

**Option 1: <Name>**
- Description: ...
- Pros: ...
- Cons: ...
- Best when: ...

**Option 2: <Name>**
- Description: ...
- Pros: ...
- Cons: ...
- Best when: ...

[Option 3 if applicable]

### Naming Conventions
What's standard in the Go ecosystem and what's consistent with this repo.

### Decision Points
- Question 1?
- Question 2?

### Recommendation
[Clear recommendation] or [Depends on X, Y, Z]

### Next Steps
- Continue discussing specific aspects
- Or: Ready to `/plan <specific feature>` based on chosen approach
```

## Rules

1. **NO CODE** - Do not show code snippets, type definitions, or implementation details
2. **NO FILES** - Do not write any files
3. **RESEARCH FIRST** - Use web search for industry standards when relevant
4. **STAY CONCEPTUAL** - Discuss ideas, patterns, and approaches in plain language
5. **BE OBJECTIVE** - Present trade-offs honestly, don't oversell any option
6. **RESPECT CONTEXT** - Consider existing project patterns and constraints

## Examples

Good `/discuss` topics:
- `/discuss error handling strategy for multi-service communication`
- `/discuss naming convention for domain model types`
- `/discuss whether to use channels or shared state for updates`
- `/discuss graceful shutdown approach for long-running workers`
- `/discuss splitting the service into smaller packages`

These would then lead to:
- `/plan <specific feature>` once an approach is chosen
