#!/bin/bash
# Agentbook installer. Turns a folder into an agentbook workspace:
#
#   <folder>/          the workspace repo (keep it private): what you clone and open in your agent
#     AGENTS.md, README.md, STATUS.md, DECISIONS.md, HISTORY.md, ARCHITECTURE.md, DESIGN.md,
#     agentbook.conf, kit/, history/
#     code/            the code
#
#   ./install.sh <folder> ["<Project Name>"] [options]
#       - empty or new folder:   a new project, ONE repo (LAYOUT=single): code/ is a normal folder in
#                                it. --split makes code/ its own repo instead.
#       - files, not a git repo: same as new; the files move into <folder>/code/
#       - an existing code repo: adopted IN PLACE as LAYOUT=split: its contents move into
#                                <folder>/code/ (history, branches and untracked files move with it)
#   ./install.sh --upgrade <folder>        replace kit/ scripts in an existing workspace
#
# Options (all optional; everything can be changed later in agentbook.conf):
#   --split                    code/ is its own repo
#   --git-name <name>          commit identity for both repos
#   --git-email <email>
#   --test-cmd <command>       one command that runs every test, from inside code/
#   --agent-files "<files>"    agent entry files linked to AGENTS.md (default: AGENTS.md CLAUDE.md GEMINI.md)
#   --hide-ai-traces           turn on the optional "no AI traces in pushed code" guard
set -euo pipefail

KIT="$(cd "$(dirname "$0")" && pwd -P)"
VERSION="$(cat "$KIT/VERSION")"
# shellcheck source=/dev/null
. "$KIT/kit/lib.sh"

if [ "${1:-}" = "--upgrade" ]; then
    WS="$(cd "${2:?usage: install.sh --upgrade <folder>}" && pwd -P)"
    [ -f "$WS/agentbook.conf" ] && [ -d "$WS/.git" ] || { echo "error: $WS isn't an agentbook workspace" >&2; exit 1; }
    old="$(cat "$WS/kit/VERSION" 2>/dev/null || echo unknown)"
    rm -rf "$WS/kit"
    cp -R "$KIT/kit" "$WS/kit"
    echo "$VERSION" > "$WS/kit/VERSION"
    "$WS/kit/setup.sh" --no-install >/dev/null
    git -C "$WS" add -A kit .gitignore agentbook.conf
    git -C "$WS" commit -q -m "Upgrade agentbook kit $old → $VERSION" -- kit .gitignore agentbook.conf || true
    echo "Upgraded kit $old → $VERSION. Compare AGENTS.md and README.md with $KIT/templates/ for new rules."
    exit 0
fi

WS=""; NAME_ARG=""; WANT_SPLIT=0; HIDE=no
SET_NAME=""; SET_EMAIL=""; SET_TEST=""; SET_AGENTS=""
while [ $# -gt 0 ]; do
    case "$1" in
        --split) WANT_SPLIT=1 ;;
        --hide-ai-traces) HIDE=yes ;;
        --git-name) SET_NAME="${2:?--git-name needs a value}"; shift ;;
        --git-email) SET_EMAIL="${2:?--git-email needs a value}"; shift ;;
        --test-cmd) SET_TEST="${2:?--test-cmd needs a value}"; shift ;;
        --agent-files) SET_AGENTS="${2:?--agent-files needs a value}"; shift ;;
        -h|--help) sed -n '2,23p' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
        -*) echo "error: unknown option $1" >&2; exit 2 ;;
        *) if [ -z "$WS" ]; then WS="$1"; elif [ -z "$NAME_ARG" ]; then NAME_ARG="$1"; else echo "error: too many arguments" >&2; exit 2; fi ;;
    esac
    shift
done
[ -n "$WS" ] || { echo "usage: install.sh <folder> [\"<Project Name>\"] [--split] [options]" >&2; exit 2; }
mkdir -p "$WS"
WS="$(cd "$WS" && pwd -P)"
NAME="${NAME_ARG:-$(basename "$WS")}"
TODAY="$(date +%F)"
CODE="$WS/code"

if [ -f "$WS/agentbook.conf" ]; then
    echo "error: $WS is already an agentbook workspace (use --upgrade)" >&2
    exit 1
fi

# Moves everything in $WS (including .git and hidden files) into $WS/code, keeping $WS's path.
move_into_code() {
    local tmp="$WS.moving-$$"
    mv "$WS" "$tmp"
    mkdir "$WS"
    mv "$tmp" "$CODE"
}

NEW_LAYOUT=single
if [ -d "$WS/.git" ]; then
    # Refuse before moving anything if the code tracks an agent file (it would clash with the link).
    for f in ${SET_AGENTS:-AGENTS.md CLAUDE.md GEMINI.md}; do
        if [ -n "$(git -C "$WS" ls-files -- "$f")" ]; then
            echo "error: '$f' is tracked in this repo. Move its contents somewhere safe, 'git rm --cached $f', commit, then run again." >&2
            exit 1
        fi
    done
    echo "Existing code repo found: moving it into $WS/code/ (history and untracked files move with it)."
    move_into_code
    NEW_LAYOUT=split
elif [ -n "$(ls -A "$WS")" ]; then
    echo "Existing files (not a git repo) found: moving them into $WS/code/."
    move_into_code
    [ "$WANT_SPLIT" = 1 ] && git -C "$CODE" init -q -b main
fi
[ "$WANT_SPLIT" = 1 ] && NEW_LAYOUT=split

mkdir -p "$WS/history/sessions"
cp -R "$KIT/kit" "$WS/kit"
echo "$VERSION" > "$WS/kit/VERSION"
esc() { printf '%s' "$1" | sed -e 's/[\/&|]/\\&/g'; }
for f in "$KIT"/templates/*; do
    sed -e "s|{{PROJECT}}|$(esc "$NAME")|g" -e "s|{{DATE}}|$TODAY|g" "$f" > "$WS/$(basename "$f")"
done
HB="$WS"
hb_set_conf LAYOUT "$NEW_LAYOUT"
hb_set_conf HIDE_AI_TRACES "$HIDE"
[ -z "$SET_NAME" ] || hb_set_conf GIT_NAME "$SET_NAME"
[ -z "$SET_EMAIL" ] || hb_set_conf GIT_EMAIL "$SET_EMAIL"
[ -z "$SET_TEST" ] || hb_set_conf TEST_CMD "$SET_TEST"
[ -z "$SET_AGENTS" ] || hb_set_conf AGENT_FILES "$SET_AGENTS"
printf '# Code commit log\n\nAutomatic: one entry per code commit, newest last. The why lives in DECISIONS.md.\n\n' > "$WS/history/commits.md"
git -C "$WS" init -q -b main
"$WS/kit/setup.sh" --no-install
# Handbook first, then any existing code as its own commit (code and handbook never share a commit).
if [ "$NEW_LAYOUT" = single ]; then git -C "$WS" add -A -- . ':(exclude)code'; else git -C "$WS" add -A; fi
git -C "$WS" commit -q -m "Start the $NAME agentbook (kit $VERSION)"
if [ "$NEW_LAYOUT" = single ] && [ -n "$(git -C "$WS" status --porcelain -- code)" ]; then
    git -C "$WS" add -A code
    git -C "$WS" commit -q -m "Initial import"
fi

echo
if [ "$NEW_LAYOUT" = single ]; then
    echo "Workspace ready at $WS: one repo, code in code/."
else
    echo "Workspace ready at $WS: code in code/, its own repo."
fi
echo "Next: open the folder in your coding agent and say: \"Read AGENTS.md and finish setting up this agentbook.\""
