#!/bin/bash
# Agentbook: shared helpers. Source it; don't run it.
#
# Every project is a WORKSPACE: a private git repo holding the handbook, with the code in <CODE_DIR>/.
# Two layouts (LAYOUT in agentbook.conf):
#   single  one private repo: the handbook AND the code (code/ is a normal folder in it). New projects
#           start here. Code and handbook changes are committed separately, so the code's history can
#           be split out cleanly if the project ever gets its own (public) repo: kit/go-public.sh.
#   split   two repos: the workspace (private) and code/ as its own repo inside it (public or private),
#           ignored by the workspace and fetched by setup. Projects that already had a code repo, and
#           projects after go-public.sh, use this.

# From a kit script: sets HB, CODE, CODE_DIR, LAYOUT.
hb_from_kit() {
    HB="$(cd "$1/.." && pwd -P)"
    # shellcheck source=/dev/null
    . "$HB/agentbook.conf"
    HIDE_AI_TRACES="${HIDE_AI_TRACES:-no}"
    CODE_DIR="${CODE_DIR:-code}"
    CODE="$HB/$CODE_DIR"
    LAYOUT="${LAYOUT:-split}"
}

# From a git hook: sets HB, CODE, CODE_DIR, LAYOUT and REPO (the repo the hook runs in). Returns 1 when
# the repo isn't part of a handbook workspace (for example a contributor's clone): hooks do nothing.
hb_locate() {
    local top
    top="$(git rev-parse --show-toplevel 2>/dev/null)" || return 1
    top="$(cd "$top" && pwd -P)"
    if [ -f "$top/agentbook.conf" ]; then
        hb_from_kit "$top/kit"
        REPO="$HB"
        [ "$LAYOUT" = single ]
        return
    fi
    [ -f "$top/../agentbook.conf" ] && [ -d "$top/../.git" ] || return 1
    hb_from_kit "$top/../kit"
    REPO="$CODE"
    [ "$LAYOUT" = split ] && [ "$CODE" = "$top" ]
}

# In-place sed that works with both BSD (macOS) and GNU (Linux) sed: hb_sed_i <expr> <file>
hb_sed_i() {
    local tmp
    tmp="$(mktemp)" || return 1
    sed -e "$1" "$2" > "$tmp" && cat "$tmp" > "$2"
    rm -f "$tmp"
}

# Sets KEY="VALUE" in agentbook.conf (adds the line if missing): hb_set_conf KEY VALUE
hb_set_conf() {
    local conf="$HB/agentbook.conf" val
    val="$(printf '%s' "$2" | sed -e 's/[\\|&]/\\&/g' -e 's/"/\\\\"/g')"
    if grep -q "^$1=" "$conf"; then
        hb_sed_i "s|^$1=.*|$1=\"$val\"|" "$conf"
    else
        printf '%s="%s"\n' "$1" "$(printf '%s' "$2" | sed 's/"/\\"/g')" >> "$conf"
    fi
}

# Unix time of the last workspace commit that changed real docs (not code/, not the commit log).
hb_last_doc_update() {
    local t
    t="$(git -C "$HB" log -1 --format=%ct -- . ':(exclude)history' ":(exclude)$CODE_DIR" 2>/dev/null)"
    echo "${t:-0}"
}

# Unix time of the latest code commit.
hb_last_code_commit() {
    local t
    if [ "$LAYOUT" = single ]; then
        t="$(git -C "$HB" log -1 --format=%ct -- "$CODE_DIR" 2>/dev/null)"
    else
        t="$(git -C "$CODE" log -1 --format=%ct 2>/dev/null)"
    fi
    echo "${t:-0}"
}

# Commits any uncommitted handbook edits (never code), so nothing written is ever lost.
hb_autocommit() {
    if [ "$LAYOUT" = single ]; then
        git -C "$HB" add -A -- . ":(exclude)$CODE_DIR" 2>/dev/null
        if ! git -C "$HB" diff --cached --quiet -- . ":(exclude)$CODE_DIR"; then
            git -C "$HB" commit -q -m "${1:-Update handbook}" -- . ":(exclude)$CODE_DIR"
        fi
    elif [ -n "$(git -C "$HB" status --porcelain)" ]; then
        git -C "$HB" add -A
        git -C "$HB" commit -q -m "${1:-Update handbook}"
    fi
}

hb_has_remote() {
    if [ -n "$(git -C "$1" remote 2>/dev/null)" ]; then echo yes; else echo no; fi
}

# Commits not yet on the remote branch: a number, "none" without a remote, "nobranch" if never pushed.
hb_unpushed() {
    local repo="$1"
    [ "$(hb_has_remote "$repo")" = yes ] || { echo none; return; }
    if git -C "$repo" rev-parse --abbrev-ref --symbolic-full-name '@{u}' >/dev/null 2>&1; then
        git -C "$repo" rev-list --count '@{u}..HEAD'
    else
        echo nobranch
    fi
}

# Pushes the current branch, setting the upstream the first time. If another computer pushed in the
# meantime, replays this computer's commits on top and tries again; on a real conflict it stops
# cleanly and leaves everything as it was.
hb_push() {
    local repo="$1"
    if ! git -C "$repo" rev-parse --abbrev-ref --symbolic-full-name '@{u}' >/dev/null 2>&1; then
        git -C "$repo" push -q -u origin HEAD
        return
    fi
    git -C "$repo" push -q 2>/dev/null && return 0
    if git -C "$repo" pull -q --rebase --autostash 2>/dev/null; then
        git -C "$repo" push -q
    else
        git -C "$repo" rebase --abort 2>/dev/null
        echo "agentbook: another computer changed the same lines; merge by hand (git -C '$repo' pull)" >&2
        return 1
    fi
}

# Relative "../" prefix for a path like "a/b/file" (one "../" per folder level).
hb_updirs() {
    local depth i=0 prefix=""
    depth="$(echo "$1" | awk -F/ '{print NF-1}')"
    while [ "$i" -lt "$depth" ]; do prefix="../$prefix"; i=$((i + 1)); done
    echo "$prefix"
}

# Single layout: writes the code's own history (only commits that touched code/, with code/ as the
# root, the workspace's handbook files never included) into a new repo at <dest>, branch main.
hb_split_code() {
    local dest="$1" branch
    branch="$(git -C "$HB" rev-parse --abbrev-ref HEAD)"
    git clone -q --no-local --single-branch --branch "$branch" "$HB" "$dest" || return 1
    git -C "$dest" remote remove origin
    for tag in $(git -C "$dest" tag -l); do git -C "$dest" tag -d "$tag" >/dev/null; done
    (cd "$dest" && FILTER_BRANCH_SQUELCH_WARNING=1 git filter-branch -f --prune-empty \
        --subdirectory-filter "$CODE_DIR" HEAD >/dev/null 2>&1) || return 1
    git -C "$dest" branch -q -M main
    rm -rf "$dest/.git/refs/original"
    git -C "$dest" reflog expire --expire=now --all
    git -C "$dest" gc -q --prune=now 2>/dev/null
}

# ---- AI trace patterns (used by the optional HIDE_AI_TRACES guard and the public audit) ----
# Names of AI coding tools and assistants, as whole words.
HB_TOOLS='(^|[^a-z0-9])(claude|anthropic|openai|chatgpt|gpt-[0-9][a-z0-9.]*|copilot|cursor (ai|agent)|gemini|codex|windsurf|aider|devin|llm)([^a-z0-9]|$)'
# Credit phrases and trailers.
HB_CREDIT='co-authored-by|(generated|written|created|built) (with|by|using) (an? )?(ai|llm|assistant|agent|claude|chatgpt|copilot|cursor|gemini|codex)|(^|[^a-z])ai[- ](generated|assisted|written)|🤖'
# Files and folders AI tools keep in a project. They must never be committed to the code.
HB_TOOL_PATHS='AGENTS.md CLAUDE.md GEMINI.md .claude/ .cursor/ .cursorrules .windsurfrules .windsurf/ .aider* .codex/ .continue/ .github/copilot-instructions.md .github/instructions/'
# Same list as one regex over paths relative to the code root.
HB_TOOL_PATHS_RE='^(AGENTS\.md|CLAUDE\.md|GEMINI\.md|\.claude/|\.cursor/|\.cursorrules|\.windsurfrules|\.windsurf/|\.aider|\.codex/|\.continue/|\.github/copilot-instructions\.md|\.github/instructions/)'
