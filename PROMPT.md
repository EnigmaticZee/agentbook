# Agentbook prompt

Paste everything below the line into any coding agent, from the project's folder (or an empty
folder for a brand-new project). It works at any stage: new, local, private, public, or a project
that already has notes.

---

Set up (or repair) the **agentbook workspace** for this project, so I can clone one repo on any
computer, open it in any coding agent, and carry on as if nothing happened, with the docs kept
current automatically.

## The fixed layout (don't change it)

```
<Project>/            the WORKSPACE: a private git repo (the handbook). This is what I clone and open.
  AGENTS.md            rules every agent reads first (CLAUDE.md etc. are links to it)
  README.md            how I work, in plain words
  STATUS.md, DECISIONS.md, HISTORY.md, ARCHITECTURE.md, DESIGN.md
  agentbook.conf       settings the scripts read
  kit/                 scripts and git hooks
  history/commits.md   automatic log of code commits
  code/                the CODE
```

Two layouts (`LAYOUT` in `agentbook.conf`):
- **single** (default for new projects): ONE private repo. `code/` is a normal folder in the
  workspace. Code and handbook changes are always committed separately (a hook enforces it), so the
  code's own history can later be split out cleanly. On GitHub: one private repo `<project>`.
- **split**: `code/` is its own repo (public or private) inside the workspace, ignored by it, fetched
  by setup from `CODE_REMOTE`. Used for projects that already had their own code repo, and after
  going public. On GitHub: code `<project>`, workspace `<project>-handbook` (always private).
- The handbook never reaches the code's public history in either layout.
- On a new computer, the agent's first step (`kit/sync.sh`) fetches what's missing, installs
  dependencies and connects the hooks. I never run scripts myself.

Get the kit from https://github.com/EnigmaticZee/agentbook (`git clone` it into a temporary folder,
or use the `agentbook` command if it's installed: `agentbook init` runs the installer for you). If you
can't get the kit, write scripts that match the spec at the end.

## Step 0: Look, then ask (don't create anything until I answer)

First find out on your own and tell me:
- Is this folder a workspace already (`agentbook.conf` + `kit/`)? Which kit version (`kit/VERSION`)?
  Does `kit/check.sh` pass?
- Or a code repo? Any remotes? Public or private (`gh repo view --json visibility`)?
- Any other notes: `AGENTS.md`, `CLAUDE.md`, `docs/`, `NOTES.md`, a `.research/` folder, a separate
  notes repo, notes committed inside the code repo?
- How is it built, run and tested today?

Then ask me, in one message, only what you couldn't work out:
1. Project name, and my GitHub account (if it's going on GitHub).
2. Stage now: **local**, **private** or **public**. (I usually start local and decide later.)
3. **Commit identity.** Default: my GitHub no-reply address.
4. **Credit lines:** keep or remove the co-author / "generated with" lines your tool adds? And do I
   want the optional `HIDE_AI_TRACES` guard (pushes blocked if the code mentions AI tools)? Default off.
5. Anything about the project you can't work out yourself.

Tell me your plan (which situation below) before doing it.

## Situations

**A. Brand-new project.** `install.sh <folder> "<Name>"` on an empty or new folder. Then Steps 1–5.

**B. Existing code, no handbook.** `install.sh <project folder> "<Name>"` adopts it in place: the
folder keeps its path, the code (with history and untracked files) moves into `code/`, the workspace
is created around it. Tell me before running it, because the code's path changes (apps that remember
the old path, like Xcode's recent projects, need to reopen it). Then Steps 1–5.

**C. Notes in another form** (a notes repo, `PROJECT_CONTEXT.md`, docs committed in the code). Create
the workspace (A or B) and move every fact into the right file. If the notes were their own repo, keep
its history: rename it to `<project>-handbook` and restructure it. If notes were **committed in the
code repo**, remove them there in a normal commit and tell me they stay in its history; if the code is
or may become public, offer a history rewrite (I decide).

**D. Workspace exists.** Run `kit/sync.sh` and `kit/check.sh`. If the kit is older, run
`install.sh --upgrade <workspace>` and merge new rules from the kit's `templates/AGENTS.md` and
`templates/README.md` without losing project specifics. Fill gaps (Step 2), then Step 5.

**E. Local → private GitHub.** Single layout: create one private repo `<project>`, add it as
`origin`, push, set `STAGE` and `HANDBOOK_REMOTE`, log it. (Split layout: two private repos,
`<project>` and `<project>-handbook`.)

**F. Going public.** Single layout: run `kit/go-public.sh --check` (it extracts exactly the history
that would be published, only commits that changed `code/`, and audits it). Walk me through every
finding. Check README and LICENSE in `code/`. With my yes: `gh repo rename <project>-handbook` for the
private repo, `gh repo create <owner>/<project> --public` (empty), then `kit/go-public.sh <its URL>`;
it switches the workspace to split. Split layout with private code: run `kit/public-audit.sh`,
history rewrites only with my yes, then flip the code repo to public. Either way set
`STAGE="public"`, update the handbook's remotes and README, log it. The workspace stays private.

**G. Back to private.** Flip the code repo's visibility, set `STAGE`, log it; remind me published
copies may exist. (It stays split: two private repos.)

**H. New computer.** I clone the workspace and open it; you run `kit/sync.sh` (AGENTS.md says so).

**I. Someone else's clone.** The code must build and test without the workspace; the hooks only exist
inside a workspace. If you're working for me and there's no workspace, ask me for it.

**J. Renaming.** Rename on GitHub (it redirects), update remotes, `agentbook.conf`, docs; log it.

**K. A hook blocks a push.** Do what its message says. Bypass only if I say so in this session:
`AGENTBOOK_BYPASS=1 git push`.

**L. Docs and code disagree.** The code is the truth for behaviour: fix the doc, add a "Correction".

**M. Retiring.** Archive both repos (don't delete) unless I say otherwise; write a final STATUS.

## Step 1: Learn the project

- Read the code, README, existing notes, git history, issues, and any conversation history you have.
- Build and run it, and run its tests. Write down what works, what doesn't, what's half-done.
- Collect every decision with its reason and the rejected options. If a reason isn't recorded
  anywhere, write "reason not recorded"; never invent one.

## Step 2: Write the handbook (in the workspace)

- `AGENTS.md`: the kit's generic rules (keep them) plus project specifics: owner and accounts (where
  they're signed in, never secrets), product invariants, my working preferences, quick commands.
- `README.md`: the kit's plain-words "how you work", with the real project name.
- `STATUS.md`: stage, what's live, current state, the exact next step, what's waiting on me, what was
  and wasn't verified, and a `Last updated: YYYY-MM-DD` line.
- `DECISIONS.md`: append-only, dated: decision, why, rejected.
- `HISTORY.md`: the story so far, including directions that were dropped.
- `history/sessions/YYYY-MM-DD.md`: one per day of work: every question I asked with the substance of
  the answer, research with sources and results, what was tried and dropped, my feedback. When adopting
  an existing project, backfill these from any conversation history you can read, so questions about
  the past never come up empty.
- `ARCHITECTURE.md`: repos and stages, files (paths inside `code/`), data flow, storage, external
  services, build, release, tests.
- `DESIGN.md` if there's a UI: colours, type, icons, patterns, how assets are made, rejected directions.
- `agentbook.conf`: stage, `CODE_DIR`, remotes, identity, `HIDE_AI_TRACES`, `AGENT_FILES`, `TEST_CMD`,
  `REQUIRE_CMDS`, `INSTALL_CMD`, `MANUAL_SETUP`.
Record facts you checked, not guesses.

## Step 3: Tests

- `TEST_CMD` must be one command that runs all tests from inside `code/`; `INSTALL_CMD` installs
  everything scriptable on a new computer; `MANUAL_SETUP` lists the rest.
- If there are no tests, add a real smoke test that runs the actual app the way a user would, and
  record what it covers and what it doesn't. Tests live in the code repo.
- Run them and record the result in STATUS.

## Step 4: Automation

`kit/setup.sh` (also run by `kit/sync.sh`) does it: ignores `code/`, links agent files, fetches or
starts the code, hides the links inside `code/`, sets `core.hooksPath` to `../kit/hooks`, sets
identity, installs dependencies. Confirm with `kit/check.sh`.

## Step 5: Prove it works

1. Clone the workspace fresh into a temporary folder (copy it if local only) and run `kit/sync.sh`.
2. Give a fresh agent with no context **only that folder**. Ask it 12 specific questions: what this is,
   the stage, what's live, the exact next step, three key decisions and why, one rejected idea and why,
   two things I asked about in past sessions and what the answers were, how to set up, run and test
   it, the rules for credit and pushing, and where accounts live.
3. Score its answers. Fix every gap in the handbook and repeat until it gets all of them right.
4. Run the tests in that fresh copy. Delete it afterwards.

## Step 6: Report back

Tell me where everything lives, what's automatic and what depends on agents following the rules, the
test results, the cold-read score, and what I do on a new computer (clone the workspace, open it).
Commit and push wherever the stage allows (ask first for public code).

## Kit spec (only if you can't get the kit)

- `kit/lib.sh`: `hb_from_kit` (workspace from the kit folder, code = workspace/`CODE_DIR`);
  `hb_locate` for hooks (code = repo top, workspace = its parent, must hold `agentbook.conf`);
  `hb_push` (on rejection, `pull --rebase` and retry; on conflict, abort and stop).
- `kit/setup.sh`: link each agent file in the workspace to `AGENTS.md`. Split: add `/<CODE_DIR>/` to
  the workspace `.gitignore`; if `code/` has no repo, clone `CODE_REMOTE` (or `git init`); exclude
  agent files and AI tool folders in `code/.git/info/exclude`; `core.hooksPath=../kit/hooks` on the
  code repo. Single: `code/` must not be ignored or be its own repo; exclude `/code/<agent files>`
  and `/code/<AI tool folders>` in the workspace's `.git/info/exclude`; `core.hooksPath=kit/hooks` on
  the workspace. Both: link `code/<agent file>` to `../AGENTS.md`; refuse tracked agent files or
  another hooksPath; identity; `REQUIRE_CMDS`; `INSTALL_CMD` unless `--no-install`.
- `kit/hooks/pre-commit` (single only): block a commit whose staged paths include both `code/…` and
  anything else.
- `kit/go-public.sh --check | <empty repo URL> [--stage private]` (single only): clone the workspace,
  `filter-branch --subdirectory-filter code --prune-empty`, audit the result; with a URL (and a clean
  audit or `AGENTBOOK_APPROVED=1`): push it, untrack `code/`, ignore it, turn `code/` into a clone of the
  new repo with `init`+`fetch`+`reset --mixed` (keeps local files), set `LAYOUT=split`,
  `CODE_REMOTE`, `STAGE`, run setup, commit.
- `kit/sync.sh`: commit leftover workspace edits, then `pull --rebase` the workspace (abort and stop on
  a conflict); if `code/` is missing run `setup.sh`, else
  `pull --ff-only` it and run `setup.sh --no-install`; on a failed pull, stop, explain, lose nothing.
- `kit/hooks/post-commit` (code; in single, only commits touching `code/`): append date, short hash, branch, subject and changed files to
  `history/commits.md`; commit only that file in the workspace; skip during rebase; never fail.
- `kit/hooks/pre-push`: nothing outside a workspace, with `AGENTBOOK_BYPASS=1`, or when no pushed commit
  touches the code; find the oldest code commit being pushed; all checks below apply to code commits; only with `HIDE_AI_TRACES=yes`, block if any pushed commit's message, author or
  committer identity, file paths or added lines suggest an AI tool (tool names, credit phrases,
  co-author lines, AI config files); auto-commit workspace edits; block if today's session log is
  missing; block if the last workspace commit that changed
  anything outside `history/` and `code/` is older (single: as of the pushed tip; if notes were just
  auto-committed, ask to push again); run `TEST_CMD` in `code/` and block on failure; split: push the
  workspace if it has a remote.
- `kit/log.sh`: create today's `history/sessions/YYYY-MM-DD.md` from a skeleton if missing; print its path.
- `kit/find.sh <words…>`: case-insensitive search of every `.md` in the workspace (not `code/` or
  `kit/`), printing file:line hits, or "(nothing recorded)"; exit 1 when nothing matches.
- `kit/check.sh`: both repos clean and pushed (or no remote), STATUS dated today, today's
  `history/sessions/YYYY-MM-DD.md` exists, handbook newer than
  the last code commit, required docs present, `.gitignore`, links and hooks in place, tests pass.
- `kit/public-audit.sh [--repo <dir>]` (read-only; split: `code/`; single: the extracted code history): across files and history, reports workspace or agent
  files, `.env`/keys, tool or assistant credit, commit identities, tool mentions, home paths,
  personal emails, likely secrets, files over 5 MB, missing LICENSE/README.
- `install.sh <folder> "<Name>" [--split]`: new folder or plain files → single layout (files into
  `code/`, committed as "Initial import" after the handbook commit); `--split` → split; existing
  code repo → split, moved into `code/` in place; `--upgrade` replaces `kit/`.
