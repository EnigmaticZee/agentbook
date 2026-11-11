package ui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// FormTheme is huh's form styling in agentbook colours.
func FormTheme() *huh.Theme {
	t := huh.ThemeBase()
	t.Focused.Base = t.Focused.Base.BorderForeground(Brand)
	t.Focused.Title = t.Focused.Title.Foreground(Ink).Bold(true)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(Brand).Bold(true)
	t.Focused.Description = t.Focused.Description.Foreground(Muted)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(Red)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(Red)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(Brand)
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(Brand)
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(Brand)
	t.Focused.Option = t.Focused.Option.Foreground(Ink)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(Brand)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(Brand)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(Green).SetString("◉ ")
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(Faint).SetString("○ ")
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(Ink)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(lipgloss.Color("#FFFFFF")).Background(Brand).Bold(true)
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(Muted).Background(Chip)
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(Gold)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(Faint)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(Brand)
	t.Blurred = t.Focused
	t.Blurred.Base = t.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.Title = t.Blurred.Title.Foreground(Muted).Bold(false)
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()
	t.Help.ShortKey = t.Help.ShortKey.Foreground(Brand)
	t.Help.ShortDesc = t.Help.ShortDesc.Foreground(Faint)
	t.Help.ShortSeparator = t.Help.ShortSeparator.Foreground(Faint)
	return t
}

// Confirm asks a yes/no question with the agentbook look. Default is no.
func Confirm(title, desc, yes, no string) (bool, error) {
	ok := false
	err := huh.NewConfirm().Title(title).Description(desc).Affirmative(yes).Negative(no).Value(&ok).
		WithTheme(FormTheme()).Run()
	return ok, err
}
