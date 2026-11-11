#!/bin/bash
# Tests the kit through every stage on throwaway repos. Local "remotes" are bare repos, so nothing
# touches GitHub. Usage: test/run.sh   (exits 1 on the first failure)
set -uo pipefail

KIT="$(cd "$(dirname "$0")/.." && pwd -P)"
T="$(mktemp -d "${TMPDIR:-/tmp}/agentbook-test.XXXXXX")"
T="$(cd "$T" && pwd -P)"
trap 'rm -rf "$T"' EXIT
# A throwaway global git config, so the test never reads or changes yours.
export GIT_CONFIG_NOSYSTEM=1 GIT_CONFIG_GLOBAL="$T/gitconfig"
printf '[user]\n\tname = Tester\n\temail = t@example.invalid\n[init]\n\tdefaultBranch = main\n' > "$GIT_CONFIG_GLOBAL"

pass=0
ok()   { echo "  ✓ $1"; pass=$((pass + 1)); }
die()  { echo "  ✗ $1"; [ -n "${2:-}" ] && echo "$2" | sed 's/^/      /'; exit 1; }
tick() { sleep 1.1; }   # hooks compare commit times in whole seconds
conf() { (. "$KIT/kit/lib.sh"; HB="$1"; hb_set_conf "$2" "$3"); }   # conf <workspace> <key> <value>

echo "1. Two-repo layout (--split): new project, started locally"
W="$T/MyApp"
"$KIT/install.sh" "$W" "My App" --split >/dev/null || die "install failed"
C="$W/code"
[ -d "$W/.git" ] && [ -d "$C/.git" ] || die "workspace or code repo missing"
[ -f "$W/AGENTS.md" ] && [ -L "$W/CLAUDE.md" ] || die "agent files missing in the workspace"
[ -L "$C/AGENTS.md" ] && grep -q "My App" "$C/AGENTS.md" || die "agent link inside code/ missing or broken"
[ "$(git -C "$C" config core.hooksPath)" = "../kit/hooks" ] || die "hooks not set"
grep -qx "/code/" "$W/.gitignore" || die "workspace doesn't ignore code/"
git -C "$W" ls-files | grep -q '^code/' && die "workspace tracks code/"
[ -L "$W/GEMINI.md" ] && [ -L "$C/GEMINI.md" ] || die "GEMINI.md not linked by default"
grep -q '^HIDE_AI_TRACES="no"' "$W/agentbook.conf" || die "HIDE_AI_TRACES isn't off by default"
ok "workspace created: handbook at the top, code/ inside, linked and hidden"

echo "hello" > "$C/app.txt"
mkdir -p "$W/history/sessions"
LOG="$W/history/sessions/$(date +%F).md"
conf "$W" TEST_CMD "test -f app.txt"
conf "$W" GIT_NAME "Owner"; conf "$W" GIT_EMAIL "owner@users.noreply.github.com"
"$W/kit/setup.sh" --no-install >/dev/null || die "setup re-run failed"
git -C "$W" commit -qam "Configure"
tick
git -C "$C" add -A && git -C "$C" commit -qm "First code commit"
[ "$(git -C "$C" log -1 --format=%ae)" = "owner@users.noreply.github.com" ] || die "identity not applied"
git -C "$C" show --name-only --format= HEAD | grep -qE 'AGENTS|CLAUDE' && die "git add -A picked up agent files"
grep -q "First code commit" "$W/history/commits.md" || die "commit not logged"
[ -z "$(git -C "$W" status --porcelain)" ] || die "log not committed in the workspace"
ok "git add -A in code/ ignores agent files; commit logged and saved in the workspace"

out="$("$W/kit/check.sh" 2>&1)"
echo "$out" | grep -q "latest code commit is newer" || die "check missed a stale handbook" "$out"
echo "$out" | grep -q "no remote yet" || die "check didn't treat local-only correctly" "$out"
tick; echo "- worked on app" >> "$W/STATUS.md"; git -C "$W" commit -qam "Status"
out="$("$W/kit/check.sh" 2>&1)"
echo "$out" | grep -q "no session log for today" || die "check didn't require a session log" "$out"
[ "$("$W/kit/log.sh")" = "history/sessions/$(date +%F).md" ] && grep -q "## Questions and answers" "$LOG" || die "log.sh didn't create today's log"
printf '# Session log\n\n### Q: what is the release date?\n**A:** the 14th, decided with the owner.\n' > "$LOG"
"$W/kit/log.sh" >/dev/null; grep -q "release date" "$LOG" || die "log.sh overwrote an existing log"
git -C "$W" add -A; git -C "$W" commit -qm "Session log"
out="$("$W/kit/check.sh" 2>&1)" || die "check failed on a clean local project" "$out"
ok "end-of-session check: requires today's session log and a current handbook"

out="$("$W/kit/find.sh" "release date" 2>&1)" || die "find didn't find a recorded answer" "$out"
echo "$out" | grep -q "history/sessions/" || die "find didn't point at the session log" "$out"
"$W/kit/find.sh" "dock badge" >"$T/out" 2>&1 && die "find claimed a hit for something never discussed"
grep -q "nothing recorded" "$T/out" || die "find didn't say nothing was recorded" "$(cat "$T/out")"
ok "find.sh answers from the session log, and says plainly when nothing was recorded"

echo "2. Local → private (bare repos stand in for GitHub)"
mkdir -p "$T/remote"
git init -q --bare "$T/remote/myapp.git"; git init -q --bare "$T/remote/myapp-handbook.git"
git -C "$C" remote add origin "$T/remote/myapp.git"
git -C "$W" remote add origin "$T/remote/myapp-handbook.git"
conf "$W" CODE_REMOTE "$T/remote/myapp.git"; conf "$W" STAGE "private"
git -C "$W" commit -qam "Stage: private"
git -C "$W" push -q -u origin HEAD
git -C "$C" push -q -u origin HEAD 2>/dev/null || die "first push blocked unexpectedly"
ok "first push of both repos"

tick; echo "change" >> "$C/app.txt"; git -C "$C" commit -qam "Change without docs"
if git -C "$C" push -q 2>"$T/err"; then die "push allowed without a handbook update"; fi
grep -q "handbook was last updated" "$T/err" || die "wrong block message" "$(cat "$T/err")"
ok "push blocked when the handbook wasn't updated"
mv "$LOG" "$T/log.bak"; git -C "$W" commit -qam "Remove log for test" >/dev/null
echo "- changed app" >> "$W/STATUS.md"
if git -C "$C" push -q 2>"$T/err"; then die "push allowed without today's session log"; fi
grep -q "no session log for today" "$T/err" || die "wrong block message" "$(cat "$T/err")"
mv "$T/log.bak" "$LOG"; git -C "$W" checkout -q -- STATUS.md 2>/dev/null; git -C "$W" add -A; git -C "$W" commit -qm "Restore log" >/dev/null
ok "push blocked without today's session log"

echo "- changed app" >> "$W/STATUS.md"     # edited but NOT committed: the hook saves it
git -C "$C" push -q 2>"$T/err" || die "push blocked after the handbook update" "$(cat "$T/err")"
[ -z "$(git -C "$W" status --porcelain)" ] || die "handbook edit not auto-committed"
[ "$(git -C "$W" rev-parse HEAD)" = "$(git -C "$T/remote/myapp-handbook.git" rev-parse HEAD)" ] \
    || die "handbook not pushed along with the code"
ok "handbook edit auto-saved and pushed with the code"

tick; rm "$C/app.txt"; git -C "$C" commit -qam "Break the tests"
echo "- broke it" >> "$W/STATUS.md"
if git -C "$C" push -q 2>"$T/err"; then die "push allowed with failing tests"; fi
grep -q "tests failed" "$T/err" || die "wrong block message" "$(cat "$T/err")"
ok "push blocked when the tests fail"

AGENTBOOK_BYPASS=1 git -C "$C" push -q 2>/dev/null || die "bypass didn't work"
git -C "$C" revert --no-edit HEAD >/dev/null; tick; echo "- fixed" >> "$W/STATUS.md"
git -C "$C" push -q 2>/dev/null || die "push after fix failed"
ok "owner-approved bypass works"

git -C "$C" tag v1; git -C "$C" push -q origin v1 2>/dev/null || die "tag push blocked"
ok "pushing only a tag isn't checked"

tick; echo "z" >> "$C/app.txt"; git -C "$C" commit -qam "Tweak with credit

Co-Authored-By: Some Agent <agent@example.com>"; echo "- z" >> "$W/STATUS.md"
git -C "$C" push -q 2>"$T/err" || die "push blocked a co-author line with HIDE_AI_TRACES off" "$(cat "$T/err")"
mkdir -p "$C/.claude"; echo '{}' > "$C/.claude/settings.json"
git -C "$C" status --porcelain | grep -q claude || die ".claude/ hidden although HIDE_AI_TRACES is off"
ok "by default, credit lines and AI tool folders are the owner's business: nothing blocked or hidden"

conf "$W" HIDE_AI_TRACES yes; git -C "$W" commit -qam "Turn on HIDE_AI_TRACES"
"$W/kit/setup.sh" --no-install >/dev/null
git -C "$C" status --porcelain | grep -q claude && die ".claude/ not hidden with HIDE_AI_TRACES on"
ok "HIDE_AI_TRACES=yes: AI tool folders like .claude/ are hidden from the code repo"

tick; echo "x" >> "$C/app.txt"; git -C "$C" commit -qam "Tweak

Co-Authored-By: Claude <noreply@anthropic.com>"; echo "- tweak" >> "$W/STATUS.md"
if git -C "$C" push -q 2>"$T/err"; then die "push allowed with an AI co-author line"; fi
grep -q "AI tool traces" "$T/err" || die "wrong block message" "$(cat "$T/err")"
git -C "$C" commit -q --amend -m "Tweak"; tick; echo "- tweak reworded" >> "$W/STATUS.md"
git -C "$C" push -q 2>"$T/err" || die "push blocked after rewording" "$(cat "$T/err")"
ok "push blocked for an AI credit line; passes once reworded"

tick; git -C "$C" add -f .claude/settings.json; git -C "$C" commit -qm "Settings"; echo "- s" >> "$W/STATUS.md"
if git -C "$C" push -q 2>"$T/err"; then die "push allowed with .claude/ committed"; fi
grep -q "adds AI tool files" "$T/err" || die "wrong block message" "$(cat "$T/err")"
git -C "$C" reset -q --hard HEAD~1
ok "push blocked when an AI tool folder is committed (even force-added)"

tick; echo "y" >> "$C/app.txt"; git -C "$C" -c user.email=someone@else.com commit -qam "Other identity"; echo "- y" >> "$W/STATUS.md"
if git -C "$C" push -q 2>"$T/err"; then die "push allowed with another identity"; fi
grep -q "identity: someone@else.com" "$T/err" || die "wrong block message" "$(cat "$T/err")"
git -C "$C" reset -q --hard HEAD~1
ok "push blocked when a commit isn't made as the owner"
git -C "$W" checkout -q -- STATUS.md   # drop the test's unsaved status line

echo "3. New computer: clone only the workspace, the agent runs sync"
N="$T/newmac"; mkdir -p "$N"
git clone -q "$T/remote/myapp-handbook.git" "$N/MyApp"
[ -f "$N/MyApp/AGENTS.md" ] || die "AGENTS.md not at the top of the clone (agents wouldn't find it)"
"$N/MyApp/kit/sync.sh" >"$T/out" 2>&1 || die "sync failed on a fresh clone" "$(cat "$T/out")"
NC="$N/MyApp/code"
[ "$(git -C "$NC" rev-parse HEAD)" = "$(git -C "$T/remote/myapp.git" rev-parse main)" ] || die "code not at the latest commit"
[ "$(git -C "$NC" rev-parse --abbrev-ref '@{u}')" = "origin/main" ] || die "code doesn't track origin"
git -C "$NC" rev-parse -q --verify v1 >/dev/null || die "tags not fetched"
[ -z "$(git -C "$N/MyApp" status --porcelain)$(git -C "$NC" status --porcelain)" ] || die "fresh clone shows changes"
[ "$(git -C "$NC" config core.hooksPath)" = "../kit/hooks" ] || die "hooks not installed by sync"
[ "$(git -C "$NC" config user.email)" = "owner@users.noreply.github.com" ] || die "identity not restored"
tick; echo "- new computer ready" >> "$N/MyApp/STATUS.md"; git -C "$N/MyApp" commit -qam "Status"; git -C "$N/MyApp" push -q
out="$("$N/MyApp/kit/check.sh" 2>&1)" || die "check failed on the new computer" "$out"
ok "clone the workspace, run sync: code fetched, hooks, links, identity, check all good"

echo "4. Work on one computer, continue on another"
echo "- note left unsaved by yesterday's session" >> "$W/DECISIONS.md"
"$W/kit/sync.sh" >"$T/out" 2>&1 || die "sync on the first computer failed" "$(cat "$T/out")"
git -C "$W" log -1 --format=%s | grep -q "Save handbook notes" >/dev/null || git -C "$W" log --format=%s | grep -q "Save handbook notes" || die "leftover notes not saved by sync"
grep -q "left unsaved by yesterday" "$W/DECISIONS.md" || die "leftover notes lost"
tick; echo "feature" >> "$C/app.txt"; git -C "$C" commit -qam "New feature"
echo "- shipped feature" >> "$W/STATUS.md"
git -C "$C" push -q 2>/dev/null || die "push from the first computer failed"
"$N/MyApp/kit/sync.sh" >"$T/out" 2>&1 || die "sync failed" "$(cat "$T/out")"
grep -q "feature" "$NC/app.txt" || die "sync didn't bring the code"
grep -q "shipped feature" "$N/MyApp/STATUS.md" || die "sync didn't bring the handbook"
grep -q "New feature" "$N/MyApp/history/commits.md" || die "sync didn't bring the commit log"
echo "local edit" >> "$NC/app.txt"; git -C "$NC" commit -qam "Local work"
tick; echo "other" >> "$C/app.txt"; git -C "$C" commit -qam "Other change"; echo "- other" >> "$W/STATUS.md"
git -C "$C" push -q 2>/dev/null || die "second push failed"
"$N/MyApp/kit/sync.sh" >"$T/out" 2>&1 && die "sync claimed success on diverged history"
grep -q "Nothing was lost" "$T/out" || die "sync didn't explain the divergence" "$(cat "$T/out")"
git -C "$NC" log -1 --format=%s | grep -q "Local work" || die "sync lost local work"
ok "one sync brings both; diverged work is never thrown away"

echo "5. A contributor's clone of the code (no workspace)"
K="$T/contributor"; git clone -q "$T/remote/myapp.git" "$K"
[ ! -e "$K/AGENTS.md" ] && [ ! -e "$K/agentbook.conf" ] || die "private files leaked into the code repo"
echo "pr" >> "$K/app.txt"; git -C "$K" commit -qam "Contribution"; git -C "$K" push -q 2>/dev/null || die "contributor push failed"
ok "contributors see nothing private and aren't affected"

echo "6. Going public: audit"
"$N/MyApp/kit/sync.sh" >/dev/null 2>&1; git -C "$NC" reset -q --hard origin/main
out="$("$N/MyApp/kit/public-audit.sh" 2>&1)"; code=$?
[ "$code" = 1 ] || die "audit passed a repo with no LICENSE/README" "$out"
echo "$out" | grep -q "Tester <t@example.invalid>" || die "audit didn't list identities" "$out"
printf 'MIT\n' > "$NC/LICENSE"; printf '# My App\n' > "$NC/README.md"
echo 'path = "/Users/someone/secret"' > "$NC/config.txt"
conf "$N/MyApp" HIDE_AI_TRACES yes
git -C "$NC" add -A; git -C "$NC" commit -qm "Add files

Co-Authored-By: Some Assistant <bot@example.com>"
out="$("$N/MyApp/kit/public-audit.sh" 2>&1)"
echo "$out" | grep -q "Co-Authored-By" || die "audit missed a co-author line" "$out"
echo "$out" | grep -q "/Users/someone" || die "audit missed a home path" "$out"
ok "audit flags identities, co-author lines, home paths and missing basics"

echo "7. Adopting an existing project in place"
A="$T/OldApp"; mkdir -p "$A"; git -C "$A" init -q; echo "v1" > "$A/main.txt"
git -C "$A" add -A; git -C "$A" commit -qm "Old history"; echo "scratch" > "$A/untracked.txt"
"$KIT/install.sh" "$A" "Old App" >/dev/null || die "adopt failed"
[ -f "$A/AGENTS.md" ] && [ -d "$A/code/.git" ] || die "adopted layout wrong"
git -C "$A/code" log -1 --format=%s | grep -q "Old history" || die "history lost"
[ -f "$A/code/untracked.txt" ] || die "untracked file lost"
ok "existing repo moved into code/ with history and untracked files; same folder path"

S="$T/tracked"; mkdir -p "$S"; git -C "$S" init -q; echo x > "$S/AGENTS.md"; git -C "$S" add -A; git -C "$S" commit -qm x
"$KIT/install.sh" "$S" >/dev/null 2>"$T/err" && die "install accepted a code repo with a tracked AGENTS.md"
grep -q "is tracked" "$T/err" || die "wrong error" "$(cat "$T/err")"
"$KIT/install.sh" "$W" >/dev/null 2>&1 && die "install overwrote an existing workspace"
ok "refuses tracked agent files and won't overwrite a workspace"

echo "8. One-repo layout: private project, later public"
O1="$T/Solo"
"$KIT/install.sh" "$O1" "Solo" >/dev/null || die "single install failed"
grep -q '^LAYOUT="single"' "$O1/agentbook.conf" || die "new project isn't single layout"
[ -d "$O1/code" ] && [ ! -e "$O1/code/.git" ] || die "code/ should be a plain folder"
[ "$(git -C "$O1" config core.hooksPath)" = "kit/hooks" ] || die "workspace hooks not set"
grep -qx "/code/" "$O1/.gitignore" 2>/dev/null && die "single layout ignores code/"
conf "$O1" TEST_CMD "test -f main.txt"; conf "$O1" GIT_NAME "Owner"; conf "$O1" GIT_EMAIL "owner@users.noreply.github.com"; conf "$O1" HIDE_AI_TRACES yes
"$O1/kit/setup.sh" --no-install >/dev/null
mkdir -p "$O1/history/sessions"; echo "# Session log" > "$O1/history/sessions/$(date +%F).md"
git -C "$O1" add -A; git -C "$O1" commit -qm "Configure" || die "handbook commit failed"
ok "new project is one repo: handbook and code/ together, hooks on the workspace"

echo "v1" > "$O1/code/main.txt"; echo "- started" >> "$O1/STATUS.md"; git -C "$O1" add -A
if git -C "$O1" commit -qm "Mixed" 2>"$T/err"; then die "mixed code+handbook commit allowed"; fi
grep -q "mixes code" "$O1/../err" 2>/dev/null || grep -q "mixes code" "$T/err" || die "wrong message" "$(cat "$T/err")"
git -C "$O1" commit -qm "Add main" -- code || die "code-only commit failed"
git -C "$O1" commit -qam "Status" || die "handbook-only commit failed"
grep -q "Add main" "$O1/history/commits.md" || die "code commit not logged"
grep -q "Status" "$O1/history/commits.md" && die "handbook commit was logged as code"
[ -z "$(git -C "$O1" status --porcelain)" ] || die "log not committed"
mkdir -p "$O1/code/.claude"; echo x > "$O1/code/.claude/s.json"
git -C "$O1" status --porcelain | grep -q claude && die "code/.claude not hidden"
git -C "$O1" status --porcelain | grep -q 'code/AGENTS' && die "code/AGENTS.md link not hidden"
ok "code and handbook commits stay separate; code commits logged; AI folders and links hidden"

git init -q --bare "$T/remote/solo.git"; git -C "$O1" remote add origin "$T/remote/solo.git"
conf "$O1" STAGE private; tick; git -C "$O1" commit -qam "Stage: private"
git -C "$O1" push -q -u origin main 2>"$T/err" || die "first push of the single repo failed" "$(cat "$T/err")"
tick; echo "v2" >> "$O1/code/main.txt"; git -C "$O1" commit -qm "Change code" -- code
if git -C "$O1" push -q 2>"$T/err"; then die "push allowed with a stale handbook"; fi
grep -q "handbook was last updated" "$T/err" || die "wrong block" "$(cat "$T/err")"
echo "- v2" >> "$O1/STATUS.md"
if git -C "$O1" push -q 2>"$T/err"; then die "push allowed before saved notes were included"; fi
grep -q "Push again" "$T/err" || die "wrong block" "$(cat "$T/err")"
git -C "$O1" push -q 2>"$T/err" || die "push failed after notes saved" "$(cat "$T/err")"
tick; echo "v3" >> "$O1/code/main.txt"; git -C "$O1" commit -qm "Tweak

Co-Authored-By: Claude <noreply@anthropic.com>" -- code; tick; echo "- v3" >> "$O1/STATUS.md"; git -C "$O1" commit -qam "Status v3"
if git -C "$O1" push -q 2>"$T/err"; then die "AI co-author line allowed in single layout"; fi
git -C "$O1" reset -q --hard origin/main      # throw the credited commit away and redo it cleanly
tick; echo "v3" >> "$O1/code/main.txt"; git -C "$O1" commit -qm "Tweak" -- code; tick; echo "- v3" >> "$O1/STATUS.md"; git -C "$O1" commit -qam "Status v3"
git -C "$O1" push -q 2>"$T/err" || die "clean push failed" "$(cat "$T/err")"
tick; git -C "$O1" commit -q --allow-empty -m "Notes about Claude research"; git -C "$O1" push -q 2>"$T/err" \
    || die "handbook-only push was checked like code" "$(cat "$T/err")"
ok "single-repo pushes: stale handbook, unsaved notes and AI credit blocked; handbook-only pushes free"

M="$T/solo-mac2"; git clone -q "$T/remote/solo.git" "$M"
"$M/kit/sync.sh" >"$T/out" 2>&1 || die "sync on a fresh single clone failed" "$(cat "$T/out")"
[ -f "$M/code/main.txt" ] && [ "$(git -C "$M" config core.hooksPath)" = "kit/hooks" ] || die "fresh single clone not set up"
out="$("$M/kit/check.sh" 2>&1)" || die "check failed on a fresh single clone" "$out"
ok "new computer: clone the one repo, sync, everything works"

echo "local build output" > "$O1/code/build.log"; echo "build.log" > "$O1/code/.gitignore"
git -C "$O1" add code/.gitignore; git -C "$O1" commit -qm "Ignore build log" -- code; tick; echo "- ignore" >> "$O1/STATUS.md"; git -C "$O1" commit -qam "Status"
git -C "$O1" push -q 2>/dev/null || die "push before going public failed"
out="$("$O1/kit/go-public.sh" --check 2>&1)"
echo "$out" | grep -q "single layout\|commits," || die "go-public --check didn't audit the split" "$out"
git init -q --bare "$T/remote/solo-public.git"
AGENTBOOK_APPROVED=1 "$O1/kit/go-public.sh" "$T/remote/solo-public.git" >"$T/out" 2>&1 || die "go-public failed" "$(cat "$T/out")"
PUB="$T/pubcheck"; git clone -q "$T/remote/solo-public.git" "$PUB"
[ -f "$PUB/main.txt" ] && [ ! -e "$PUB/STATUS.md" ] && [ ! -e "$PUB/agentbook.conf" ] && [ ! -e "$PUB/code" ] || die "public repo has the wrong files" "$(ls -a "$PUB")"
git -C "$PUB" log --format=%s | grep -qiE "status|stage|configure|handbook|claude" && die "handbook commits leaked into the public history" "$(git -C "$PUB" log --format=%s)"
git -C "$PUB" log --all --format=%B | grep -qi "co-authored" && die "AI credit leaked into the public history"
[ "$(git -C "$PUB" log --format=%s | wc -l | tr -d ' ')" = "4" ] || die "unexpected public commits" "$(git -C "$PUB" log --format=%s)"
ok "go-public publishes only the code's own commits, rooted at code/; no handbook files or messages"

grep -q '^LAYOUT="split"' "$O1/agentbook.conf" && grep -q '^STAGE="public"' "$O1/agentbook.conf" || die "settings not switched"
[ -d "$O1/code/.git" ] && [ "$(git -C "$O1/code" rev-parse --abbrev-ref '@{u}')" = "origin/main" ] || die "code/ isn't a clone of the public repo"
[ -f "$O1/code/build.log" ] || die "local-only file lost"
[ -z "$(git -C "$O1/code" status --porcelain)" ] || die "code/ shows changes after the switch" "$(git -C "$O1/code" status --porcelain)"
git -C "$O1" ls-files | grep -q '^code/' && die "workspace still tracks code/"
git -C "$O1" push -q 2>/dev/null || die "pushing the workspace after the switch failed"
ok "workspace switched to split: code/ now tracks the public repo, local files kept"

M2="$T/solo-mac3"; git clone -q "$T/remote/solo.git" "$M2"
"$M2/kit/sync.sh" >"$T/out" 2>&1 || die "sync after going public failed" "$(cat "$T/out")"
[ "$(git -C "$M2/code" remote get-url origin)" = "$T/remote/solo-public.git" ] || die "code not fetched from the public repo"
tick; echo "v4" >> "$M2/code/main.txt"; git -C "$M2/code" commit -qam "Public change"; echo "- public change" >> "$M2/STATUS.md"
git -C "$M2/code" push -q 2>"$T/err" || die "push after going public failed" "$(cat "$T/err")"
[ "$(git -C "$M2/code" rev-parse HEAD)" = "$(git -C "$T/remote/solo-public.git" rev-parse main)" ] || die "public repo not updated"
[ "$(git -C "$M2" rev-parse HEAD)" = "$(git -C "$T/remote/solo.git" rev-parse main)" ] || die "workspace not pushed with the code"
grep -q "Public change" "$M2/history/commits.md" || die "public commit not logged in the workspace"
ok "after going public: clone the workspace, code comes from the public repo, one push updates both"

echo "9. Installer options and portability"
P="$T/Opts"
"$KIT/install.sh" "$P" "Opts & Co" --git-name "Ada L" --git-email "ada@example.invalid" \
    --test-cmd 'test -f "a b.txt" && echo "x|y&z"' --agent-files "AGENTS.md CLAUDE.md" --hide-ai-traces >/dev/null \
    || die "install with options failed"
( . "$P/agentbook.conf"
  [ "$PROJECT_NAME" = "Opts & Co" ] || die "project name with & mangled: $PROJECT_NAME"
  [ "$GIT_NAME" = "Ada L" ] && [ "$GIT_EMAIL" = "ada@example.invalid" ] || die "identity options not saved"
  [ "$TEST_CMD" = 'test -f "a b.txt" && echo "x|y&z"' ] || die "test command with quotes, | and & mangled: $TEST_CMD"
  [ "$HIDE_AI_TRACES" = yes ] && [ "$AGENT_FILES" = "AGENTS.md CLAUDE.md" ] || die "flags not saved" ) || exit 1
[ ! -e "$P/GEMINI.md" ] || die "--agent-files ignored"
[ "$(git -C "$P" config user.email)" = "ada@example.invalid" ] || die "identity not applied at install"
touch "$P/code/a b.txt"; "$P/kit/check.sh" >"$T/out" 2>&1; grep -q "tests pass" "$T/out" || die "quoted test command didn't run" "$(cat "$T/out")"
ok "installer options are saved exactly (quotes, |, & and spaces survive) and applied"
if grep -rn "sed -i" "$KIT/kit" "$KIT/install.sh" >/dev/null; then die "kit uses sed -i (not portable between macOS and Linux)"; fi
for f in "$KIT"/kit/*.sh "$KIT"/kit/hooks/* "$KIT/install.sh"; do bash -n "$f" || die "syntax error in $f"; done
ok "scripts are syntax-clean and avoid platform-specific sed -i"
"$KIT/install.sh" "$T/bad" --bogus >/dev/null 2>&1 && die "installer accepted an unknown option"
ok "installer rejects unknown options"

echo "10. Upgrade"
echo "# new" >> "$W/kit/check.sh"
"$KIT/install.sh" --upgrade "$W" >/dev/null || die "upgrade failed"
grep -q "# new" "$W/kit/check.sh" && die "upgrade didn't replace scripts"
ok "upgrade replaces kit scripts and keeps docs"

echo
echo "All $pass checks passed."
