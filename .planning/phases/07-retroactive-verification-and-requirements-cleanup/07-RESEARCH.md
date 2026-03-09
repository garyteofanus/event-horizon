# Phase 7: Retroactive Verification and Requirements Cleanup - Research

**Researched:** 2026-03-09
**Domain:** Documentation gap closure, dead code removal, requirements traceability
**Confidence:** HIGH

## Summary

Phase 7 is a documentation and cleanup phase with no new features. The v1.1 milestone audit identified five specific gaps: missing VERIFICATION.md files for phases 3, 4, and 5; a missing SUMMARY.md for phase 5; REQUIREMENTS.md not updated with Phase 6 requirement IDs; stale checkboxes in REQUIREMENTS.md; and dead `formatRequestLine` code in handler.go. All features are already implemented and tested -- this phase only closes process/documentation gaps.

The work is entirely mechanical: examine existing code and summaries, produce verification reports following the established template (see phases 01 and 06), write a phase 5 summary from the existing plan, update REQUIREMENTS.md, and remove dead code with its tests.

**Primary recommendation:** Execute this as a single plan with 3-4 small tasks since all work is independent documentation/cleanup with no code risk beyond the dead code removal.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| TUI-01 | TUI on main goroutine | Needs retroactive VERIFICATION.md for Phase 3 confirming this is implemented |
| TUI-02 | Channel bridge | Needs retroactive VERIFICATION.md for Phase 3 confirming this is implemented |
| TUI-03 | File-only logging | Needs retroactive VERIFICATION.md for Phase 3 confirming this is implemented |
| DISP-01 | Compact one-line rows | Needs retroactive VERIFICATION.md for Phase 4 confirming this is implemented |
| DISP-02 | Method color coding | Needs retroactive VERIFICATION.md for Phase 4 confirming this is implemented |
| DISP-03 | Status code colors | Needs retroactive VERIFICATION.md for Phase 4 confirming this is implemented |
| DISP-04 | Visual separation | Needs retroactive VERIFICATION.md for Phase 4 confirming this is implemented |
| INTR-01 | j/k navigation | Needs retroactive VERIFICATION.md for Phase 5 + SUMMARY.md confirming this is implemented |
| INTR-02 | Expand/collapse | Needs retroactive VERIFICATION.md for Phase 5 + SUMMARY.md confirming this is implemented |
| INTR-03 | Clear entries | Needs retroactive VERIFICATION.md for Phase 5 + SUMMARY.md confirming this is implemented |
| INTR-04 | Help footer | Needs retroactive VERIFICATION.md for Phase 5 + SUMMARY.md confirming this is implemented |
| ROBU-01 | Resize handling | Needs retroactive VERIFICATION.md for Phase 5 + SUMMARY.md confirming this is implemented |
| COPY-01 | Copy body with c | Already verified in 06-VERIFICATION.md; needs REQUIREMENTS.md entry |
| COPY-02 | Copy full with Shift+C | Already verified in 06-VERIFICATION.md; needs REQUIREMENTS.md entry |
| COPY-03 | Copied flash feedback | Already verified in 06-VERIFICATION.md; needs REQUIREMENTS.md entry |
| COPY-04 | Empty body message | Already verified in 06-VERIFICATION.md; needs REQUIREMENTS.md entry |
| COPY-05 | Flash auto-dismiss | Already verified in 06-VERIFICATION.md; needs REQUIREMENTS.md entry |
| FMT-01 | JSON auto-format | Already verified in 06-VERIFICATION.md; needs REQUIREMENTS.md entry |
| FMT-02 | Format toggle f key | Already verified in 06-VERIFICATION.md; needs REQUIREMENTS.md entry |
| FMT-03 | Body label JSON/raw | Already verified in 06-VERIFICATION.md; needs REQUIREMENTS.md entry |
| FMT-04 | Syntax highlighting | Already verified in 06-VERIFICATION.md; needs REQUIREMENTS.md entry |
| KEY-01 | Clear remapped to x | Already verified in 06-VERIFICATION.md; needs REQUIREMENTS.md entry |
| KEY-02 | Shift+C copies full | Already verified in 06-VERIFICATION.md; needs REQUIREMENTS.md entry |
</phase_requirements>

## Standard Stack

No new libraries or tools needed. This phase uses only:

| Tool | Purpose |
|------|---------|
| `go test ./...` | Verify tests still pass after dead code removal |
| `go build -o /dev/null .` | Verify build after dead code removal |
| Existing VERIFICATION.md template | Follow the pattern from phases 01 and 06 |

## Architecture Patterns

### VERIFICATION.md Template

Phases 01 and 06 have VERIFICATION.md files that follow a consistent pattern. Key sections:

```
---
phase: {phase-name}
verified: {timestamp}
status: passed
score: N/N must-haves verified
---

# Phase N: {Name} Verification Report

## Goal Achievement
### Observable Truths (table with # | Truth | Status | Evidence)
### Required Artifacts (table)
### Key Link Verification (table)
### Requirements Coverage (table with Requirement | Source Plan | Description | Status | Evidence)
### Anti-Patterns Found (table)
### Human Verification Required (if applicable)
### Gaps Summary
```

**For retroactive verification:** The evidence column should cite current code line numbers, existing test names, and SUMMARY.md entries. Mark all as "VERIFIED" since the audit confirmed all features are implemented.

### SUMMARY.md Template

Existing summaries follow this pattern (see 03-01-SUMMARY.md, 04-01-SUMMARY.md, 06-01-SUMMARY.md):

```
---
phase: {phase-name}
plan: {number}
subsystem: {area}
tags: [...]
requires/provides/affects
tech-stack/key-files/key-decisions/patterns-established
requirements-completed: [...]
duration/completed
---

# Phase N Plan M: {Title} Summary
## Performance
## Accomplishments
## Task Commits
## Files Created/Modified
## Decisions Made
## Deviations from Plan
```

**For Phase 5 SUMMARY.md:** Reconstruct from the plan (05-01-PLAN.md) and the current codebase state. The plan was executed but never documented. Check git log for the relevant commits.

### Dead Code Removal Pattern

The `formatRequestLine` function in handler.go (lines 82-100) and its three test functions in handler_test.go (TestFormatRequestLine, TestFormatRequestLineURITruncation, TestFormatRequestLineSubMillisecond, lines ~170-227) must be removed. This function was created in Phase 3 and superseded by `renderRequestRow` in Phase 4.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Verification format | New template | Copy structure from 01-VERIFICATION.md and 06-VERIFICATION.md | Consistency across all phases |
| Summary format | New template | Copy structure from existing SUMMARY.md files | Project convention already established |
| Requirements update | Manual editing | Mechanical: add Phase 6 section, update checkboxes, update traceability | REQUIREMENTS.md already has the structure |

## Common Pitfalls

### Pitfall 1: Inconsistent Verification Evidence
**What goes wrong:** Verification cites outdated line numbers or wrong test names
**How to avoid:** Read actual current source files (tui.go, handler.go, tui_test.go, handler_test.go) to get current line numbers and test function names before writing verification

### Pitfall 2: Breaking Tests When Removing Dead Code
**What goes wrong:** Removing formatRequestLine breaks the 3 tests that call it
**How to avoid:** Remove both the function AND its tests simultaneously. Run `go test ./...` after to confirm no other code references it. Current grep shows no callers besides the 3 test functions.

### Pitfall 3: Phase 5 Summary Missing Context
**What goes wrong:** Phase 5 summary is fabricated rather than evidence-based
**How to avoid:** Use git log to find actual commits, check the plan for what was specified, and verify against current code state

### Pitfall 4: REQUIREMENTS.md Checkbox Discrepancy
**What goes wrong:** Marking checkboxes [x] without verifying features actually work
**How to avoid:** The audit already confirmed all features are implemented. The checkboxes in REQUIREMENTS.md were already updated to [x] in a recent edit (current file shows all [x]). Double-check against audit findings.

## Current State Analysis

### What Already Exists (no work needed)
- REQUIREMENTS.md: Already contains COPY-01 through COPY-05, FMT-01 through FMT-04, KEY-01 and KEY-02 in both the requirements list and traceability table (lines 65-77 and 153-163). All checkboxes are already `[x]`.
- Phase 1 VERIFICATION.md: Complete
- Phase 2 VERIFICATION.md: Complete
- Phase 6 VERIFICATION.md: Complete
- Phase 3 SUMMARY.md (2 plans): Complete
- Phase 4 SUMMARY.md: Complete

### What Is Missing (work needed)
1. **Phase 3 VERIFICATION.md** -- must verify TUI-01, TUI-02, TUI-03
2. **Phase 4 VERIFICATION.md** -- must verify DISP-01, DISP-02, DISP-03, DISP-04
3. **Phase 5 VERIFICATION.md** -- must verify INTR-01, INTR-02, INTR-03, INTR-04, ROBU-01
4. **Phase 5 SUMMARY.md** -- must document implementation from 05-01-PLAN.md
5. **Dead code removal** -- formatRequestLine in handler.go + 3 tests in handler_test.go

### REQUIREMENTS.md Status
The REQUIREMENTS.md file has ALREADY been updated with Phase 6 entries and all checkboxes marked [x] (verified by reading the current file). The audit was performed before this update. The success criteria item 3 ("REQUIREMENTS.md contains all Phase 6 requirement IDs") and item 4 ("All implemented v1.1 requirement checkboxes are marked [x]") appear to already be satisfied. The planner should verify this and skip redundant work.

## Code Examples

### Dead Code to Remove from handler.go (lines 82-100)
```go
// formatRequestLine formats a RequestData into a human-readable line:
// "HH:MM:SS METHOD /path STATUS TIMEms"
func formatRequestLine(d RequestData) string {
	ts := d.Timestamp.Format("15:04:05")
	uri := d.URI
	if len(uri) > 40 {
		uri = uri[:37] + "..."
	}
	var timing string
	if d.ResponseTime < time.Millisecond {
		timing = "<1ms"
	} else {
		timing = fmt.Sprintf("%dms", d.ResponseTime.Milliseconds())
	}
	return fmt.Sprintf("%s %s %s %d %s", ts, d.Method, uri, d.Status, timing)
}
```

### Tests to Remove from handler_test.go (3 functions, lines ~170-227)
- `TestFormatRequestLine`
- `TestFormatRequestLineURITruncation`
- `TestFormatRequestLineSubMillisecond`

### Test Counts
- handler_test.go: 8 test functions (3 will be removed, leaving 5)
- tui_test.go: 46 test functions (unchanged)

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | none (Go convention) |
| Quick run command | `go test ./... -count=1` |
| Full suite command | `go test -v ./...` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| (all) | Dead code removal doesn't break build | smoke | `go build -o /dev/null . && go test ./...` | N/A |

### Sampling Rate
- **Per task commit:** `go test ./...`
- **Per wave merge:** `go test -v ./...`
- **Phase gate:** Full suite green before verify

### Wave 0 Gaps
None -- existing test infrastructure covers all phase requirements. This phase removes tests, it does not add them.

## Open Questions

1. **Phase 5 commit history**
   - What we know: Phase 5 features are implemented in tui.go and tui_test.go
   - What's unclear: Exact commit SHAs and dates for Phase 5 implementation (need `git log` during planning/execution)
   - Recommendation: Use git log to find commits that added navigation, expansion, clear, footer, and resize features

2. **REQUIREMENTS.md already updated?**
   - What we know: Current REQUIREMENTS.md file already has all Phase 6 entries and all [x] checkboxes
   - What's unclear: Whether the audit's gap findings are now stale
   - Recommendation: Planner should verify current state and skip if already done

## Sources

### Primary (HIGH confidence)
- Current codebase files (handler.go, tui.go, handler_test.go, tui_test.go) -- read directly
- Existing VERIFICATION.md files (phases 01, 06) -- template reference
- Existing SUMMARY.md files (phases 03, 04, 06) -- template reference
- v1.1-MILESTONE-AUDIT.md -- gap identification source
- REQUIREMENTS.md -- current state verified by direct read
- 05-01-PLAN.md -- Phase 5 plan specification

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new libraries, purely documentation + dead code removal
- Architecture: HIGH -- existing templates provide exact patterns to follow
- Pitfalls: HIGH -- scope is narrow and well-defined, risks are minimal

**Research date:** 2026-03-09
**Valid until:** 2026-04-09 (stable -- documentation patterns don't change)
