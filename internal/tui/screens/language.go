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
	{Value: string(profile.EditionModeFirmware), Label: "Use key stored in BIOS/UEFI firmware"},
}

var editionOptions = []widgets.SelectOption{
	{Value: string(profile.EditionHome), Label: "Home"},
	{Value: string(profile.EditionPro), Label: "Pro"},
	{Value: string(profile.EditionEducation), Label: "Education"},
	{Value: string(profile.EditionEnterprise), Label: "Enterprise"},
}

var processorArchOptions = []widgets.SelectOption{
	{Value: string(profile.ArchAMD64), Label: "x64 (amd64)"},
	{Value: string(profile.ArchX86), Label: "x86 (32-bit)"},
	{Value: string(profile.ArchARM64), Label: "ARM64"},
}

// Language is the language/locale/keyboard/edition screen. Slice 24 added
// two always-visible fields at the end (activationKey, processorArch),
// independent of edition mode: an activation-only product key (separate
// from the install key above, reused from a custom install key when left
// blank — see components.ResolveActivationKey) and the target CPU
// architecture (single value; see profile.ProcessorArchitecture for why
// this project doesn't support the reference's multi-architecture mode).
type Language struct {
	profile       *profile.Profile
	uiLanguage    widgets.LabeledInput
	locale        widgets.LabeledInput
	keyboard      widgets.LabeledInput
	editionMode   widgets.LabeledSelect
	editionChoice widgets.LabeledSelect
	productKey    widgets.LabeledInput
	activationKey widgets.LabeledInput
	processorArch widgets.LabeledSelect
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
		activationKey: widgets.NewLabeledInput("Activation key (optional, separate from install key)", "XXXXX-XXXXX-XXXXX-XXXXX-XXXXX"),
		processorArch: widgets.NewLabeledSelect("Processor architecture", processorArchOptions),
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
	if p.ActivationKey != nil {
		l.activationKey.SetValue(*p.ActivationKey)
	}
	l.processorArch.SetValue(string(p.ProcessorArchitecture))
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

// editionFieldVisible reports whether the edition-mode-specific field
// (editionChoice or productKey) is shown between editionMode and
// activationKey.
func (l Language) editionFieldVisible() bool {
	switch profile.EditionMode(l.editionMode.Value()) {
	case profile.EditionModeGenericKey, profile.EditionModeCustomKey:
		return true
	default:
		return false
	}
}

// fieldCount returns how many fields are focusable: the 4 always-present
// fields (uiLanguage/locale/keyboard/editionMode), +1 if the edition mode
// needs an extra field, +2 for activationKey/processorArch (always shown).
func (l Language) fieldCount() int {
	n := 4 + 2
	if l.editionFieldVisible() {
		n++
	}
	return n
}

// activationKeyIdx and processorArchIdx are computed, not fixed, since the
// edition-mode field shifts them by one.
func (l Language) activationKeyIdx() int {
	if l.editionFieldVisible() {
		return 5
	}
	return 4
}

func (l Language) processorArchIdx() int {
	return l.activationKeyIdx() + 1
}

func (l *Language) setFocus(i int) tea.Cmd {
	l.uiLanguage.Blur()
	l.locale.Blur()
	l.keyboard.Blur()
	l.productKey.Blur()
	l.activationKey.Blur()
	l.focus = i
	switch {
	case i == 0:
		return l.uiLanguage.Focus()
	case i == 1:
		return l.locale.Focus()
	case i == 2:
		return l.keyboard.Focus()
	case i == 4 && l.editionFieldVisible() && profile.EditionMode(l.editionMode.Value()) == profile.EditionModeCustomKey:
		return l.productKey.Focus()
	case i == l.activationKeyIdx():
		return l.activationKey.Focus()
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
	if v := l.activationKey.Value(); v != "" {
		l.profile.ActivationKey = &v
	} else {
		l.profile.ActivationKey = nil
	}
	l.profile.ProcessorArchitecture = profile.ProcessorArchitecture(l.processorArch.Value())
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
	switch {
	case l.focus == 0:
		l.uiLanguage, cmd = l.uiLanguage.Update(msg)
	case l.focus == 1:
		l.locale, cmd = l.locale.Update(msg)
	case l.focus == 2:
		l.keyboard, cmd = l.keyboard.Update(msg)
	case l.focus == 3:
		l.editionMode, cmd = l.editionMode.Update(msg)
	case l.focus == 4 && l.editionFieldVisible() && profile.EditionMode(l.editionMode.Value()) == profile.EditionModeGenericKey:
		l.editionChoice, cmd = l.editionChoice.Update(msg)
	case l.focus == 4 && l.editionFieldVisible():
		l.productKey, cmd = l.productKey.Update(msg)
	case l.focus == l.activationKeyIdx():
		l.activationKey, cmd = l.activationKey.Update(msg)
	case l.focus == l.processorArchIdx():
		l.processorArch, cmd = l.processorArch.Update(msg)
	}
	l.sync()
	return l, cmd
}

// View renders the fields relevant to the current edition mode, plus the
// always-shown activation key and processor architecture fields.
func (l Language) View() string {
	out := l.uiLanguage.View() + "\n\n" + l.locale.View() + "\n\n" + l.keyboard.View() + "\n\n" + l.editionMode.View()
	switch profile.EditionMode(l.editionMode.Value()) {
	case profile.EditionModeGenericKey:
		out += "\n\n" + l.editionChoice.View()
	case profile.EditionModeCustomKey:
		out += "\n\n" + l.productKey.View()
	}
	out += "\n\n" + l.activationKey.View()
	out += "\n\n" + l.processorArch.View()
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
	l.activationKey.Err = ""
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
		case strings.Contains(e, "activation_key"):
			l.activationKey.Err = e
		}
	}
	return l
}
