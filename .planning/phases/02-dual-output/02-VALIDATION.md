---
phase: 2
slug: dual-output
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-06
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib, go1.25+) |
| **Config file** | none — `go test` works out of the box |
| **Quick run command** | `go test ./... -v -count=1` |
| **Full suite command** | `go test ./... -v -count=1 -race` |
| **Estimated runtime** | ~2 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -v -count=1`
- **After every plan wave:** Run `go test ./... -v -count=1 -race`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 2-01-01 | 01 | 0 | OUT-02, OUT-03 | unit | `go test -run TestDualOutput -v -count=1` | No — W0 | pending |
| 2-01-02 | 01 | 1 | OUT-02 | integration | `go test -run TestDualOutput -v -count=1` | No — W0 | pending |
| 2-01-03 | 01 | 1 | OUT-03 | unit | `go test -run TestLogFilePath -v -count=1` | No — W0 | pending |

*Status: pending / green / red / flaky*

---

## Wave 0 Requirements

- [ ] `main_test.go` — test stubs for OUT-02 (dual output) and OUT-03 (LOG_FILE config)
- [ ] Test helper: create temp file, build `io.MultiWriter`, verify both destinations receive identical JSON

*Existing infrastructure covers Go test framework.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Visual stdout output during server run | OUT-02 | Requires human observation of terminal output | Start server, send curl request, observe JSON in terminal |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
