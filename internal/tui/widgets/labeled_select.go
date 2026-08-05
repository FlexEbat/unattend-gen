package widgets

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// selectItem is one option of a LabeledSelect. Value is the enum constant
// carried by the option; Label is what the user sees.
type selectItem struct {
	Label string
	Value string
}

func (i selectItem) FilterValue() string { return i.Label }
func (i selectItem) Title() string       { return i.Label }
func (i selectItem) Description() string { return "" }

// LabeledSelect is a compact single-select list for an enum-typed field.
type LabeledSelect struct {
	Label string
	Err   string
	list  list.Model
}

// SelectOption is one choice offered by a LabeledSelect: Value is the enum
// constant, Label is the text shown to the user.
type SelectOption struct {
	Value string
	Label string
}

// NewLabeledSelect builds a select list from options, sized for its content.
func NewLabeledSelect(label string, options []SelectOption) LabeledSelect {
	items := make([]list.Item, len(options))
	for i, o := range options {
		items[i] = selectItem{Label: o.Label, Value: o.Value}
	}
	l := list.New(items, list.NewDefaultDelegate(), 40, len(options)*3+2)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)
	return LabeledSelect{Label: label, list: l}
}

// Value returns the currently selected option's Value, or "" if empty.
func (s LabeledSelect) Value() string {
	if item, ok := s.list.SelectedItem().(selectItem); ok {
		return item.Value
	}
	return ""
}

// SetValue moves the selection to the option whose Value matches v. No-op if
// v is not one of the options.
func (s *LabeledSelect) SetValue(v string) {
	for i, it := range s.list.Items() {
		if item, ok := it.(selectItem); ok && item.Value == v {
			s.list.Select(i)
			return
		}
	}
}

// Update forwards msg to the underlying list.
func (s LabeledSelect) Update(msg tea.Msg) (LabeledSelect, tea.Cmd) {
	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

// View renders the label, the list and, if set, the error line.
func (s LabeledSelect) View() string {
	out := labelStyle.Render(s.Label) + "\n" + s.list.View()
	if s.Err != "" {
		out += "\n" + errorStyle.Render(s.Err)
	}
	return out
}
