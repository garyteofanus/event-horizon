---
phase: 1
slug: structured-logging-core
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-06
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) |
| **Config file** | none — Go testing needs no config |
| **Quick run command** | `go test -v -run TestX ./...` |
| **Full suite command** | `go test -v -race ./...` |
| **Estimated runtime** | ~2 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test -v -race ./...`
- **After every plan wave:** Run `go test -v -race ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 0 | QA-01 | smoke | `go list -m all` | N/A | ⬜ pending |
| 01-01-02 | 01 | 0 | LOG-01..06, SRV-01..02 | unit | `go test -v -race ./...` | ❌ W0 | ⬜ pending |
| 01-01-03 | 01 | 1 | LOG-01, OUT-01, QA-02 | unit | `go test -v -run TestLogJSON ./...` | ❌ W0 | ⬜ pending |
| 01-01-04 | 01 | 1 | LOG-02..06 | unit | `go test -v -run TestLogFields ./...` | ❌ W0 | ⬜ pending |
| 01-01-05 | 01 | 1 | SRV-01, SRV-02 | unit | `go test -v -run TestEmptyResponse ./...` | ❌ W0 | ⬜ pending |
| 01-01-06 | 01 | 1 | SRV-03 | unit | `go test -v -run TestPortConfig ./...` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `main_test.go` — test stubs for LOG-01 through SRV-03, QA-01
- [ ] Handler extraction — refactor anonymous closure to named function accepting logger for testability
- [ ] Test helpers — `bytes.Buffer` + `slog.NewJSONHandler` capture pattern, `httptest` request builders

*Existing infrastructure covers: Go module (`go.mod`), Go toolchain*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| LogAttrs with typed constructors only | QA-02 | Static analysis / code review — cannot assert API choice at runtime | `grep -n 'logger\.\(Info\|Warn\|Error\|Debug\)(' main.go` should return 0 matches |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
