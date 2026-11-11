#!/bin/bash
# Agentbook: makes sure today's session log exists and prints its path. Agents run it at the start
# of a session, then add every question, answer and finding to that file as they go.
#   kit/log.sh
set -uo pipefail

KITDIR="$(cd "$(dirname "$0")" && pwd -P)"
# shellcheck source=/dev/null
. "$KITDIR/lib.sh"
hb_from_kit "$KITDIR"

today="$(date +%F)"
log="$HB/history/sessions/$today.md"
mkdir -p "$HB/history/sessions"
if [ ! -s "$log" ]; then
    cat > "$log" <<LOG
# Session log: $today

## Summary
- 

## Questions and answers

## Research and checks

## Tried and dropped

## Feedback from the owner
LOG
fi
echo "history/sessions/$today.md"
