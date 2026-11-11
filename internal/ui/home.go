package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MenuItem is one action on the home screen.
type MenuItem struct {
	Key, Icon, Title, Desc string
}

type homeModel struct {
	header, card string
	items        []MenuItem
	cursor       int
	chosen       string
	width        int
}

func (m homeModel) Init() tea.Cmd { return nil }

func (m homeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, 100)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k", "shift+tab":
			m.cursor = (m.cursor - 1 + len(m.items)) % len(m.items)
		case "down", "j", "tab":
			m.cursor = (m.cursor + 1) % len(m.items)
		case "home", "g":
			m.cursor = 0
		case "end", "G":
			m.cursor = len(m.items) - 1
		case "enter", " ":
			m.chosen = m.items[m.cursor].Key
			return m, tea.Quit
		default:
			// Number keys and first letters jump straight to an item.
			k := msg.String()
			for i, it := range m.items {
				if k == string(rune('1'+i)) || strings.EqualFold(k, it.Key[:1]) && len(k) == 1 {
					m.cursor = i
					if k >= "1" && k <= "9" {
						m.chosen = it.Key
						return m, tea.Quit
					}
					break
				}
			}
		}
	}
	return m, nil
}

func (m homeModel) View() string {
	if m.chosen != "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n" + m.header + "\n\n")
	if m.card != "" {
		b.WriteString(m.card + "\n\n")
	}
	for i, it := range m.items {
		sel := i == m.cursor
		num := Fainter.Render(string(rune('1' + i)))
		icon := Dim.Render(it.Icon)
		title := Text.Render(it.Title)
		desc := Fainter.Render(Truncate(it.Desc, m.width-30))
		bar := "  "
		if sel {
			bar = Accent.Render("▌ ")
			icon = Accent.Render(it.Icon)
			title = Accent.Render(it.Title)
			desc = Dim.Render(Truncate(it.Desc, m.width-30))
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, bar, num, "  ", icon, "  ",
			lipgloss.NewStyle().Width(16).Render(title), desc)
		b.WriteString(Indent + row + "\n")
	}
	b.WriteString("\n" + Indent + strings.Join([]string{Key("↑↓", "move"), Key("enter", "run"), Key("1-9", "jump"), Key("q", "quit")}, Fainter.Render("   ")) + "\n")
	return b.String()
}

// Home shows the interactive start screen and returns the chosen item's key ("" if the user quit).
func Home(header, card string, items []MenuItem) (string, error) {
	m := homeModel{header: header, card: card, items: items, width: Width()}
	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", err
	}
	return final.(homeModel).chosen, nil
}
