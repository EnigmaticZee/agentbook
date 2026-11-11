#!/bin/bash
# Agentbook: makes this workspace ready to work in. Safe to run any number of times; agents run
# it for you (AGENTS.md tells them to, through sync.sh).
#
#   kit/setup.sh               connect everything (fetch the code first if needed), install dependencies
#   kit/setup.sh --no-install  same, without running INSTALL_CMD
set -euo pipefail

KITDIR="$(cd "$(dirname "$0")" && pwd -P)"
# shellcheck source=/dev/null
. "$KITDIR/lib.sh"
hb_from_kit "$KITDIR"
INSTALL=1
[ "${1:-}" = "--no-install" ] && INSTALL=0
AGENT_FILES="${AGENT_FILES:-AGENTS.md}"

case "$CODE_DIR" in */*|""|.|..) echo "error: CODE_DIR must be a single folder name" >&2; exit 1 ;; esac
case "$LAYOUT" in single|split) ;; *) echo "error: LAYOUT must be single or split" >&2; exit 1 ;; esac

WS_EXCLUDE="$HB/.git/info/exclude"
mkdir -p "$HB/.git/info"
add_line() { grep -qxF "$2" "$1" 2>/dev/null || echo "$2" >> "$1"; }

# 1. Agent entry files in the workspace (committed in the private workspace) point at AGENTS.md.
for f in $AGENT_FILES; do
    [ "$f" = "AGENTS.md" ] && continue
    mkdir -p "$(dirname "$HB/$f")"
    if [ -e "$HB/$f" ] && [ ! -L "$HB/$f" ]; then
        echo "error: $f exists as a real file in the workspace; merge it into AGENTS.md, delete it, re-run" >&2
        exit 1
    fi
    ln -sfn "$(hb_updirs "$f")AGENTS.md" "$HB/$f"
done

# 2. The code folder.
if [ "$LAYOUT" = split ]; then
    # Its own repo, ignored by the workspace. Fetch it if this computer doesn't have it yet.
    add_line "$HB/.gitignore" "/$CODE_DIR/"
    if [ ! -e "$CODE/.git" ]; then
        if [ -d "$CODE" ] && [ -n "$(ls -A "$CODE")" ]; then
            echo "error: $CODE exists but isn't a git repo; ask the owner what it is" >&2
            exit 1
        fi
        if [ -n "${CODE_REMOTE:-}" ]; then
            echo "Fetching the code from $CODE_REMOTE"
            git clone -q "$CODE_REMOTE" "$CODE"
        else
            mkdir -p "$CODE"
            git -C "$CODE" init -q -b main
            echo "Started a new local code repo in $CODE_DIR/"
        fi
    fi
    CODE_EXCLUDE="$(git -C "$CODE" rev-parse --git-path info/exclude)"
    case "$CODE_EXCLUDE" in /*) ;; *) CODE_EXCLUDE="$CODE/$CODE_EXCLUDE" ;; esac
    mkdir -p "$(dirname "$CODE_EXCLUDE")"
    EXCL_FILE="$CODE_EXCLUDE"; EXCL_PREFIX="/"
else
    # Part of this repo: a normal folder. It must not be ignored, and must not be a repo of its own.
    if grep -qxF "/$CODE_DIR/" "$HB/.gitignore" 2>/dev/null; then
        echo "error: .gitignore ignores /$CODE_DIR/ but LAYOUT is single; remove that line" >&2
        exit 1
    fi
    if [ -e "$CODE/.git" ]; then
        echo "error: $CODE_DIR/ is its own git repo but LAYOUT is single; ask the owner which is right" >&2
        exit 1
    fi
    mkdir -p "$CODE"
    EXCL_FILE="$WS_EXCLUDE"; EXCL_PREFIX="/$CODE_DIR/"
fi

# 3. Inside the code folder: agent entry files are local links to the workspace rules, hidden from git,
#    so they never land in the code's history. With HIDE_AI_TRACES=yes, AI tools' own folders
#    (.claude/, .cursor/…) stay local too.
for f in $AGENT_FILES; do
    add_line "$EXCL_FILE" "$EXCL_PREFIX$f"
    if [ "$LAYOUT" = split ]; then tracked="$(git -C "$CODE" ls-files -- "$f")"
    else tracked="$(git -C "$HB" ls-files -- "$CODE_DIR/$f")"; fi
    if [ -n "$tracked" ]; then
        echo "error: '$CODE_DIR/$f' is tracked. Remove it from git first (git rm --cached)." >&2
        exit 1
    fi
    mkdir -p "$(dirname "$CODE/$f")"
    if [ -e "$CODE/$f" ] && [ ! -L "$CODE/$f" ]; then
        echo "error: $CODE_DIR/$f is a real file; merge it into AGENTS.md, delete it, re-run" >&2
        exit 1
    fi
    ln -sfn "$(hb_updirs "$f")../AGENTS.md" "$CODE/$f"
done
if [ "$HIDE_AI_TRACES" = "yes" ]; then
    for p in $HB_TOOL_PATHS; do add_line "$EXCL_FILE" "$EXCL_PREFIX$p"; done
fi

# 4. Hooks.
if [ "$LAYOUT" = split ]; then HOOK_REPO="$CODE"; HOOK_PATH="../kit/hooks"
else HOOK_REPO="$HB"; HOOK_PATH="kit/hooks"; fi
current="$(git -C "$HOOK_REPO" config --get core.hooksPath || true)"
case "$current" in
    ""|"$HOOK_PATH") ;;
    *) echo "error: core.hooksPath is already '$current'. Ask the owner how to combine the hooks." >&2; exit 1 ;;
esac
git -C "$HOOK_REPO" config core.hooksPath "$HOOK_PATH"
chmod +x "$HB"/kit/*.sh "$HB"/kit/hooks/*

# 5. Commit identity (these clones only).
if [ -n "${GIT_NAME:-}" ] && [ -n "${GIT_EMAIL:-}" ]; then
    for repo in "$HB" "$CODE"; do
        [ -e "$repo/.git" ] || continue
        git -C "$repo" config user.name "$GIT_NAME"
        git -C "$repo" config user.email "$GIT_EMAIL"
    done
fi

# 6. Dependencies.
missing=""
for cmd in ${REQUIRE_CMDS:-}; do
    command -v "$cmd" >/dev/null 2>&1 || missing="$missing $cmd"
done
if [ "$INSTALL" = 1 ]; then
    if [ -n "${INSTALL_CMD:-}" ]; then
        echo "Installing dependencies: $INSTALL_CMD"
        (cd "$CODE" && eval "$INSTALL_CMD") || echo "warning: the install command failed; see above" >&2
    fi
    touch "$HB/.git/agentbook-installed"
fi

echo "Ready: $HB (code in $CODE_DIR/, layout: $LAYOUT)"
[ -z "$missing" ] || echo "Still missing (install by hand):$missing"
[ -z "${MANUAL_SETUP:-}" ] || echo "Manual steps, if not done on this computer yet: $MANUAL_SETUP"
