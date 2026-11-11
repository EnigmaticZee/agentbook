#!/bin/bash
# Agentbook: checks the code repo (all files AND all history) before it becomes public.
# Read-only: it reports, it never changes anything.
#   kit/public-audit.sh              split layout: the code repo; single layout: exactly the history
#                                    kit/go-public.sh would publish
#   kit/public-audit.sh --repo <dir> audit another repo (used by go-public.sh)
# Exits 1 if anything needs a decision from the owner.

KITDIR="$(cd "$(dirname "$0")" && pwd -P)"
# shellcheck source=/dev/null
. "$KITDIR/lib.sh"
hb_from_kit "$KITDIR"
if [ "${1:-}" = "--repo" ]; then
    TARGET="${2:?--repo needs a folder}"
elif [ "$LAYOUT" = single ]; then
    TMPA="$(mktemp -d)"; trap 'rm -rf "$TMPA"' EXIT
    TARGET="$TMPA/code"
    hb_split_code "$TARGET" || { echo "couldn't extract the code's history"; exit 1; }
    echo "(single layout: auditing the code's own history, as go-public.sh would publish it)"
else
    TARGET="$CODE"
fi
cd "$TARGET" 2>/dev/null && [ -e .git ] || { echo "No code repo in $TARGET (run kit/sync.sh)"; exit 1; }

found=0
section() { echo; echo "== $1"; }
hit() { echo "  ✗ $1"; found=$((found + 1)); }
clean() { echo "  ✓ $1"; }

echo "Public-readiness audit for ${PROJECT_NAME:-$(basename "$CODE")}"
echo "Branches and tags: $(git for-each-ref --format='%(refname:short)' refs/heads refs/tags | tr '\n' ' ')"

section "Private files in the repo or its history"
# Agent entry files are links to the private handbook; AI tool folders count only with HIDE_AI_TRACES=yes.
entry_re="$(for f in ${AGENT_FILES:-AGENTS.md}; do printf '%s|' "$(printf '%s' "$f" | sed 's/[.]/\\./g')"; done)"
if [ "$HIDE_AI_TRACES" = yes ]; then entry_re="$entry_re${HB_TOOL_PATHS_RE#^(}"; entry_re="^(${entry_re}"; else entry_re="^(${entry_re%|})$"; fi
private="$(git log --all --format= --name-only | sort -u | grep -E "$entry_re|^(agentbook\.conf$|STATUS\.md$|DECISIONS\.md$|\.env($|\.)|.*\.(pem|key|p12|mobileprovision)$)" || true)"
if [ -n "$private" ]; then echo "$private" | sed 's/^/  ✗ /'; found=$((found + 1)); else clean "none"; fi

# Names of coding tools and assistants, and the usual credit phrases (shared with the pre-push hook).
TOOLS="$HB_TOOLS"
CREDIT="$HB_CREDIT"

if [ "$HIDE_AI_TRACES" = yes ]; then
    section "AI tool credit in commit messages (HIDE_AI_TRACES is on)"
    credits="$(git log --all --format='%h %s%n%b' | grep -i -E "$CREDIT|$TOOLS" || true)"
    if [ -n "$credits" ]; then echo "$credits" | head -20 | sed 's/^/  ✗ /'; found=$((found + 1)); else clean "none"; fi
fi

section "Commit authors (should all be the public identity)"
git log --all --format='%an <%ae>|%cn <%ce>' | tr '|' '\n' | sort | uniq -c | sed 's/^/    /'
if [ -n "${GIT_EMAIL:-}" ] && git log --all --format='%ae%n%ce' | grep -v -x -F "$GIT_EMAIL" | grep -v -x 'noreply@github.com' | grep -q .; then
    hit "some commits use an email other than $GIT_EMAIL"
fi

if [ "$HIDE_AI_TRACES" = yes ]; then
    section "Mentions of AI tools in current files (HIDE_AI_TRACES is on)"
    mentions="$(git grep -n -i -I -E "$CREDIT|$TOOLS" -- . ':(exclude)*.lock' 2>/dev/null | head -20 || true)"
    if [ -n "$mentions" ]; then echo "$mentions" | sed 's/^/  ? /'; echo "  (review these: some may be fine, like a how-to)"; found=$((found + 1)); else clean "none"; fi
fi

section "Personal paths, emails and likely secrets (current files and history)"
patterns='/Users/[A-Za-z0-9._-]+/|/home/[A-Za-z0-9._-]+/|[A-Za-z0-9._%+-]+@(gmail|outlook|hotmail|icloud|yahoo)\.com|BEGIN [A-Z ]*PRIVATE KEY|(api[_-]?key|secret|token|password)[\"'"'"' ]*[:=][\"'"'"' ]*[A-Za-z0-9/+_-]{12,}|sk-[A-Za-z0-9]{20,}|ghp_[A-Za-z0-9]{20,}|AKIA[0-9A-Z]{16}'
now="$(git grep -n -I -E "$patterns" 2>/dev/null | head -20 || true)"
past="$(git log --all -p --format='commit %h' | grep -E "^\+.*($patterns)" | head -20 || true)"
if [ -n "$now$past" ]; then
    [ -z "$now" ] || echo "$now" | sed 's/^/  ✗ now: /'
    [ -z "$past" ] || echo "$past" | sed 's/^/  ✗ history: /'
    found=$((found + 1))
else
    clean "none"
fi

section "Big files in history (over 5 MB)"
big="$(git rev-list --objects --all | git cat-file --batch-check='%(objecttype) %(objectsize) %(rest)' | awk '$1=="blob" && $2>5242880 {printf "%.1f MB %s\n", $2/1048576, $3}')"
if [ -n "$big" ]; then echo "$big" | sed 's/^/  ? /'; else clean "none"; fi

section "Basics a public repo needs"
for f in LICENSE README.md; do
    if git ls-files --error-unmatch "$f" >/dev/null 2>&1; then clean "$f"; else hit "no $f"; fi
done

echo
if [ "$found" = 0 ]; then
    echo "Nothing found. Still read through README and the code once before publishing."
else
    echo "$found area(s) need a look. Anything in HISTORY needs a history rewrite to remove;"
    echo "that is destructive, so ask the owner first."
    exit 1
fi
