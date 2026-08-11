package widgets

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// LabeledTextArea is a multi-line text field with a label above it and a
// slot for an error message below it, used for script content.
type LabeledTextArea struct {
	Label string
	Err   string
	Input textarea.Model
}

// NewLabeledTextArea builds a text area with the given label and placeholder.
func NewLabeledTextArea(label, placeholder string) LabeledTextArea {
	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.SetHeight(6)
	return LabeledTextArea{Label: label, Input: ta}
}

// Focus focuses the underlying text area.
func (l *LabeledTextArea) Focus() tea.Cmd {
	return l.Input.Focus()
}

// Blur unfocuses the underlying text area.
func (l *LabeledTextArea) Blur() {
	l.Input.Blur()
}

// Value returns the current text.
func (l *LabeledTextArea) Value() string {
	return l.Input.Value()
}

// SetValue replaces the current text.
func (l *LabeledTextArea) SetValue(v string) {
	l.Input.SetValue(v)
}

// Update forwards msg to the underlying text area.
func (l LabeledTextArea) Update(msg tea.Msg) (LabeledTextArea, tea.Cmd) {
	var cmd tea.Cmd
	l.Input, cmd = l.Input.Update(msg)
	return l, cmd
}

// View renders the label, the text area and, if set, the error line.
func (l LabeledTextArea) View() string {
	out := labelStyle.Render(l.Label) + "\n" + l.Input.View()
	if l.Err != "" {
		out += "\n" + errorStyle.Render(l.Err)
	}
	return out
}
