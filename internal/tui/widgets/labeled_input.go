// Package widgets holds the TUI primitives shared by all screens: labeled
// text inputs, a password input, a labeled select list, an accounts table
// and a bottom confirm bar. Screens must use these instead of building their
// own input/table handling.
package widgets

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	labelStyle = lipgloss.NewStyle().Bold(true)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

// LabeledInput is a single-line text field with a label above it and a slot
// for an error message below it.
type LabeledInput struct {
	Label string
	Err   string
	Input textinput.Model
}

// NewLabeledInput builds a focused-capable text input with the given label
// and placeholder.
func NewLabeledInput(label, placeholder string) LabeledInput {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 256
	return LabeledInput{Label: label, Input: ti}
}

// Focus focuses the underlying input.
func (l *LabeledInput) Focus() tea.Cmd {
	return l.Input.Focus()
}

// Blur unfocuses the underlying input.
func (l *LabeledInput) Blur() {
	l.Input.Blur()
}

// Value returns the current text.
func (l *LabeledInput) Value() string {
	return l.Input.Value()
}

// SetValue replaces the current text.
func (l *LabeledInput) SetValue(v string) {
	l.Input.SetValue(v)
}

// Update forwards msg to the underlying input.
func (l LabeledInput) Update(msg tea.Msg) (LabeledInput, tea.Cmd) {
	var cmd tea.Cmd
	l.Input, cmd = l.Input.Update(msg)
	return l, cmd
}

// View renders the label, the input and, if set, the error line.
func (l LabeledInput) View() string {
	out := labelStyle.Render(l.Label) + "\n" + l.Input.View()
	if l.Err != "" {
		out += "\n" + errorStyle.Render(l.Err)
	}
	return out
}
