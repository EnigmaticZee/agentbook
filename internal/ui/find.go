package ui

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Hit is one search match: file, line number and text.
type Hit struct {
	File string
	Line string
	Text string
}

// ParseFind reads kit/find.sh output ("== term", "file:line:text", "(nothing recorded)").
func ParseFind(out []string) map[string][]Hit {
	res := map[string][]Hit{}
	term := ""
	for _, l := range out {
		if strings.HasPrefix(l, "== ") {
			term = strings.Trim(l[3:], "\"")
			res[term] = nil
			continue
		}
		parts := strings.SplitN(l, ":", 3)
		if len(parts) == 3 && term != "" {
			res[term] = append(res[term], Hit{parts[0], parts[1], strings.TrimSpace(parts[2])})
		}
	}
	return res
}

// Mark highlights every case-insensitive occurrence of term in s.
func Mark(s, term string, base func(...string) string) string {
	if term == "" {
		return base(s)
	}
	re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(term))
	if err != nil {
		return base(s)
	}
	var b strings.Builder
	last := 0
	for _, m := range re.FindAllStringIndex(s, -1) {
		b.WriteString(base(s[last:m[0]]))
		b.WriteString(Warm.Render(s[m[0]:m[1]]))
		last = m[1]
	}
	b.WriteString(base(s[last:]))
	return b.String()
}

// RenderHits draws matches grouped by file, newest session logs first.
func RenderHits(term string, hits []Hit, width, limit int) string {
	if len(hits) == 0 {
		return Indent + Fainter.Render(IconSkip) + " " + Dim.Render("Nothing recorded about ") + Bold.Render("“"+term+"”") +
			Dim.Render(". It hasn't been discussed or researched, as far as the records show.")
	}
	groups := []string{}
	byFile := map[string][]Hit{}
	for _, h := range hits {
		if _, ok := byFile[h.File]; !ok {
			groups = append(groups, h.File)
		}
		byFile[h.File] = append(byFile[h.File], h)
	}
	// Session logs first, newest first: they hold the answers to past questions.
	sorted := []string{}
	for i := len(groups) - 1; i >= 0; i-- {
		if strings.HasPrefix(groups[i], "history/sessions/") {
			sorted = append(sorted, groups[i])
		}
	}
	for _, g := range groups {
		if !strings.HasPrefix(g, "history/sessions/") {
			sorted = append(sorted, g)
		}
	}
	var b strings.Builder
	shown := 0
	for _, f := range sorted {
		if limit > 0 && shown >= limit {
			b.WriteString(Indent + Fainter.Render(fmt.Sprintf("… and more (%d matches in all)", len(hits))) + "\n")
			break
		}
		b.WriteString(Indent + Code.Render(f) + "\n")
		for _, h := range byFile[f] {
			if limit > 0 && shown >= limit {
				break
			}
			num := Fainter.Render(fmt.Sprintf("%5s ", h.Line))
			b.WriteString(Indent + num + Fainter.Render("│ ") + Mark(Truncate(h.Text, width-14), term, Text.Render) + "\n")
			shown++
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// ---- Live search ----

type searchResult struct {
	query string
	hits  []Hit
}

type findModel struct {
	header string
	input  textinput.Model
	search func(string) []Hit
	query  string
	hits   []Hit
	width  int
	height int
	seq    int
}

type debounceMsg struct{ seq int }

func (m findModel) Init() tea.Cmd { return textinput.Blink }

func (m findModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = min(msg.Width, 100), msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "enter":
			return m, tea.Quit
		}
	case debounceMsg:
		if msg.seq == m.seq {
			q := strings.TrimSpace(m.input.Value())
			return m, func() tea.Msg { return searchResult{q, m.search(q)} }
		}
		return m, nil
	case searchResult:
		if msg.query == strings.TrimSpace(m.input.Value()) {
			m.query, m.hits = msg.query, msg.hits
		}
		return m, nil
	}
	var c tea.Cmd
	before := m.input.Value()
	m.input, c = m.input.Update(msg)
	if m.input.Value() != before {
		m.seq++
		seq := m.seq
		return m, tea.Batch(c, tea.Tick(140*time.Millisecond, func(time.Time) tea.Msg { return debounceMsg{seq} }))
	}
	return m, c
}

func (m findModel) View() string {
	var b strings.Builder
	b.WriteString("\n" + m.header + "\n\n")
	b.WriteString(Card("", m.input.View(), Brand, m.width) + "\n\n")
	q := strings.TrimSpace(m.input.Value())
	switch {
	case q == "":
		b.WriteString(Indent + Dim.Render("Search every decision, status note, history entry and session log.") + "\n")
		b.WriteString(Indent + Fainter.Render("Try a feature, a library, a question you asked last week…") + "\n")
	case q != m.query:
		b.WriteString(Indent + Fainter.Render("searching…") + "\n")
	default:
		lim := m.height - 14
		if lim < 5 {
			lim = 5
		}
		b.WriteString(RenderHits(q, m.hits, m.width, lim) + "\n")
	}
	b.WriteString("\n" + Indent + Key("type", "to search") + Fainter.Render("   ") + Key("esc", "done") + "\n")
	return b.String()
}

// LiveFind runs the interactive search box. search is called (debounced) as the user types.
func LiveFind(header string, search func(string) []Hit) (string, []Hit, error) {
	ti := textinput.New()
	ti.Placeholder = "What did we decide about…"
	ti.Prompt = Accent.Render("⌕ ")
	ti.Focus()
	ti.CharLimit = 120
	m := findModel{header: header, input: ti, search: search, width: Width(), height: 30}
	final, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return "", nil, err
	}
	fm := final.(findModel)
	return fm.query, fm.hits, nil
}
