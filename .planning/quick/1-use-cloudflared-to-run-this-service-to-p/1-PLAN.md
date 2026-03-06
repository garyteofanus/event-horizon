---
phase: quick
plan: 1
type: execute
wave: 1
depends_on: []
files_modified: [CLAUDE.md]
autonomous: true
requirements: [QUICK-01]

must_haves:
  truths:
    - "Echo server is accessible from public internet via cloudflared tunnel"
    - "CLAUDE.md documents the exact commands to start server + tunnel"
  artifacts:
    - path: "CLAUDE.md"
      provides: "Quick-start documentation for cloudflared tunnel"
      contains: "cloudflared tunnel"
  key_links:
    - from: "cloudflared"
      to: "localhost:8080"
      via: "cloudflared quick tunnel"
      pattern: "cloudflared tunnel --url"
---

<objective>
Run the echo server publicly via cloudflared quick tunnel and document the process in CLAUDE.md so it can be repeated quickly.

Purpose: Make the echo server accessible from the public internet using Cloudflare's free quick tunnel, and leave a runbook for next time.
Output: Running public tunnel + updated CLAUDE.md with instructions.
</objective>

<execution_context>
@/Users/teo/.claude/get-shit-done/workflows/execute-plan.md
@/Users/teo/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@CLAUDE.md
@main.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Start echo server and cloudflared tunnel, document in CLAUDE.md</name>
  <files>CLAUDE.md</files>
  <action>
1. Start the echo server in the background: `go run main.go &`
2. Confirm it is listening on :8080 with a quick curl to localhost:8080
3. Start cloudflared quick tunnel: `cloudflared tunnel --url http://localhost:8080`
   - This will print a public URL like `https://xxx-xxx-xxx.trycloudflare.com`
   - The tunnel runs in the foreground; start it in background and capture the URL from its output
4. Test the public URL with curl to confirm it echoes back
5. Update CLAUDE.md to add a "Public Access (cloudflared)" section with these commands:

```
## Public Access (cloudflared)

To expose the echo server publicly via a Cloudflare quick tunnel:

1. Start the server: `go run main.go &`
2. Start the tunnel: `cloudflared tunnel --url http://localhost:8080`
3. Use the printed `https://...trycloudflare.com` URL to access the server from anywhere
4. To stop: kill both processes (`fg` then Ctrl-C, or `pkill -f "go run main.go"` and `pkill cloudflared`)

No Cloudflare account required. The URL changes each time you restart the tunnel.
```

Note: After documenting, stop both processes (kill the background server and cloudflared) so the user can start them manually when needed.
  </action>
  <verify>
    <automated>grep -q "cloudflared" CLAUDE.md && echo "PASS: CLAUDE.md updated" || echo "FAIL"</automated>
  </verify>
  <done>CLAUDE.md contains clear cloudflared quick-tunnel instructions. User can copy-paste commands to go public in seconds.</done>
</task>

<task type="checkpoint:human-verify" gate="informational">
  <what-built>Added cloudflared tunnel documentation to CLAUDE.md</what-built>
  <how-to-verify>
    1. Run: go run main.go &
    2. Run: cloudflared tunnel --url http://localhost:8080
    3. Copy the trycloudflare.com URL from the output
    4. Open it in a browser or curl it from another device
    5. Confirm you see a 200 OK response
  </how-to-verify>
  <resume-signal>Confirm the tunnel works from outside your network</resume-signal>
</task>

</tasks>

<verification>
- CLAUDE.md contains a "Public Access" section with cloudflared commands
- Instructions are copy-paste ready
</verification>

<success_criteria>
- CLAUDE.md documents the full cloudflared quick-tunnel workflow
- A developer can go from zero to public URL in under 30 seconds following the docs
</success_criteria>

<output>
After completion, create `.planning/quick/1-use-cloudflared-to-run-this-service-to-p/1-SUMMARY.md`
</output>
