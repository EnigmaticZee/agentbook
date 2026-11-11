#!/bin/bash
# Agentbook: searches everything recorded about the project: docs, decisions, history, session logs
# and the code commit log. Agents use it before answering any question about the project.
#   kit/find.sh <word or phrase> [more words…]     (case-insensitive; each word searched separately)
# Example: kit/find.sh "dock badge" transmission
set -uo pipefail

KITDIR="$(cd "$(dirname "$0")" && pwd -P)"
# shellcheck source=/dev/null
. "$KITDIR/lib.sh"
hb_from_kit "$KITDIR"
[ $# -gt 0 ] || { echo "usage: kit/find.sh <word or phrase> [more…]"; exit 2; }

found=0
for term in "$@"; do
    echo "== \"$term\""
    hits="$(cd "$HB" && grep -r -n -i -F --include='*.md' --exclude-dir="$CODE_DIR" --exclude-dir=kit -- "$term" . 2>/dev/null \
        | sed 's|^\./||' | head -60)"
    if [ -n "$hits" ]; then
        echo "$hits" | cut -c1-220
        found=1
    else
        echo "  (nothing recorded)"
    fi
done
[ "$found" = 1 ] || exit 1
