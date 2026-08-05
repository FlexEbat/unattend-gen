package widgets

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// PasswordInput is a LabeledInput with masked input (EchoPassword).
type PasswordInput struct {
	LabeledInput
}

// NewPasswordInput builds a masked text input with the given label.
func NewPasswordInput(label string) PasswordInput {
	ti := textinput.New()
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '*'
	ti.CharLimit = 256
	return PasswordInput{LabeledInput{Label: label, Input: ti}}
}

// Update forwards msg to the underlying input.
func (p PasswordInput) Update(msg tea.Msg) (PasswordInput, tea.Cmd) {
	var cmd tea.Cmd
	p.Input, cmd = p.Input.Update(msg)
	return p, cmd
}
