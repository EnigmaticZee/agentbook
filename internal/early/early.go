// Package early runs before the terminal libraries start. Bubble Tea asks the terminal for its
// background colour when it loads and waits up to 5 seconds for an answer, which some terminals
// never send. This package asks first, waits at most a quarter of a second, and hands the answer
// to Lip Gloss, so start-up is instant everywhere.
//
// AGENTBOOK_THEME=light or dark skips the question entirely.
package early

import (
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Theme is "light", "dark" or "" (detected).
var Theme = os.Getenv("AGENTBOOK_THEME")

var rgbRe = regexp.MustCompile(`rgb:([0-9a-fA-F]+)/([0-9a-fA-F]+)/([0-9a-fA-F]+)`)

func init() {
	switch Theme {
	case "light":
		lipgloss.SetHasDarkBackground(false)
		return
	case "dark":
		lipgloss.SetHasDarkBackground(true)
		return
	}
	dark := true
	if v, ok := fromColorFGBG(); ok {
		dark = v
	}
	in, out := int(os.Stdin.Fd()), int(os.Stdout.Fd())
	if runtime.GOOS != "windows" && term.IsTerminal(in) && term.IsTerminal(out) {
		if v, ok := queryBackground(in); ok {
			dark = v
		}
	}
	lipgloss.SetHasDarkBackground(dark)
}

// fromColorFGBG reads COLORFGBG ("15;0": light text on black), set by some terminals.
func fromColorFGBG() (bool, bool) {
	v := os.Getenv("COLORFGBG")
	parts := strings.Split(v, ";")
	if len(parts) < 2 {
		return false, false
	}
	bg, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return false, false
	}
	return bg < 7 || bg == 8, true
}

// queryBackground sends OSC 11 (background colour) followed by DA1, which every terminal answers,
// so we know when to stop reading even if OSC 11 is ignored.
func queryBackground(fd int) (bool, bool) {
	state, err := term.MakeRaw(fd)
	if err != nil {
		return false, false
	}
	defer term.Restore(fd, state)
	if _, err := os.Stdout.WriteString("\x1b]11;?\x1b\\\x1b[c"); err != nil {
		return false, false
	}
	resp := readUntilDA1(fd)
	m := rgbRe.FindStringSubmatch(resp)
	if m == nil {
		return false, false
	}
	ch := func(h string) float64 {
		n, _ := strconv.ParseUint(h, 16, 64)
		return float64(n) / float64(uint64(1)<<(4*len(h))-1)
	}
	r, g, b := ch(m[1]), ch(m[2]), ch(m[3])
	lum := 0.2126*r + 0.7152*g + 0.0722*b
	return lum < 0.5, true
}
