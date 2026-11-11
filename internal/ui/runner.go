package ui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Result is what a script run produced.
type Result struct {
	Err      error
	OK, Bad  int
	Verdict  string
	Title    string
	Lines    []string // raw output
	Duration time.Duration
}

type lineMsg string
type doneMsg struct{ err error }

type runModel struct {
	label   string
	spin    spinner.Model
	lines   <-chan string
	errc    <-chan error
	res     *Result
	current string
	start   time.Time
	width   int
	first   bool
	done    bool
	show    bool
}

func waitLine(lines <-chan string, errc <-chan error) tea.Cmd {
	return func() tea.Msg {
		if l, ok := <-lines; ok {
			return lineMsg(l)
		}
		return doneMsg{<-errc}
	}
}

func (m runModel) Init() tea.Cmd { return tea.Batch(m.spin.Tick, waitLine(m.lines, m.errc)) }

func (m runModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, 100)
	case spinner.TickMsg:
		var c tea.Cmd
		m.spin, c = m.spin.Update(msg)
		return m, c
	case lineMsg:
		out := m.record(string(msg))
		return m, tea.Batch(out, waitLine(m.lines, m.errc))
	case doneMsg:
		m.res.Err = msg.err
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}

// record stores a line and returns a command that prints it above the spinner.
func (m *runModel) record(raw string) tea.Cmd {
	m.res.Lines = append(m.res.Lines, raw)
	k, t := Classify(raw, m.first)
	m.first = false
	switch k {
	case KTitle:
		m.res.Title = t
	case KOK:
		m.res.OK++
	case KBad:
		m.res.Bad++
	case KVerdict:
		m.res.Verdict = t
	}
	if t != "" && k != KDetail {
		m.current = t
	}
	if !m.show {
		return nil
	}
	if s := RenderLine(k, t, m.width); s != "" {
		return tea.Println(s)
	}
	return nil
}

func (m runModel) View() string {
	if m.done {
		return ""
	}
	el := time.Since(m.start).Round(100 * time.Millisecond)
	status := Fainter.Render(Truncate(m.current, m.width-len(m.label)-20))
	return Indent + m.spin.View() + " " + Bold.Render(m.label) + " " + Fainter.Render(el.String()) + "  " + status
}

// Run executes cmd with a live spinner (or plain output when not in a terminal) and returns what
// happened. With show, every output line is drawn with the shared styles as it arrives; without,
// lines are only collected (the caller decides what to print).
func Run(label string, cmd *exec.Cmd, stream func(*exec.Cmd, chan<- string) error, show bool) Result {
	res := &Result{}
	lines := make(chan string, 64)
	errc := make(chan error, 1)
	start := time.Now()
	go func() { errc <- stream(cmd, lines) }()

	if !IsTTY() {
		first := true
		for l := range lines {
			res.Lines = append(res.Lines, l)
			k, t := Classify(l, first)
			first = false
			switch k {
			case KTitle:
				res.Title = t
			case KOK:
				res.OK++
			case KBad:
				res.Bad++
			case KVerdict:
				res.Verdict = t
			}
			if s := RenderLine(k, t, 100); s != "" && show {
				fmt.Println(s)
			}
		}
		res.Err = <-errc
		res.Duration = time.Since(start)
		return *res
	}

	sp := spinner.New(spinner.WithSpinner(spinner.Spinner{
		Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		FPS:    time.Second / 12,
	}), spinner.WithStyle(lipgloss.NewStyle().Foreground(Violet)))
	m := runModel{label: label, spin: sp, lines: lines, errc: errc, res: res, start: start, width: Width(), first: true, show: show}
	final, err := tea.NewProgram(m).Run()
	if err != nil && res.Err == nil {
		res.Err = err
	}
	if fm, ok := final.(runModel); ok && !fm.done && res.Err == nil {
		// The program ended before the script did (ctrl+c): stop it.
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		res.Err = fmt.Errorf("stopped")
	}
	res.Duration = time.Since(start)
	return *res
}

// Summary draws the closing card for a run.
func Summary(r Result, okTitle, badTitle, okText, badText string) string {
	w := Width()
	stats := []string{}
	if r.OK > 0 {
		stats = append(stats, OKStyle.Render(fmt.Sprintf("%d %s", r.OK, IconOK)))
	}
	if r.Bad > 0 {
		stats = append(stats, BadStyle.Render(fmt.Sprintf("%d %s", r.Bad, IconBad)))
	}
	stats = append(stats, Fainter.Render(r.Duration.Round(100*time.Millisecond).String()))
	if r.Err == nil && r.Bad == 0 {
		body := OKStyle.Render(okTitle) + "\n" + Highlight(okText, Dim.Render) + "\n" + strings.Join(stats, Fainter.Render("  ·  "))
		return "\n" + Card("", body, Green, w)
	}
	body := BadStyle.Render(badTitle) + "\n" + Highlight(badText, Dim.Render) + "\n" + strings.Join(stats, Fainter.Render("  ·  "))
	return "\n" + Card("", body, Red, w)
}
