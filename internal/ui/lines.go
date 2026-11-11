package ui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Hang wraps text to width and indents continuation lines under the first line's text.
func Hang(prefix, text string, width int, style func(string) string) string {
	pw := lipgloss.Width(prefix)
	wrapped := lipgloss.NewStyle().Width(max(width-pw, 20)).Render(text)
	lines := strings.Split(wrapped, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
		if i == 0 {
			lines[i] = prefix + style(lines[i])
		} else {
			lines[i] = strings.Repeat(" ", pw) + style(lines[i])
		}
	}
	return strings.Join(lines, "\n")
}

// Kind is what a line of kit script output means.
type Kind int

const (
	KPlain Kind = iota
	KTitle
	KSection
	KOK
	KBad
	KSkip
	KWarn
	KDetail
	KVerdict
	KBlank
)

// Classify reads the kit scripts' output conventions: "  ✓ …", "  ✗ …", "  - …", "  ? …",
// "== section", deeper-indented detail lines, and the closing verdict.
func Classify(line string, first bool) (Kind, string) {
	t := strings.TrimSpace(line)
	switch {
	case t == "":
		return KBlank, ""
	case strings.HasPrefix(t, "✓ "):
		return KOK, t[len("✓ "):]
	case strings.HasPrefix(t, "✗ "):
		return KBad, t[len("✗ "):]
	case strings.HasPrefix(line, "  - "):
		return KSkip, t[2:]
	case strings.HasPrefix(t, "? "):
		return KWarn, t[2:]
	case strings.HasPrefix(t, "== "):
		return KSection, t[3:]
	case strings.HasPrefix(line, "    "):
		return KDetail, t
	case strings.HasPrefix(t, "All clear") || strings.HasPrefix(t, "Ready to work") ||
		strings.Contains(t, "thing(s) to fix") || strings.Contains(t, "area(s) need a look") ||
		strings.HasPrefix(t, "Nothing found."):
		return KVerdict, t
	case first && (strings.HasPrefix(t, "Syncing ") || strings.HasPrefix(t, "Agentbook check for") ||
		strings.HasPrefix(t, "Public-readiness audit for")):
		return KTitle, t
	}
	return KPlain, t
}

var codeSpan = regexp.MustCompile("`[^`]+`|kit/[a-z-]+\\.sh|history/[^ ]+\\.md")

// Highlight colours `code`, kit script names and log paths inside plain text.
func Highlight(s string, base func(...string) string) string {
	var b strings.Builder
	last := 0
	for _, m := range codeSpan.FindAllStringIndex(s, -1) {
		b.WriteString(base(s[last:m[0]]))
		b.WriteString(Code.Render(strings.Trim(s[m[0]:m[1]], "`")))
		last = m[1]
	}
	b.WriteString(base(s[last:]))
	return b.String()
}

// RenderLine draws one classified line for the terminal ("" means: print nothing).
func RenderLine(k Kind, t string, width int) string {
	hl := func(base func(...string) string) func(string) string {
		return func(s string) string { return Highlight(s, base) }
	}
	switch k {
	case KTitle:
		return ""
	case KSection:
		return "\n" + Indent + Warm.Render(t)
	case KOK:
		return Hang(Indent+OKStyle.Render(IconOK)+" ", t, width-2, hl(Text.Render))
	case KBad:
		return Hang(Indent+BadStyle.Render(IconBad)+" ", t, width-2, hl(Bold.Render))
	case KSkip:
		return Hang(Indent+Fainter.Render(IconSkip)+" ", t, width-2, hl(Dim.Render))
	case KWarn:
		return Hang(Indent+Warm.Render(IconWarn)+" ", t, width-2, hl(Text.Render))
	case KDetail:
		return Indent + "    " + Fainter.Render(Truncate(t, width-8))
	case KBlank, KVerdict:
		return ""
	}
	return Hang(Indent, t, width-2, hl(Dim.Render))
}
