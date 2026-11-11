// Package ui draws agentbook's terminal interface: colours that work on light and dark terminals,
// cards, script output, spinners, menus and forms.
package ui

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Palette: ink on paper with a bookmark-ribbon accent. Each colour has a light-terminal and a
// dark-terminal value with enough contrast on both.
var (
	Ink   = lipgloss.AdaptiveColor{Light: "#1C1B17", Dark: "#EEE9DC"}
	Muted = lipgloss.AdaptiveColor{Light: "#5F5B50", Dark: "#B5AE9C"}
	Faint = lipgloss.AdaptiveColor{Light: "#8E887A", Dark: "#7A7466"}
	Line  = lipgloss.AdaptiveColor{Light: "#D2CAB5", Dark: "#403D33"}
	Brand = lipgloss.AdaptiveColor{Light: "#C2410C", Dark: "#FF8A5C"}
	Gold  = lipgloss.AdaptiveColor{Light: "#8A6200", Dark: "#F2C14E"}
	Green = lipgloss.AdaptiveColor{Light: "#2F7A50", Dark: "#6CC497"}
	Red   = lipgloss.AdaptiveColor{Light: "#B42318", Dark: "#FF7A7A"}
	Chip  = lipgloss.AdaptiveColor{Light: "#EFE9DA", Dark: "#2C2A22"}
)

var (
	Bold     = lipgloss.NewStyle().Bold(true).Foreground(Ink)
	Text     = lipgloss.NewStyle().Foreground(Ink)
	Dim      = lipgloss.NewStyle().Foreground(Muted)
	Fainter  = lipgloss.NewStyle().Foreground(Faint)
	Accent   = lipgloss.NewStyle().Foreground(Brand).Bold(true)
	Warm     = lipgloss.NewStyle().Foreground(Gold).Bold(true)
	OKStyle  = lipgloss.NewStyle().Foreground(Green).Bold(true)
	BadStyle = lipgloss.NewStyle().Foreground(Red).Bold(true)
	Code     = lipgloss.NewStyle().Foreground(Brand)
	Italic   = lipgloss.NewStyle().Foreground(Muted).Italic(true)
	Indent   = "  "
)

// Icons used everywhere, so output reads the same in every command.
const (
	IconOK   = "✓"
	IconBad  = "✗"
	IconSkip = "·"
	IconWarn = "!"
	IconNext = "→"
)

// Width is the usable terminal width (capped so cards stay readable on wide screens).
func Width() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		w = 80
	}
	if w > 100 {
		w = 100
	}
	return w
}

// IsTTY reports whether stdout is an interactive terminal.
func IsTTY() bool { return term.IsTerminal(int(os.Stdout.Fd())) && term.IsTerminal(int(os.Stdin.Fd())) }

// Wordmark is the product name as drawn in headers: a ribbon mark and the name.
func Wordmark() string {
	return lipgloss.NewStyle().Foreground(Brand).Bold(true).Render("▍") + Bold.Render("agentbook")
}

// Pill draws a small rounded-looking label.
func Pill(s string, fg lipgloss.TerminalColor) string {
	return lipgloss.NewStyle().Foreground(fg).Background(Chip).Padding(0, 1).Render(s)
}

// Header is the one-line top of every screen: wordmark, version and context.
func Header(version, context string) string {
	left := Wordmark() + " " + Fainter.Render("v"+version)
	if context != "" {
		left += Fainter.Render("  ·  ") + Dim.Render(context)
	}
	return Indent + left
}

// Card draws a rounded box with an optional title in the border colour.
func Card(title, body string, border lipgloss.TerminalColor, width int) string {
	st := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(border).
		Padding(0, 2).Width(width - 4)
	box := st.Render(body)
	if title == "" {
		return indent(box)
	}
	// Put the title into the top border: ╭─ Title ───╮
	lines := strings.Split(box, "\n")
	t := " " + title + " "
	bc := lipgloss.NewStyle().Foreground(border)
	top := bc.Render("╭─") + lipgloss.NewStyle().Foreground(border).Bold(true).Render(t)
	rest := lipgloss.Width(lines[0]) - lipgloss.Width(top) - 1
	if rest < 0 {
		rest = 0
	}
	lines[0] = top + bc.Render(strings.Repeat("─", rest)+"╮")
	return indent(strings.Join(lines, "\n"))
}

func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = Indent + lines[i]
	}
	return strings.Join(lines, "\n")
}

// Rule is a faint horizontal line.
func Rule(width int) string { return Indent + Fainter.Render(strings.Repeat("─", width-4)) }

// Key renders a keyboard hint like "↑↓ move".
func Key(k, what string) string { return Accent.Render(k) + " " + Fainter.Render(what) }

// Truncate shortens s to n display cells with an ellipsis.
func Truncate(s string, n int) string {
	if n <= 1 || lipgloss.Width(s) <= n {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r))+1 > n {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}
