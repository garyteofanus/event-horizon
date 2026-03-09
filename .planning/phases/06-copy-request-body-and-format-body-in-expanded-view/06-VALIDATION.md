---
phase: 6
slug: copy-request-body-and-format-body-in-expanded-view
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-09
---

# Phase 6 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | none (stdlib) |
| **Quick run command** | `go test ./... -count=1` |
| **Full suite command** | `go test ./... -count=1 -race` |
| **Estimated runtime** | ~3 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -count=1`
- **After every plan wave:** Run `go test ./... -count=1 -race`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 6-01-01 | 01 | 1 | COPY-01 | unit | `go test -run TestCopyBody -count=1` | ❌ W0 | ⬜ pending |
| 6-01-02 | 01 | 1 | COPY-02 | unit | `go test -run TestCopyFull -count=1` | ❌ W0 | ⬜ pending |
| 6-01-03 | 01 | 1 | COPY-03 | unit | `go test -run TestFlashCopied -count=1` | ❌ W0 | ⬜ pending |
| 6-01-04 | 01 | 1 | COPY-04 | unit | `go test -run TestFlashNoBody -count=1` | ❌ W0 | ⬜ pending |
| 6-01-05 | 01 | 1 | COPY-05 | unit | `go test -run TestFlashExpires -count=1` | ❌ W0 | ⬜ pending |
| 6-01-06 | 01 | 1 | KEY-01 | unit | `go test -run TestClearOnlyX -count=1` | ❌ W0 | ⬜ pending |
| 6-02-01 | 02 | 1 | FMT-01 | unit | `go test -run TestJSONFormat -count=1` | ❌ W0 | ⬜ pending |
| 6-02-02 | 02 | 1 | FMT-02 | unit | `go test -run TestFormatToggle -count=1` | ❌ W0 | ⬜ pending |
| 6-02-03 | 02 | 1 | FMT-03 | unit | `go test -run TestBodyLabel -count=1` | ❌ W0 | ⬜ pending |
| 6-02-04 | 02 | 1 | FMT-04 | unit | `go test -run TestJSONHighlight -count=1` | ❌ W0 | ⬜ pending |
| 6-03-01 | 01 | 1 | KEY-02 | unit | `go test -run TestFooterHelp -count=1` | ✅ (needs update) | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] Update `TestModelUpdateClearRequests` to use `'x'` instead of `'c'`
- [ ] Update `TestRenderViewShowsHelpFooter` for new keybinding text
- [ ] New test stubs for copy body/full, flash messages, format toggle, JSON highlighting

*Wave 0 creates test stubs; implementation fills them in during Wave 1.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| OSC52 clipboard works in terminal | COPY-01/02 | Requires real terminal with OSC52 support | Press `c`/`C` in TUI, paste in editor to verify |
| Flash message timing feels right | COPY-03/05 | Subjective UX timing | Copy a request, observe ~2s flash duration |
| JSON syntax colors are readable | FMT-04 | Visual appearance, theme-dependent | Expand a JSON request, verify distinct colors for keys/strings/numbers |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
