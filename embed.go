// Package agentbook embeds the kit (scripts, hooks, templates and installer) so the agentbook
// command can create and upgrade workspaces without a copy of this repository.
package agentbook

import "embed"

// Files holds install.sh, VERSION, PROMPT.md, kit/ and templates/ exactly as they are in the repo.
//
//go:embed install.sh VERSION PROMPT.md all:kit all:templates
var Files embed.FS
