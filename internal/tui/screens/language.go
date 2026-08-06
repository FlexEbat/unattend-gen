package screens

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
)

var editionModeOptions = []widgets.SelectOption{
	{Value: string(profile.EditionModeInteractive), Label: "Interactive (ask during setup)"},
	{Value: string(profile.EditionModeGenericKey), Label: "Generic key (pick an edition)"},
	{Value: string(profile.EditionModeCustomKey), Label: "Custom product key"},
}

var editionOptions = []widgets.SelectOption{
	{Value: string(profile.EditionHome), Label: "Home"},
	{Value: string(profile.EditionPro), Label: "Pro"},
	{Value: string(profile.EditionEducation), Label: "Education"},
	{Value: string(profile.EditionEnterprise), Label: "Enterprise"},
}

// Language is the language/locale/keyboard/edition screen.
type Language struct {
	profile       *profile.Profile
	uiLanguage    widgets.LabeledInput
	locale        widgets.LabeledInput
	keyboard      widgets.LabeledInput
	editionMode   widgets.LabeledSelect
	editionChoice widgets.LabeledSelect
	productKey    widgets.LabeledInput
	focus         int
	bar           widgets.ConfirmBar
}

// NewLanguage builds the language screen backed by profile.
func NewLanguage(p *profile.Profile) Language {
	l := Language{
		profile:       p,
		uiLanguage:    widgets.NewLabeledInput("UI language (BCP-47)", "en-US"),
		locale:        widgets.NewLabeledInput("Locale (BCP-47)", "en-US"),
		keyboard:      widgets.NewLabeledInput("Keyboard layout (BCP-47)", "en-US"),
		editionMode:   widgets.NewLabeledSelect("Edition", editionModeOptions),
		editionChoice: widgets.NewLabeledSelect("Which edition", editionOptions),
		productKey:    widgets.NewLabeledInput("Product key", "XXXXX-XXXXX-XXXXX-XXXXX-XXXXX"),
		bar:           widgets.NewConfirmBar("Tab: next field", "Ctrl+N: next screen", "Esc: back", "Ctrl+R: review"),
	}
	l.uiLanguage.SetValue(p.Language.UILanguage)
	l.locale.SetValue(p.Language.Locale)
	l.keyboard.SetValue(p.Language.KeyboardLayout)
	l.editionMode.SetValue(string(p.Edition.Mode))
	if p.Edition.Edition != nil {
		l.editionChoice.SetValue(string(*p.Edition.Edition))
	}
	if p.Edition.ProductKey != nil {
		l.productKey.SetValue(*p.Edition.ProductKey)
	}
	l.focus = 0
	// Set focus here, not in Init: bubbletea's Init only returns a tea.Cmd,
	// it cannot mutate the model, so a Focus() call made there would be lost.
	l.uiLanguage.Focus()
	return l
}

// Init returns the cursor-blink command for the focused field. The focus
// state itself was already set in NewLanguage.
func (l Language) Init() tea.Cmd {
	return textinput.Blink
}

// fieldCount returns how many fields are focusable given the current
// edition mode.
func (l Language) fieldCount() int {
	switch profile.EditionMode(l.editionMode.Value()) {
	case profile.EditionModeGenericKey, profile.EditionModeCustomKey:
		return 5
	default:
		return 4
	}
}

func (l *Language) setFocus(i int) tea.Cmd {
	l.uiLanguage.Blur()
	l.locale.Blur()
	l.keyboard.Blur()
	l.productKey.Blur()
	l.focus = i
	switch i {
	case 0:
		return l.uiLanguage.Focus()
	case 1:
		return l.locale.Focus()
	case 2:
		return l.keyboard.Focus()
	case 4:
		if profile.EditionMode(l.editionMode.Value()) == profile.EditionModeCustomKey {
			return l.productKey.Focus()
		}
	}
	return nil
}

func (l *Language) sync() {
	l.profile.Language.UILanguage = l.uiLanguage.Value()
	l.profile.Language.Locale = l.locale.Value()
	l.profile.Language.KeyboardLayout = l.keyboard.Value()
	l.profile.Edition.Mode = profile.EditionMode(l.editionMode.Value())
	switch l.profile.Edition.Mode {
	case profile.EditionModeGenericKey:
		edition := profile.WindowsEdition(l.editionChoice.Value())
		l.profile.Edition.Edition = &edition
		l.profile.Edition.ProductKey = nil
	case profile.EditionModeCustomKey:
		key := l.productKey.Value()
		l.profile.Edition.ProductKey = &key
		l.profile.Edition.Edition = nil
	default:
		l.profile.Edition.Edition = nil
		l.profile.Edition.ProductKey = nil
	}
}

// Update handles focus cycling, screen navigation and field input.
func (l Language) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			cmd := l.setFocus((l.focus + 1) % l.fieldCount())
			l.sync()
			return l, cmd
		case "shift+tab":
			cmd := l.setFocus((l.focus - 1 + l.fieldCount()) % l.fieldCount())
			l.sync()
			return l, cmd
		case "ctrl+n":
			l.sync()
			return l, Navigate(ScreenAccounts)
		case "esc":
			l.sync()
			return l, Navigate(ScreenWelcome)
		case "ctrl+r":
			l.sync()
			return l, Navigate(ScreenReview)
		}
	}

	var cmd tea.Cmd
	switch l.focus {
	case 0:
		l.uiLanguage, cmd = l.uiLanguage.Update(msg)
	case 1:
		l.locale, cmd = l.locale.Update(msg)
	case 2:
		l.keyboard, cmd = l.keyboard.Update(msg)
	case 3:
		l.editionMode, cmd = l.editionMode.Update(msg)
	case 4:
		if profile.EditionMode(l.editionMode.Value()) == profile.EditionModeGenericKey {
			l.editionChoice, cmd = l.editionChoice.Update(msg)
		} else {
			l.productKey, cmd = l.productKey.Update(msg)
		}
	}
	l.sync()
	return l, cmd
}

// View renders the fields relevant to the current edition mode.
func (l Language) View() string {
	out := l.uiLanguage.View() + "\n\n" + l.locale.View() + "\n\n" + l.keyboard.View() + "\n\n" + l.editionMode.View()
	switch profile.EditionMode(l.editionMode.Value()) {
	case profile.EditionModeGenericKey:
		out += "\n\n" + l.editionChoice.View()
	case profile.EditionModeCustomKey:
		out += "\n\n" + l.productKey.View()
	}
	out += "\n\n" + l.bar.View()
	return out
}

// SetErrors routes validate.go's messages to the field each one names.
// Implements screens.ErrorReceiver.
func (l Language) SetErrors(errs []string) tea.Model {
	l.uiLanguage.Err = ""
	l.locale.Err = ""
	l.keyboard.Err = ""
	l.editionChoice.Err = ""
	l.productKey.Err = ""
	for _, e := range errs {
		switch {
		case strings.Contains(e, "ui_language"):
			l.uiLanguage.Err = e
		case strings.Contains(e, "locale"):
			l.locale.Err = e
		case strings.Contains(e, "keyboard_layout"):
			l.keyboard.Err = e
		case strings.Contains(e, "generic_key"):
			l.editionChoice.Err = e
		case strings.Contains(e, "custom_key"):
			l.productKey.Err = e
		}
	}
	return l
}
