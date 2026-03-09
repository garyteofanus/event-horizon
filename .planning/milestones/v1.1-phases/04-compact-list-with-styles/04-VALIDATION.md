---
phase: 4
slug: compact-list-with-styles
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-06
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

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
| 4-01-01 | 01 | 1 | DISP-01 | unit | `go test ./...` | ✅ | ⬜ pending |
| 4-01-02 | 01 | 1 | DISP-02 | unit | `go test ./...` | ✅ | ⬜ pending |
| 4-01-03 | 01 | 1 | DISP-03 | unit | `go test ./...` | ✅ | ⬜ pending |
| 4-01-04 | 01 | 1 | DISP-04 | unit + manual | `go test ./...` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Method colors are perceptibly distinct in the active terminal theme/profile | DISP-02 | Terminal rendering differs by profile/theme and readability is ultimately visual | Run `go run .`, send GET/POST/DELETE/PUT/PATCH requests, and confirm the method colors are obviously distinct |
| Status colors are perceptibly distinct by response class | DISP-03 | Terminal output is user-perceived behavior, not just ANSI presence | Trigger representative 2xx/4xx/5xx statuses or use helper render checks plus a live run to confirm visual distinction |
| Adjacent entries are easy to distinguish at a glance | DISP-04 | Visual separation is subjective enough to require a quick live sanity check | Send several similar requests in sequence and confirm rows remain distinct on an 80-column terminal |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
