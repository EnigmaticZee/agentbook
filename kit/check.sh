#!/bin/bash
# Agentbook: end-of-session check. Agents run it before ending a session.
#   kit/check.sh             full check
#   kit/check.sh --no-tests  skip the test run
# Exits 1 if anything needs doing, and says what.

KITDIR="$(cd "$(dirname "$0")" && pwd -P)"
# shellcheck source=/dev/null
. "$KITDIR/lib.sh"
hb_from_kit "$KITDIR"

problems=0
ok()   { echo "  ✓ $1"; }
bad()  { echo "  ✗ $1"; problems=$((problems + 1)); }
skip() { echo "  - $1"; }

echo "Agentbook check for ${PROJECT_NAME:-$(basename "$HB")} (layout: $LAYOUT)"

if [ "$LAYOUT" = split ]; then
    [ -e "$CODE/.git" ] || bad "no code in $CODE_DIR/ (run kit/sync.sh)"
    repos="workspace:$HB code:$CODE"
else
    repos="workspace:$HB"
fi

for pair in $repos; do
    name="${pair%%:*}"; repo="${pair#*:}"
    [ -e "$repo/.git" ] || continue
    if [ -z "$(git -C "$repo" status --porcelain)" ]; then
        ok "$name: everything committed"
    else
        bad "$name: uncommitted changes ($(git -C "$repo" status --porcelain | wc -l | tr -d ' ') files)"
    fi
    unpushed="$(hb_unpushed "$repo")"
    case "$unpushed" in
        none)     skip "$name: no remote yet (local only)" ;;
        nobranch) bad "$name: this branch has never been pushed" ;;
        0)        ok "$name: pushed" ;;
        *)        bad "$name: $unpushed commit(s) not pushed" ;;
    esac
done

today="$(date +%F)"
status_date="$(grep -m1 -o 'Last updated: [0-9-]*' "$HB/STATUS.md" 2>/dev/null | awk '{print $3}')"
if [ -z "$status_date" ]; then
    bad "STATUS.md has no 'Last updated: YYYY-MM-DD' line"
elif [ "$status_date" \< "$today" ]; then
    bad "STATUS.md was last updated $status_date (today is $today)"
else
    ok "STATUS.md updated today"
fi

today_log="history/sessions/$today.md"
if [ -s "$HB/$today_log" ]; then
    ok "today's session log exists ($today_log)"
else
    bad "no session log for today ($today_log): record this session's questions, answers and research"
fi

if [ "$(hb_last_doc_update)" -ge "$(hb_last_code_commit)" ]; then
    ok "handbook is newer than the latest code commit"
else
    bad "the latest code commit is newer than the last handbook update"
fi
if [ "$LAYOUT" = split ]; then hook_repo="$CODE"; hook_path="../kit/hooks"; else hook_repo="$HB"; hook_path="kit/hooks"; fi
if [ -e "$hook_repo/.git" ]; then
    [ "$(git -C "$hook_repo" config --get core.hooksPath)" = "$hook_path" ] || bad "hooks aren't installed (run kit/setup.sh)"
fi
for f in ${AGENT_FILES:-AGENTS.md}; do
    [ -L "$CODE/$f" ] || bad "$CODE_DIR/$f isn't linked (run kit/setup.sh)"
done

for f in AGENTS.md README.md STATUS.md DECISIONS.md HISTORY.md ARCHITECTURE.md; do
    [ -f "$HB/$f" ] || bad "missing $f"
done
if [ "$LAYOUT" = split ]; then
    grep -qxF "/$CODE_DIR/" "$HB/.gitignore" 2>/dev/null || bad ".gitignore doesn't ignore /$CODE_DIR/ (run kit/setup.sh)"
fi

if [ "${1:-}" = "--no-tests" ]; then
    skip "tests skipped"
elif [ -z "${TEST_CMD:-}" ]; then
    bad "no TEST_CMD in agentbook.conf"
elif [ ! -d "$CODE" ]; then
    skip "tests skipped (no code)"
else
    out="$(mktemp)"
    if (cd "$CODE" && eval "$TEST_CMD") </dev/null >"$out" 2>&1; then
        ok "tests pass"
    else
        bad "tests fail (last lines below)"
        tail -15 "$out" | sed 's/^/      /'
    fi
    rm -f "$out"
fi

echo
if [ "$problems" = 0 ]; then
    echo "All clear: safe to end the session."
else
    echo "$problems thing(s) to fix before ending the session."
    exit 1
fi
