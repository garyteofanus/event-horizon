---
phase: quick-6
plan: 01
type: execute
wave: 1
depends_on: []
files_modified: []
autonomous: true
requirements: [RELEASE-01]
must_haves:
  truths:
    - "GitHub release v1.1.0 exists on garyteofanus/event-horizon"
    - "Pre-built binaries for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64 are attached"
    - "Each binary is named clearly with OS and architecture"
  artifacts:
    - path: "GitHub release v1.1.0"
      provides: "Downloadable binaries for 4 platform targets"
  key_links:
    - from: "git tag v1.1.0"
      to: "gh release create v1.1.0"
      via: "tag reference"
      pattern: "gh release create v1.1.0"
---

<objective>
Create a GitHub release (v1.1.0) for event-horizon with cross-compiled Go binaries for 4 platforms.

Purpose: Let users download pre-built binaries without needing Go installed.
Output: GitHub release with 4 binary assets attached.
</objective>

<execution_context>
@/Users/teo/.claude/get-shit-done/workflows/execute-plan.md
@/Users/teo/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@CLAUDE.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Cross-compile binaries and create GitHub release</name>
  <files>None (no repo files modified -- builds are temporary artifacts)</files>
  <action>
1. Create a temporary build directory.

2. Cross-compile the Go binary for 4 targets. The module is `event-horizon` (local module name). Build from the project root directory `/Users/teo/Developer/event-horizon`. Use the naming convention `event-horizon-{os}-{arch}`:
   - GOOS=linux GOARCH=amd64 go build -o {tmpdir}/event-horizon-linux-amd64 .
   - GOOS=linux GOARCH=arm64 go build -o {tmpdir}/event-horizon-linux-arm64 .
   - GOOS=darwin GOARCH=amd64 go build -o {tmpdir}/event-horizon-darwin-amd64 .
   - GOOS=darwin GOARCH=arm64 go build -o {tmpdir}/event-horizon-darwin-arm64 .

3. Create a git tag v1.1.0 on the current HEAD commit and push it:
   - git tag v1.1.0
   - git push origin v1.1.0

4. Create the GitHub release with all 4 binaries attached:
   - gh release create v1.1.0 {tmpdir}/event-horizon-linux-amd64 {tmpdir}/event-horizon-linux-arm64 {tmpdir}/event-horizon-darwin-amd64 {tmpdir}/event-horizon-darwin-arm64 --title "v1.1.0 - TUI Log Viewer" --notes "## What's New\n\nEvent-horizon v1.1.0 adds a real-time TUI log viewer built with Bubble Tea.\n\n### Features\n- Real-time terminal UI with color-coded HTTP methods and status codes\n- Keyboard navigation (j/k, arrows, enter/space)\n- Expanded detail view for individual requests\n- Copy request body or full request via OSC52 clipboard\n- JSON body formatting with syntax highlighting\n- Structured JSON logging to file\n\n### Binaries\n\nDownload the binary for your platform:\n| File | OS | Arch |\n|------|------|------|\n| event-horizon-linux-amd64 | Linux | x86_64 |\n| event-horizon-linux-arm64 | Linux | ARM64 |\n| event-horizon-darwin-amd64 | macOS | x86_64 (Intel) |\n| event-horizon-darwin-arm64 | macOS | ARM64 (Apple Silicon) |\n\nMake it executable: \`chmod +x event-horizon-*\`\nRun: \`./event-horizon-darwin-arm64\` (listens on :8080 by default)"

5. Clean up the temporary build directory.
  </action>
  <verify>
    <automated>gh release view v1.1.0 --repo garyteofanus/event-horizon --json tagName,assets --jq '{tag: .tagName, assets: [.assets[].name] | sort}'</automated>
  </verify>
  <done>GitHub release v1.1.0 exists with 4 binaries: event-horizon-linux-amd64, event-horizon-linux-arm64, event-horizon-darwin-amd64, event-horizon-darwin-arm64</done>
</task>

</tasks>

<verification>
- `gh release view v1.1.0` shows the release with correct title and notes
- 4 binary assets are listed and downloadable
- Tag v1.1.0 points to current HEAD
</verification>

<success_criteria>
- GitHub release v1.1.0 is published on garyteofanus/event-horizon
- All 4 platform binaries are attached as release assets
- Release notes describe the TUI Log Viewer milestone features
</success_criteria>

<output>
After completion, create `.planning/quick/6-add-github-release-with-go-binary-builds/6-SUMMARY.md`
</output>
