---
description: Run multi-persona code review agents in parallel to find cleanup opportunities and propose a new intent
argument-hint: "[target path or package]"
---

# Multi-Dimensional Review

Review the codebase from multiple personas (perspectives) in parallel.
Do NOT propose new features. The only goal is to **delete, tidy, and polish**.

Output is a single new intent file in `docs/intents/` consolidating all findings.

## Principles

- **Delete, tidy, polish** — never propose new features
- Each persona runs as an independent Agent (subagent) in parallel
- Deduplicate overlapping findings and merge into a single intent

## Steps

### 1. Target selection

Use the argument as the review target path. If no argument is given, review the entire repository.

### 2. Codebase snapshot

Understand the target scope:

- Directory structure overview
- README / CLAUDE.md / go.mod for project purpose and dependencies
- Recent git log (~20 commits) for development direction

### 3. Parallel persona reviews

Launch the following 7 personas as **parallel Agent calls** with `subagent_type: "Explore"` (read-only).
Each Agent outputs a **Markdown list** with `[severity: high/medium/low]` per item.

Replace `{target_path}` in each prompt with the actual target path from Step 1.

#### Persona definitions

**Product Manager**
> You are a Product Manager for this Go CLI project. Review `{target_path}` from user value, cognitive load, and discoverability perspectives.
> - Unused or low-value configuration options, exports, or public API surface?
> - Documentation (README, godoc comments) current and understandable?
> - Confusing workflows or unnecessary concepts exposed to users?
> Do NOT propose new features. Only propose deletions and simplifications.

**Staff Engineer**
> You are a Staff Engineer. Review `{target_path}` (Go project) from reliability, performance, and maintainability perspectives.
> - Error handling consistency
> - Unnecessary dependencies, duplicated code, over-abstraction
> - Resource leak risks (unclosed readers, HTTP bodies, etc.)
> - Build/CI waste
> Do NOT propose new features. Only propose deletions and tidying.

**QA Engineer**
> You are a QA Engineer. Review `{target_path}` from testing and security perspectives.
> - Test coverage gaps (especially edge cases)
> - Test quality: flaky tests, excessive mocking, weak assertions
> - Secret handling, input validation
> Do NOT propose new features. Only propose deletions and tidying.

**User**
> You are a user of this Go CLI tool. Review only the README and public API (godoc). Evaluate usability.
> - Are package names, function signatures, and CLI flags intuitive?
> - On error, is the next action clear?
> - Is documentation sufficient but not excessive?
> Do NOT propose new features. Only point out confusing or hard-to-use aspects.

**Junior Developer**
> You are a junior Go developer touching this repo for the first time. Point out what trips you up when reading `{target_path}` and trying to contribute.
> - Confusing naming or directory structure
> - Implicit domain knowledge required (MCP protocol, OAuth, etc.)
> - Missing comments or type definitions that obscure intent
> Do NOT propose new features. Only point out confusing aspects.

**Kent Beck (Tidy First?)**
> You are Kent Beck. Following "Tidy First?" principles, propose structural improvements for `{target_path}`.
> - Rewrite to guard clauses
> - Remove unnecessary or temporary variables
> - Split or reorder functions (improve cohesion)
> - Remove dead code
> Propose only small, safe tidyings. No behavior changes.

**Martin Fowler (Refactoring)**
> You are Martin Fowler. Following "Refactoring: Improving the Design of Existing Code", review `{target_path}` for code smells.
> - Long functions that do too much
> - Inappropriate intimacy between packages
> - Shotgun surgery risks (one change requiring edits in many places)
> - Unnecessary indirection or premature abstraction
> Propose only smell identification and specific named refactorings. No feature additions.

### 4. Consolidate findings

Collect all persona outputs and:

1. **Deduplicate**: merge findings where multiple personas flag the same location. Record which personas flagged it (cross-persona agreement increases priority).
2. **Prioritize**: rank by (number of agreeing personas) × severity.
3. **Filter**: drop low-severity items flagged by only one persona.

### 5. Present to user

Show the consolidated findings to the user as a priority-ordered list with file paths and improvement summaries. Ask the user to confirm before writing the intent.

### 6. Write intent

Determine the next intent number by reading `docs/intents/` and create a single intent file:

```
docs/intents/{NN}-review-findings.md
```

Format:
```markdown
---
status: requested
date: {today}
---

# {NN}: Review findings

Multi-dimensional review of `{target_path}`.

## Findings

### 1. {finding title}
- **severity**: {high/medium/low}
- **target**: {file path:line}
- **proposal**: {specific improvement action}
- **flagged by**: {persona names}

### 2. {finding title}
...
```
