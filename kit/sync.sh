#!/bin/bash
# Agentbook: brings this computer up to date. Agents run this at the start of every session.
#   kit/sync.sh
# Saves handbook notes an earlier session left unsaved, pulls the workspace, fetches or pulls the code
# (split layout), installs dependencies the first time on this computer, and re-runs setup.
# Never discards local work: if a pull can't complete, it stops and says so.
set -uo pipefail

KITDIR="$(cd "$(dirname "$0")" && pwd -P)"
# shellcheck source=/dev/null
. "$KITDIR/lib.sh"
hb_from_kit "$KITDIR"
ERR="$(mktemp)"
trap 'rm -f "$ERR"' EXIT

ok_line() { echo "  ✓ $1: up to date ($(git -C "$2" log -1 --format='%h %s' 2>/dev/null | cut -c1-60))"; }

pull_workspace() {
    if [ "$(hb_has_remote "$HB")" = no ]; then
        echo "  - workspace: no remote (local only), nothing to pull"
        return 0
    fi
    hb_autocommit "Save handbook notes left from an earlier session"
    if git -C "$HB" pull -q --rebase --autostash 2>"$ERR"; then
        ok_line workspace "$HB"
        return 0
    fi
    git -C "$HB" rebase --abort 2>/dev/null
    echo "  ✗ workspace: another computer changed the same lines:"
    sed 's/^/      /' "$ERR"
    echo "      Nothing was lost; merge it before working."
    return 1
}

pull_code() {
    if [ "$(hb_has_remote "$CODE")" = no ]; then
        echo "  - code: no remote (local only), nothing to pull"
    elif git -C "$CODE" pull -q --ff-only 2>"$ERR"; then
        ok_line code "$CODE"
    else
        echo "  ✗ code: couldn't pull automatically:"
        sed 's/^/      /' "$ERR"
        echo "      Local changes or diverged history. Nothing was lost; sort it out before working."
        return 1
    fi
}

echo "Syncing ${PROJECT_NAME:-$(basename "$HB")}"
status=0
pull_workspace || status=1
hb_from_kit "$KITDIR"          # the pull may have changed the settings (for example after go-public)

if [ "$LAYOUT" = split ] && [ -e "$CODE/.git" ]; then
    pull_code || status=1
fi
if [ -f "$HB/.git/agentbook-installed" ] && { [ "$LAYOUT" = single ] || [ -e "$CODE/.git" ]; }; then
    "$HB/kit/setup.sh" --no-install >/dev/null || status=1
else
    "$HB/kit/setup.sh" || status=1      # first time on this computer: fetch code, install dependencies
fi
[ "$status" = 0 ] && echo "Ready to work. The code is in ${CODE_DIR}/."
exit "$status"
