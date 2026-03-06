---
phase: 3
slug: tui-scaffolding-and-http-bridge
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-06
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | none — Go testing needs no config file |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test -v ./...` |
| **Estimated runtime** | ~3 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go build ./...` (compilation check)
- **After every plan wave:** Run `go test -v ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 3 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | TUI-02 | unit | `go test -run TestRequestChannel -v` | No — W0 | ⬜ pending |
| 03-01-02 | 01 | 1 | TUI-03 | unit | `go test -run TestFileOnlyLogging -v` | No — W0 | ⬜ pending |
| 03-01-03 | 01 | 1 | TUI-01 | manual | Manual: `go run .` then verify TUI appears | N/A | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `handler_test.go` — test that handleRequest sends RequestData to channel and logs to file
- [ ] `tui_test.go` — test model Update handles requestMsg correctly, test formatRequestLine output

*Existing infrastructure covers Go test framework — no install needed.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| TUI launches and takes over stdout | TUI-01 | Requires real terminal interaction | Run `go run .`, verify TUI screen appears instead of JSON |
| q/ctrl+c exits cleanly | TUI-01 | Requires real terminal interaction | Press q in running TUI, verify process exits with no hung goroutines |
| Request appears in TUI within 1s | TUI-02 | Requires real HTTP request + TUI observation | Send `curl localhost:8080/test` while TUI running, verify line appears |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 3s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
