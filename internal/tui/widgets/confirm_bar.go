package widgets

import "github.com/charmbracelet/lipgloss"

var hintStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("245")).
	Padding(0, 1)

// ConfirmBar renders the bottom hint bar: hotkeys and, if set, a status
// message (e.g. a validation error blocking navigation).
type ConfirmBar struct {
	Hints   []string
	Message string
}

// NewConfirmBar builds a bar with the given hotkey hints, e.g.
// "Tab: next field", "Enter: next screen", "Esc: back".
func NewConfirmBar(hints ...string) ConfirmBar {
	return ConfirmBar{Hints: hints}
}

// View renders the hints, joined by " | ", and the message on its own line
// if set.
func (c ConfirmBar) View() string {
	out := ""
	for i, h := range c.Hints {
		if i > 0 {
			out += " | "
		}
		out += h
	}
	rendered := hintStyle.Render(out)
	if c.Message != "" {
		rendered += "\n" + errorStyle.Render(c.Message)
	}
	return rendered
}
