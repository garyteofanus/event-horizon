---
phase: 7
slug: retroactive-verification-and-requirements-cleanup
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-09
---

# Phase 7 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | none (Go convention) |
| **Quick run command** | `go test ./... -count=1` |
| **Full suite command** | `go test -v ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -count=1`
- **After every plan wave:** Run `go test -v ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 7-01-01 | 01 | 1 | TUI-01, TUI-02, TUI-03 | manual/doc | Verify file exists + content | N/A | ⬜ pending |
| 7-01-02 | 01 | 1 | DISP-01, DISP-02, DISP-03, DISP-04 | manual/doc | Verify file exists + content | N/A | ⬜ pending |
| 7-01-03 | 01 | 1 | INTR-01, INTR-02, INTR-03, INTR-04, ROBU-01 | manual/doc | Verify file exists + content | N/A | ⬜ pending |
| 7-01-04 | 01 | 1 | INTR-01 through ROBU-01 | manual/doc | Verify file exists + content | N/A | ⬜ pending |
| 7-01-05 | 01 | 1 | (none - dead code) | smoke | `go build -o /dev/null . && go test ./...` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements. This phase removes tests (dead code), it does not add them.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| VERIFICATION.md content accuracy | TUI-01 through ROBU-01 | Documentation review | Verify each requirement has evidence citing actual code/tests |
| SUMMARY.md completeness | INTR-01 through ROBU-01 | Documentation review | Verify summary matches git history and plan |
| REQUIREMENTS.md correctness | COPY-01 through KEY-02 | Documentation review | Verify entries exist in both list and traceability table |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
