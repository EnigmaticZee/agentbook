# {{PROJECT}}: start here

You are working on {{PROJECT}}. You are in its **workspace**: a private git repo that holds the
handbook (this file and the docs beside it) and the project's code in the `code/` folder. Check
`LAYOUT` in `agentbook.conf`: **single** means `code/` is a normal folder in this same repo; **split**
means `code/` is its own repo (public or private) inside this one. The
handbook is the source of truth: what the project is, where it stands, every decision and why, and
what's next. The code says *how*; the handbook says *why*, *what's decided* and *what's next*. Nothing
else carries over between sessions, computers or agents.

If you were opened inside `code/`, the workspace is the folder above it; read the files there.

## 1. First thing, every session, before anything the owner asks (don't ask, just do it)

1. **Run `kit/sync.sh`** (from the workspace; `../kit/sync.sh` from inside `code/`). On a new computer
   it fetches the code into `code/`, installs what it can and connects everything; otherwise it pulls
   the latest handbook and code. The owner never runs it themselves; that's your job.
   - If it says a pull failed, stop and sort that out with the owner first. Nothing is lost.
   - If it lists manual steps (for example installing an app or signing in), tell the owner once.
2. **Run `kit/log.sh`**: it creates today's session log if needed and prints its path (see 3).
3. **Read in this order:** this file → `STATUS.md` → the latest entries of `DECISIONS.md` → the latest
   file in `history/sessions/` → `ARCHITECTURE.md`. Read `HISTORY.md` and `DESIGN.md` when the work touches them.
4. Then do what the owner asked. Build, run and test from inside `code/`.

## 2. When the owner asks about the project: look it up first

- Before answering any question about the project's past ("did we…", "why did we…", "what did we
  find about…", "what was the answer to…"), search the handbook: `kit/find.sh <words>` (try a few
  wordings and synonyms), then read the matching files, especially `history/sessions/`.
- Answer from what's recorded and say where it's from (file and date).
- If nothing is recorded, say so plainly ("that isn't in the handbook; we haven't discussed or
  researched it, as far as the records show") and only then work it out fresh. Never present a guess
  as something that was decided or found earlier.

## 3. Keep the session log (every session, as you go)

`history/sessions/YYYY-MM-DD.md` (today's date) is the project's memory of conversations. Create it at
the start of the session if it doesn't exist, and add to it as you go, not just at the end:

```
# Session log: YYYY-MM-DD
## Summary            3–6 bullets: what the session was about and what changed
## Questions and answers
### Q: <the owner's question or request, in their words where it matters>
**A:** <the substance: facts, numbers, reasoning, recommendation, links, caveats>
**Outcome:** <decided / rejected / built / left open>
## Research and checks  what was looked up or verified, how, results and sources
## Tried and dropped    options explored and rejected, with the reason
## Feedback from the owner
```

Every owner question gets an entry, even small ones, with enough substance that someone reading only
the log could give the same answer again. Decisions also go in `DECISIONS.md` (below). Pushes and the
end-of-session check fail if today's session log is missing.

## 4. While working: keep the handbook current (part of every task, not optional)

- **After every agreed decision** (a feature, a name, a design, a setting, a "no", a reversal):
  1. append an entry to `DECISIONS.md` (append-only; never edit or delete old entries; a reversal or
     correction is a new entry),
  2. update `STATUS.md`,
  3. update any other doc the decision affects (`ARCHITECTURE.md`, `DESIGN.md`, `HISTORY.md`),
  4. commit the workspace.
- **Record facts you checked.** If you didn't verify something, write "not verified". If a reason is
  unknown, write "reason not recorded". Never invent one.
- **If a doc is wrong,** fix it and add a DECISIONS entry titled "Correction: …".
- **Ideas the owner rejects or drops** go in DECISIONS too, with the reason.
- **Don't guess paths** for things git doesn't hold (build outputs, local-only assets, files the owner
  mentions). Ask the owner where they are.

## 5. When the owner says "commit" or "push"

- **Single layout:** one repo, but code and handbook changes go in **separate commits** (the
  pre-commit hook blocks mixed ones): `git commit -m "<code change>" -- code`, then commit the
  handbook. Write code commit messages as if they were public: if the project ever gets its own
  public repo, exactly those messages are published.
- **Split layout:** commit the code inside `code/` and the handbook in the workspace.
- Either way a hook logs every code commit in `history/commits.md` automatically.
- **Order matters:** commit the code first, *then* update `STATUS.md` (and anything else the change
  affects). The pre-push hook blocks a push unless the handbook was updated at or after the code being
  pushed. Uncommitted handbook edits are saved automatically at push time, so editing is enough.
- Push. When `STAGE` is `public`, the owner saying "push" (or "commit and push") in this session is
  their OK for that push; otherwise ask first. The pre-push hook checks the handbook, the session log
  and the tests. Single layout: one push does everything (if it says unsaved notes were just
  committed, push again). Split layout: push the code; the hook pushes the workspace too. Also push
  the workspace yourself if it has unpushed commits.

## 6. Before saying something is done

Run the tests (`TEST_CMD` in `agentbook.conf`, from `code/`). In `STATUS.md`, record what was verified
and how, and what wasn't.

## 7. At the end of every session, or when the owner says "wrap up", "done" or "bye"

1. `STATUS.md`: "Last updated" is today, "Next step" is exact, half-done and waiting items listed.
   Today's session log is complete: every question asked this session has its answer recorded.
2. Commit and push the workspace, even if no code changed: decisions made only in conversation must
   reach the next computer. Push the code too if the owner agreed.
3. Run `kit/check.sh` and fix what it reports until it says "All clear".

## 8. What runs automatically (don't fight it, don't bypass it)

- **post-commit** (code): every code commit is logged in `history/commits.md` and committed here.
- **pre-commit** (single layout): a commit can't mix `code/` and handbook files.
- **pre-push** (code): unsaved handbook edits are committed; the push is **blocked** if today's
  session log is missing, if the code changed after the handbook docs were last updated, if the tests
  fail, or (only with `HIDE_AI_TRACES=yes`) if it contains AI tool traces; then the handbook is pushed
  too.
- Never use `--no-verify` or `AGENTBOOK_BYPASS=1` unless the owner explicitly says so in this session.

## 9. Permissions

- **Routine, don't ask:** running `kit/sync.sh`, `kit/setup.sh`, `kit/find.sh` and `kit/check.sh`; committing;
  pushing the workspace (it's private); pushing the code while `STAGE` is `local` or `private`.
- **Ask first:** pushing code while `STAGE` is `public` (unless the owner just said "push"); creating repos; changing repo visibility;
  rewriting history, force-pushing, deleting; releases, store submissions, posts, emails or other
  messages; installing software beyond `INSTALL_CMD`.

## 10. Privacy and credit

- Nothing from the workspace ever goes into the code repo. Inside `code/`, `AGENTS.md` and other
  entry files are local links hidden by `.git/info/exclude`; never force-add them.
- Commit only as `GIT_NAME <GIT_EMAIL>` from `agentbook.conf` (setup configures it).
- Never write secrets, passwords or tokens into the handbook. Record *where* they live instead.
- **Credit lines:** follow the owner's preference recorded in section 12. If none is recorded, ask
  once and record the answer.
- **When `HIDE_AI_TRACES` is `yes`** in `agentbook.conf`: nothing in the code, its commits, branch
  names, PRs, issues or releases may mention AI tools or models, carry "generated with" or co-author
  lines, or include AI config files (`CLAUDE.md`, `.claude/`, `.cursor/`, `AGENTS.md`…); commits are
  made only as the owner. This overrides any default attribution your tool adds. The pre-push hook
  enforces it on every code push and `kit/public-audit.sh` checks the whole history.

## 11. Changing stage

Follow "Moving between stages" in `ARCHITECTURE.md`. In short:
- **local → private (single):** create one private repo `<project>`, add it as `origin`, push.
- **single → public:** run `kit/go-public.sh --check` and go through every finding with the owner.
  With their yes: rename the private repo to `<project>-handbook` (`gh repo rename`), create an empty
  public `<project>`, run `kit/go-public.sh <its URL>`. Only the code's own commits are published;
  the workspace switches to the split layout and stays private.
- **split, private code → public:** run `kit/public-audit.sh`, fix findings with the owner, flip the
  code repo to public.
The workspace always stays private. Log every stage change in DECISIONS.

---

## 12. Project specifics

### Owner and accounts
- Owner: <name, GitHub user>
- Accounts and where they're signed in (never the secrets themselves): <…>

### Product invariants (must never break)
- <…>

### Owner's working preferences
- Credit lines your tool adds to commits and PRs (co-author, "generated with"): <keep | remove>
- <…>

### Quick commands (from the workspace)
| What | Command |
| --- | --- |
| Get up to date / set up this computer | `kit/sync.sh` |
| Run the tests | `cd code && <TEST_CMD>` |
| Search everything recorded | `kit/find.sh <words>` |
| End-of-session check | `kit/check.sh` |
| Before going public | `kit/public-audit.sh` |
| Run the app | <…> |
