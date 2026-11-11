#!/bin/bash
# Agentbook: gives the code its own repo (single layout → split layout). Usually to go public.
#
#   kit/go-public.sh --check                 audit exactly what would be published; changes nothing
#   kit/go-public.sh <empty repo URL>        publish the code's history there and switch to split
#   kit/go-public.sh <empty repo URL> --stage private   same, for a separate private code repo
#
# What gets published: only commits that changed code/, rewritten so code/ is the root. Handbook files
# and handbook-only commits never leave this repo. (The pre-commit hook keeps code and handbook
# changes in separate commits, so nothing mixed can come along.)
#
# Only run the publishing form after the owner has reviewed the --check output and said yes.
# The agent creates the empty GitHub repo first (with the owner's OK); see AGENTS.md "Changing stage".
set -uo pipefail

KITDIR="$(cd "$(dirname "$0")" && pwd -P)"
# shellcheck source=/dev/null
. "$KITDIR/lib.sh"
hb_from_kit "$KITDIR"

[ "$LAYOUT" = single ] || { echo "error: the code already has its own repo (LAYOUT=$LAYOUT)" >&2; exit 1; }
[ -z "$(git -C "$HB" status --porcelain)" ] || { echo "error: commit everything first (kit/check.sh)" >&2; exit 1; }

URL="${1:?usage: kit/go-public.sh --check | <empty repo URL> [--stage public|private]}"
NEWSTAGE="public"
[ "${2:-}" = "--stage" ] && NEWSTAGE="${3:?--stage needs public or private}"

TMP="$(mktemp -d)"; TMP="$(cd "$TMP" && pwd -P)"
trap 'rm -rf "$TMP"' EXIT
SPLIT="$TMP/code"

echo "Extracting the code's own history…"
hb_split_code "$SPLIT" || { echo "error: couldn't extract the code's history" >&2; exit 1; }
echo "  $(git -C "$SPLIT" rev-list --count HEAD) commits, $(git -C "$SPLIT" ls-files | wc -l | tr -d ' ') files"

echo
"$KITDIR/public-audit.sh" --repo "$SPLIT"
audit=$?

if [ "$URL" = "--check" ]; then
    echo
    echo "Nothing was changed. If the owner is happy with the above, run: kit/go-public.sh <empty repo URL>"
    exit "$audit"
fi
if [ "$audit" != 0 ] && [ "${AGENTBOOK_APPROVED:-}" != 1 ]; then
    echo
    echo "Stopped: the audit found things to review. Fix them (or, if the owner reviewed them and says"
    echo "they're fine, run again with AGENTBOOK_APPROVED=1)."
    exit 1
fi

echo
echo "Publishing the code's history to $URL"
git -C "$SPLIT" push -q "$URL" main:main || { echo "error: push failed (is the repo empty and reachable?)" >&2; exit 1; }

echo "Switching this workspace to the split layout…"
# Hooks off while the workspace stops tracking code/ (that commit is handbook-only by design).
git -C "$HB" config --unset core.hooksPath 2>/dev/null
git -C "$HB" rm -r -q --cached "$CODE_DIR"
grep -qxF "/$CODE_DIR/" "$HB/.gitignore" 2>/dev/null || echo "/$CODE_DIR/" >> "$HB/.gitignore"
# code/ becomes a clone of the new repo without touching its files (local-only files stay put).
git -C "$CODE" init -q -b main
git -C "$CODE" remote add origin "$URL"
git -C "$CODE" fetch -q origin
git -C "$CODE" reset -q --mixed origin/main
git -C "$CODE" branch -q --set-upstream-to=origin/main main
# Single-layout lines in the workspace's local exclude file no longer apply.
hb_sed_i "\|^/$CODE_DIR/|d" "$HB/.git/info/exclude" 2>/dev/null || true

hb_set_conf LAYOUT split
hb_set_conf CODE_REMOTE "$URL"
hb_set_conf STAGE "$NEWSTAGE"

"$KITDIR/setup.sh" --no-install >/dev/null || { echo "error: setup failed after the switch" >&2; exit 1; }
git -C "$HB" add -A
git -C "$HB" commit -q -m "Code moved to its own repo ($NEWSTAGE): $URL"

echo
echo "Done. The code is now its own repo in $CODE_DIR/ (origin: $URL)."
[ -n "$(git -C "$CODE" status --porcelain)" ] && echo "Note: $CODE_DIR/ has local files git doesn't track:" \
    && git -C "$CODE" status --short | head -10 | sed 's/^/  /'
echo "Next: log this in DECISIONS.md and STATUS.md, push the workspace, and if it was named"
echo "<project>, consider renaming it to <project>-handbook (GitHub keeps redirects)."
