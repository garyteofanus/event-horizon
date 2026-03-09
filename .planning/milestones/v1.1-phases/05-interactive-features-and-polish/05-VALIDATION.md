---
phase: 5
slug: interactive-features-and-polish
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-06
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 05-01-01 | 01 | 1 | INTR-01, INTR-02, INTR-03 | unit | `go test ./...` | ✅ | ⬜ pending |
| 05-01-02 | 01 | 1 | INTR-04, ROBU-01 | unit | `go test ./...` | ✅ | ⬜ pending |
| 05-01-03 | 01 | 1 | INTR-01, INTR-02, INTR-03, INTR-04, ROBU-01 | manual | `go run .` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Cursoring, expansion visuals, footer clarity, and resize behavior in a real terminal | INTR-01, INTR-02, INTR-04, ROBU-01 | Terminal rendering and interaction fidelity are not fully provable by string/unit tests alone | Run `go run .`, send several requests, navigate with `j/k` and arrows, expand/collapse with the documented key, resize the terminal aggressively, and confirm the display stays intact |
| Clear removes visible TUI entries while preserving file logging | INTR-03 | Requires verifying UI state and persisted log file together | Run `go run .`, send requests, trigger clear, confirm the TUI empties, then inspect `requests.log` to confirm prior entries remain |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 10s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
