# event-horizon

## What This Is

A small Go HTTP sink that accepts any request, persists it as structured JSON, and shows the traffic live in a keyboard-driven terminal UI.

## Core Value

Every request that hits the server is reliably captured and immediately inspectable.

## Current State

- v1.1 shipped on 2026-03-09.
- The runtime is a multi-file Go app (`main.go`, `handler.go`, `tui.go`) using Bubble Tea v2 and Lip Gloss v2.
- The current experience includes live request streaming, styled compact rows, inline detail expansion, clipboard copy, JSON formatting/highlighting, and file-based JSON logging.
- The current Go codebase is 2,284 LOC across 6 `.go` files.

## Next Milestone Goals

- Define fresh requirements with `$gsd-new-milestone`.
- Decide whether the next release prioritizes safety and robustness (`SAFE-01`, `SAFE-02`, `ROBU-02`, `ROBU-03`, `ROBU-04`) or power-user inspection features (`FILT-01`, `FILT-02`, `COMP-01`, `LOG-07`, `LOG-08`, `LOG-09`, `LOG-10`).
- Optionally clean up residual docs and process debt by refreshing the stale Phase 06 verification wording and finishing Nyquist validation for phases 03, 04, 06, and 07.

## Requirements

### Validated

- [x] Structured JSON capture of every request, dual-output logging, and empty 200 OK responses — v1.0
- [x] Live Bubble Tea TUI with request streaming and file-only JSON logging after stdout handoff — v1.1
- [x] Styled compact request rows with color coding and visual separation — v1.1
- [x] Interactive navigation, inline detail expansion, clear behavior, footer help, and resize safety — v1.1
- [x] Clipboard copy, flash feedback, JSON pretty-printing, syntax highlighting, and format toggle — v1.1

### Active

- None yet. The next milestone should define requirements from a clean slate.

### Out of Scope

- Echo response body — the server is an inspection sink, not a request mirror
- Route-based behavior — all paths are handled identically
- Authentication or authorization — run behind a proxy if access control is needed
- TLS/HTTPS termination — use a reverse proxy
- Mouse support — keyboard-first operation is sufficient for the terminal UI
- Split-pane layouts — compact rows plus inline expansion stay simpler

## Context

- Go 1.25.0, module name `event-horizon`
- Terminal UI stack: `charm.land/bubbletea/v2` and `charm.land/lipgloss/v2`
- Structured JSON logs remain the source of truth in `requests.log`
- Milestone v1.1 covered phases 3-7 and is archived under `.planning/milestones/`

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Use `slog` for request logging | Standard library structured logging kept the sink simple | ✓ Good |
| Run Bubble Tea on the main goroutine and HTTP in the background | The TUI needs terminal ownership and clean shutdown semantics | ✓ Good |
| Keep JSON logging file-only while the TUI owns stdout | Prevents human-readable TUI output from mixing with machine-readable logs | ✓ Good |
| Keep the TUI in the main terminal buffer | Preserves a log-tail feel and exits back to a clean shell | ✓ Good |
| Use compact rows with inline expansion | Keeps the default view dense while still exposing full request detail on demand | ✓ Good |
| Use OSC52 clipboard support via `tea.SetClipboard` | Adds copy support without extra platform-specific dependencies | ✓ Good |
| Format JSON bodies by default and allow a raw toggle | Improves readability without hiding the original payload | ✓ Good |

## Archive

<details>
<summary>Pre-v1.1 planning snapshot</summary>

### Former milestone framing

- Current Milestone: v1.1 TUI Log Viewer
- Goal: Replace raw stdout JSON with a live terminal UI for inspecting requests in real time
- Original target features: live TUI, compact one-line rows, expand-to-detail, clear keybind, visual separation, persistent JSON log file

### Former in-progress view

- The v1.1 snapshot was last updated on 2026-03-06 after Phase 3 completion.
- At that point, the project had already launched the Bubble Tea shell, but styling, interaction polish, clipboard copy, JSON formatting, and retrospective verification were still pending.

</details>

---
*Last updated: 2026-03-09 after v1.1 milestone completion*
