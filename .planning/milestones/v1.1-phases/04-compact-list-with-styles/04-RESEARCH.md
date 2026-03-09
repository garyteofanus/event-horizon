# Phase 4: Compact List with Styles - Research

**Researched:** 2026-03-06
**Domain:** Bubble Tea v2 request-list styling with Lip Gloss row rendering
**Confidence:** HIGH

## Summary

Phase 4 should stay narrowly scoped inside the existing TUI rendering path. Phase 3 already established the core seams this phase needs: `RequestData` carries the required fields, `formatRequestLine` already normalizes timestamp/path/status/timing into a one-line string, and `renderView` makes the display directly testable. The missing work is presentation: color-coding method and status segments and adding visual separation between adjacent rows without breaking the compact one-line layout.

The recommended implementation is to add `github.com/charmbracelet/lipgloss` as a direct dependency and centralize styling in `tui.go`. Keep `handler.go` focused on data capture. Render each request row by composing styled segments for timestamp, method, path, status, and response time, then wrap the row with a lightweight separator strategy such as a left border plus bottom margin or a faint divider between rows. This satisfies DISP-01 through DISP-04 while preserving the existing append-at-bottom behavior and testability.

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| DISP-01 | Compact one-line row: timestamp, method, path, status code | Existing `RequestData` + `formatRequestLine` already provide the required fields |
| DISP-02 | HTTP methods are color-coded | Lip Gloss foreground styles can map method names to fixed colors |
| DISP-03 | Status codes are color-coded by range | Lip Gloss styles can map status ranges to green/yellow/red |
| DISP-04 | Visible separation between entries | Row container styling or divider lines can distinguish adjacent entries without expanding into multi-line cards |

## Current Codebase Fit

### Existing Seams

- `handler.go`
  - `RequestData` already includes `Timestamp`, `Method`, `URI`, `Status`, and `ResponseTime`
  - `formatRequestLine` already truncates long URIs and formats `<1ms`
- `tui.go`
  - `model.requests` already stores the rolling list
  - `renderView` already builds header, content, separators, and status line in a testable helper
  - auto-scroll logic already clips to visible height
- `tui_test.go`
  - existing render tests already assert header, empty state, request rows, and status count

### Recommended Boundary

- Keep raw request formatting logic in `handler.go` only if it remains data-oriented
- Move row presentation decisions into `tui.go`
- Prefer a new helper such as `renderRequestRow(RequestData, width int) string` over expanding `formatRequestLine` into ANSI-aware rendering

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `charm.land/bubbletea/v2` | v2.0.1 | Existing TUI framework | Already adopted in Phase 3 |
| `github.com/charmbracelet/lipgloss` | current stable | Styling, colors, borders, spacing | Standard Charm styling layer for Bubble Tea apps |

### Why Lip Gloss

- Purpose-built for Bubble Tea string rendering
- Lets the code express colors, padding, borders, and dimmed text without hand-writing ANSI sequences
- Keeps render logic deterministic enough for unit tests via `renderView`

## Architecture Patterns

### Pattern 1: Styled Segment Composition

Render each row from multiple styled segments rather than applying one style to the full raw line.

Recommended row structure:

`HH:MM:SS  METHOD  /path  STATUS  TIME`

Suggested segment responsibilities:

- timestamp: dim/faint neutral color
- method: bold + per-method foreground color
- path: normal foreground, truncation remains in formatter/helper
- status: bold + range-based foreground color
- time: dim neutral suffix

### Pattern 2: Row-Level Separation Without Breaking Compactness

Use one of these lightweight separators:

1. Left border per row using a muted color
2. Bottom margin of one line with a faint divider
3. Alternating subtle background or foreground treatment

**Primary recommendation:** left border plus a single-space gutter. It preserves one-line readability, works in narrow terminals, and avoids doubling vertical height the way blank-line spacing would.

### Pattern 3: Display Helper Split

Recommended helpers inside `tui.go`:

- `renderRequestRow(r RequestData, width int) string`
- `methodStyle(method string) lipgloss.Style`
- `statusStyle(code int) lipgloss.Style`
- `separatorStyle()` or `rowStyle()`

This keeps `renderView` orchestration-focused.

## Style Recommendations

### Method Colors

| Method | Color |
|--------|-------|
| GET | green |
| POST | blue |
| DELETE | red |
| PUT | yellow |
| PATCH | cyan |
| Other | neutral gray/white |

### Status Colors

| Range | Color |
|-------|-------|
| 2xx | green |
| 3xx | cyan or neutral |
| 4xx | yellow |
| 5xx | red |
| Other | neutral |

3xx is not required by the roadmap, but assigning a neutral or cyan fallback avoids unstyled gaps.

### Separator Strategy

Use a muted left border on each rendered row. If that alone feels too subtle, combine it with a faint bottom divider between rows while keeping each request itself to one visible content line.

## Testing Strategy

### Automated

- Extend `tui_test.go` to assert rendered output still contains:
  - one-line request rows
  - correct request count
  - visible separators between adjacent requests
- Add focused unit tests for helper behavior:
  - method color mapping
  - status color mapping
  - row rendering keeps timestamp/method/path/status on one line
  - long paths remain truncated
- Prefer testing `renderView` and helper outputs directly instead of driving a full Bubble Tea program

### Manual

- Run `go run .`
- Send GET, POST, DELETE, PUT, and PATCH requests
- Confirm distinct colors are perceptibly different in the terminal profile actually used
- Confirm adjacent entries remain readable at a glance on a typical 80-column terminal

## Common Pitfalls

### Pitfall 1: Embedding ANSI logic in `handler.go`

This couples presentation to transport data and makes future interactive features harder. Keep styling in `tui.go`.

### Pitfall 2: Using only color to distinguish entries

Requirement DISP-04 still needs separation independent of method/status color. Rows should remain distinct even when adjacent entries share the same color range.

### Pitfall 3: Making separators too tall

Blank lines between every request reduce information density and work against the “compact list” goal.

### Pitfall 4: Not accounting for width when styled

Styled strings include escape sequences, so any truncation or width decisions should happen before styling, not after.

## Implementation Recommendation

Plan this phase as a single execution plan in one wave:

- add Lip Gloss as a direct dependency
- refactor `tui.go` to render styled request rows and visual separators
- expand `tui_test.go` with render-level assertions
- keep `handler.go` unchanged unless a tiny helper extraction materially simplifies rendering

## Validation Architecture

### Why this phase can be Nyquist-compliant

- Rendering is already centralized in `renderView`
- Most behaviors are deterministic string output
- Only color perception needs a short manual check

### Required automated checks

- `go test ./...`
- targeted `tui_test.go` coverage for row rendering, color mapping helpers, and separators

### Manual-only check

- visual confirmation that method/status colors are distinguishable in a real terminal session

## Sources

- Bubble Tea v2 package docs: https://pkg.go.dev/charm.land/bubbletea/v2
- Bubble Tea realtime example: https://github.com/charmbracelet/bubbletea/tree/main/examples/realtime
- Realtime example source: https://raw.githubusercontent.com/charmbracelet/bubbletea/main/examples/realtime/main.go
- Lip Gloss package docs: https://pkg.go.dev/github.com/charmbracelet/lipgloss
