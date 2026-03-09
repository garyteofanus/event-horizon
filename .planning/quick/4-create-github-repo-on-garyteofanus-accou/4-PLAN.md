---
phase: quick
plan: 4
type: execute
wave: 1
depends_on: []
files_modified: []
autonomous: true
must_haves:
  truths:
    - "GitHub repo garyteofanus/event-horizon exists and is accessible"
    - "Local repo has origin remote pointing to garyteofanus/event-horizon"
    - "All commits and branches are pushed to remote"
  artifacts: []
  key_links:
    - from: "local git repo"
      to: "github.com/garyteofanus/event-horizon"
      via: "git remote origin"
---

<objective>
Create a GitHub repository on the garyteofanus account for the event-horizon project and push the existing local repository to it.

Purpose: Publish the project to GitHub for remote access and collaboration.
Output: Public GitHub repo with all code pushed.
</objective>

<execution_context>
@/Users/teo/.claude/get-shit-done/workflows/execute-plan.md
@/Users/teo/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Create GitHub repo and push</name>
  <files></files>
  <action>
1. Create a new public GitHub repo using gh CLI:
   `gh repo create event-horizon --public --source=. --remote=origin --push`
   This creates the repo on garyteofanus account (the authenticated user), adds the origin remote, and pushes the current branch in one command.

2. If the above fails because the remote already exists or needs separate steps:
   - `gh repo create garyteofanus/event-horizon --public`
   - `git remote add origin https://github.com/garyteofanus/event-horizon.git`
   - `git push -u origin main`

3. Verify the repo is accessible: `gh repo view garyteofanus/event-horizon`
  </action>
  <verify>
    <automated>gh repo view garyteofanus/event-horizon --json name,owner,url,defaultBranchRef 2>&1 | head -10</automated>
  </verify>
  <done>GitHub repo garyteofanus/event-horizon exists, origin remote is set, all commits are pushed to main branch</done>
</task>

</tasks>

<verification>
- `git remote -v` shows origin pointing to github.com/garyteofanus/event-horizon
- `gh repo view garyteofanus/event-horizon` returns repo info
- `git log origin/main` matches local main branch
</verification>

<success_criteria>
- GitHub repository garyteofanus/event-horizon is publicly accessible
- Local repo origin remote points to the GitHub repo
- All local commits are present on remote main branch
</success_criteria>

<output>
After completion, create `.planning/quick/4-create-github-repo-on-garyteofanus-accou/4-SUMMARY.md`
</output>
