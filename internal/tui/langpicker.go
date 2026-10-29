package tui

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/i18n"
)

// ErrLanguageCancelled is returned by PickLanguage when the user quits the
// picker (Esc, q or Ctrl+C) without choosing.
var ErrLanguageCancelled = errors.New("language selection cancelled")

// languagePicker is the first thing shown on every start of the TUI. It is
// deliberately independent of the translated UI: names are in their own
// language and the heading lists the word for "language" in all of them.
type languagePicker struct {
	cursor int
	chosen i18n.Lang
}

func newLanguagePicker(def i18n.Lang) languagePicker {
	p := languagePicker{}
	for i, info := range i18n.Languages {
		if info.Code == def {
			p.cursor = i
		}
	}
	return p
}

func (p languagePicker) Init() tea.Cmd { return nil }

func (p languagePicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return p, nil
	}
	n := len(i18n.Languages)
	switch key := k.String(); key {
	case "ctrl+c", "esc", "q":
		return p, tea.Quit
	case "up", "k":
		p.cursor = (p.cursor + n - 1) % n
	case "down", "j":
		p.cursor = (p.cursor + 1) % n
	case "enter":
		p.chosen = i18n.Languages[p.cursor].Code
		return p, tea.Quit
	default:
		if len(key) == 1 && key[0] >= '1' && int(key[0]-'1') < n { // 1..n picks directly
			p.chosen = i18n.Languages[key[0]-'1'].Code
			return p, tea.Quit
		}
	}
	return p, nil
}

func (p languagePicker) View() string {
	var b strings.Builder
	b.WriteString("unattend-gen\n\n")
	b.WriteString("Language · Язык · 语言 · Idioma · भाषा\n\n")
	for i, info := range i18n.Languages {
		marker := "  "
		if i == p.cursor {
			marker = "> "
		}
		fmt.Fprintf(&b, "%s%d. %s\n", marker, i+1, info.Name)
	}
	b.WriteString("\n↑/↓ · Enter · 1-" + fmt.Sprint(len(i18n.Languages)) + "\n")
	return b.String()
}

// PickLanguage asks which interface language to use and returns the choice.
// def is the entry that is pre-selected.
func PickLanguage(def i18n.Lang) (i18n.Lang, error) {
	final, err := tea.NewProgram(newLanguagePicker(def), tea.WithAltScreen()).Run()
	if err != nil {
		return "", fmt.Errorf("language picker: %w", err)
	}
	if chosen := final.(languagePicker).chosen; chosen != "" {
		return chosen, nil
	}
	return "", ErrLanguageCancelled
}
