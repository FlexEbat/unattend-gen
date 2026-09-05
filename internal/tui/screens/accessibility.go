package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
)

var stickyKeysModeOptions = []widgets.SelectOption{
	{Value: "", Label: "Windows default"},
	{Value: string(profile.StickyKeysModeDisabled), Label: "Disabled"},
	{Value: string(profile.StickyKeysModeCustom), Label: "Custom"},
}

var lockKeyInitialOptions = []widgets.SelectOption{
	{Value: string(profile.LockKeyOff), Label: "Off"},
	{Value: string(profile.LockKeyOn), Label: "On"},
}

var lockKeyBehaviorOptions = []widgets.SelectOption{
	{Value: string(profile.LockKeyToggle), Label: "Toggle normally"},
	{Value: string(profile.LockKeyIgnore), Label: "Ignore (key does nothing)"},
}

// Accessibility is the Sticky Keys / lock key screen. Sticky Keys fields
// are always visible (mode select, plus 6 flag checkboxes when mode is
// Custom); lock key fields (3 keys x initial state + behavior) only appear
// once "Configure lock keys" is checked - nil Profile.LockKeys otherwise,
// matching the reference implementation's skip-by-default behavior.
type Accessibility struct {
	profile *profile.Profile

	stickyMode  widgets.LabeledSelect
	stickyFlags []widgets.Checkbox

	configureLockKeys widgets.Checkbox
	capsInitial       widgets.LabeledSelect
	capsBehavior      widgets.LabeledSelect
	numInitial        widgets.LabeledSelect
	numBehavior       widgets.LabeledSelect
	scrollInitial     widgets.LabeledSelect
	scrollBehavior    widgets.LabeledSelect

	focus int
	bar   widgets.ConfirmBar
}

// NewAccessibility builds the accessibility screen backed by profile.
func NewAccessibility(p *profile.Profile) Accessibility {
	a := Accessibility{
		profile:           p,
		stickyMode:        widgets.NewLabeledSelect("Sticky Keys", stickyKeysModeOptions),
		configureLockKeys: widgets.Checkbox{Label: "Configure lock keys (Caps/Num/Scroll Lock)"},
		capsInitial:       widgets.NewLabeledSelect("Caps Lock: initial state", lockKeyInitialOptions),
		capsBehavior:      widgets.NewLabeledSelect("Caps Lock: behavior", lockKeyBehaviorOptions),
		numInitial:        widgets.NewLabeledSelect("Num Lock: initial state", lockKeyInitialOptions),
		numBehavior:       widgets.NewLabeledSelect("Num Lock: behavior", lockKeyBehaviorOptions),
		scrollInitial:     widgets.NewLabeledSelect("Scroll Lock: initial state", lockKeyInitialOptions),
		scrollBehavior:    widgets.NewLabeledSelect("Scroll Lock: behavior", lockKeyBehaviorOptions),
		stickyFlags:       make([]widgets.Checkbox, len(profile.StickyKeysFlags)),
		bar:               widgets.NewConfirmBar("Tab: focus", "Space: toggle", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
	}
	a.stickyMode.SetValue(string(p.StickyKeys.Mode))
	selectedFlags := make(map[profile.StickyKeysFlag]bool, len(p.StickyKeys.Flags))
	for _, f := range p.StickyKeys.Flags {
		selectedFlags[f] = true
	}
	for i, f := range profile.StickyKeysFlags {
		a.stickyFlags[i] = widgets.Checkbox{Label: string(f), Checked: selectedFlags[f]}
	}
	if lk := p.LockKeys; lk != nil {
		a.configureLockKeys.Checked = true
		a.capsInitial.SetValue(string(lk.CapsLock.Initial))
		a.capsBehavior.SetValue(string(lk.CapsLock.Behavior))
		a.numInitial.SetValue(string(lk.NumLock.Initial))
		a.numBehavior.SetValue(string(lk.NumLock.Behavior))
		a.scrollInitial.SetValue(string(lk.ScrollLock.Initial))
		a.scrollBehavior.SetValue(string(lk.ScrollLock.Behavior))
	} else {
		a.capsInitial.SetValue(string(profile.LockKeyOff))
		a.capsBehavior.SetValue(string(profile.LockKeyToggle))
		a.numInitial.SetValue(string(profile.LockKeyOff))
		a.numBehavior.SetValue(string(profile.LockKeyToggle))
		a.scrollInitial.SetValue(string(profile.LockKeyOff))
		a.scrollBehavior.SetValue(string(profile.LockKeyToggle))
	}
	return a
}

// Init is a no-op.
func (a Accessibility) Init() tea.Cmd {
	return nil
}

func (a Accessibility) stickyCustom() bool {
	return a.stickyMode.Value() == string(profile.StickyKeysModeCustom)
}

// fieldCount is: 1 (mode select) + 6 (flag checkboxes, only if Custom) + 1
// (configure-lock-keys checkbox) + 6 (lock key fields, only if checked).
func (a Accessibility) fieldCount() int {
	n := 1 + 1
	if a.stickyCustom() {
		n += len(a.stickyFlags)
	}
	if a.configureLockKeys.Checked {
		n += 6
	}
	return n
}

// configureLockKeysIdx is the focus index of the "configure lock keys"
// checkbox: right after the mode select and its flag checkboxes (if any).
func (a Accessibility) configureLockKeysIdx() int {
	if a.stickyCustom() {
		return 1 + len(a.stickyFlags)
	}
	return 1
}

func (a *Accessibility) sync() {
	if a.stickyCustom() {
		var flags []profile.StickyKeysFlag
		for i, c := range a.stickyFlags {
			if c.Checked {
				flags = append(flags, profile.StickyKeysFlags[i])
			}
		}
		a.profile.StickyKeys = profile.StickyKeysSettings{Mode: profile.StickyKeysModeCustom, Flags: flags}
	} else {
		a.profile.StickyKeys = profile.StickyKeysSettings{Mode: profile.StickyKeysMode(a.stickyMode.Value())}
	}

	if !a.configureLockKeys.Checked {
		a.profile.LockKeys = nil
		return
	}
	a.profile.LockKeys = &profile.LockKeySettings{
		CapsLock:   profile.LockKeySetting{Initial: profile.LockKeyInitial(a.capsInitial.Value()), Behavior: profile.LockKeyBehavior(a.capsBehavior.Value())},
		NumLock:    profile.LockKeySetting{Initial: profile.LockKeyInitial(a.numInitial.Value()), Behavior: profile.LockKeyBehavior(a.numBehavior.Value())},
		ScrollLock: profile.LockKeySetting{Initial: profile.LockKeyInitial(a.scrollInitial.Value()), Behavior: profile.LockKeyBehavior(a.scrollBehavior.Value())},
	}
}

// Update handles focus cycling, select/checkbox input and screen navigation.
func (a Accessibility) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			a.focus = (a.focus + 1) % a.fieldCount()
			return a, nil
		case "shift+tab":
			a.focus = (a.focus - 1 + a.fieldCount()) % a.fieldCount()
			return a, nil
		case " ":
			switch {
			case a.stickyCustom() && a.focus >= 1 && a.focus < 1+len(a.stickyFlags):
				i := a.focus - 1
				a.stickyFlags[i].Checked = !a.stickyFlags[i].Checked
				a.sync()
			case a.focus == a.configureLockKeysIdx():
				a.configureLockKeys.Checked = !a.configureLockKeys.Checked
				a.sync()
			}
			return a, nil
		case "ctrl+n":
			a.sync()
			return a, Navigate(ScreenScripts)
		case "esc":
			a.sync()
			return a, Navigate(ScreenPersonalization)
		case "ctrl+r":
			a.sync()
			return a, Navigate(ScreenReview)
		}
	}

	var cmd tea.Cmd
	lockBase := a.configureLockKeysIdx() + 1
	switch {
	case a.focus == 0:
		a.stickyMode, cmd = a.stickyMode.Update(msg)
	case a.configureLockKeys.Checked && a.focus == lockBase:
		a.capsInitial, cmd = a.capsInitial.Update(msg)
	case a.configureLockKeys.Checked && a.focus == lockBase+1:
		a.capsBehavior, cmd = a.capsBehavior.Update(msg)
	case a.configureLockKeys.Checked && a.focus == lockBase+2:
		a.numInitial, cmd = a.numInitial.Update(msg)
	case a.configureLockKeys.Checked && a.focus == lockBase+3:
		a.numBehavior, cmd = a.numBehavior.Update(msg)
	case a.configureLockKeys.Checked && a.focus == lockBase+4:
		a.scrollInitial, cmd = a.scrollInitial.Update(msg)
	case a.configureLockKeys.Checked && a.focus == lockBase+5:
		a.scrollBehavior, cmd = a.scrollBehavior.Update(msg)
	}
	a.sync()
	return a, cmd
}

// View renders the Sticky Keys and lock key fields.
func (a Accessibility) View() string {
	out := a.stickyMode.View()
	if a.stickyCustom() {
		out += "\n\n"
		for i, c := range a.stickyFlags {
			out += c.View(a.focus == 1+i) + "\n"
		}
	}
	out += "\n" + a.configureLockKeys.View(a.focus == a.configureLockKeysIdx())
	if a.configureLockKeys.Checked {
		out += "\n\n" + a.capsInitial.View() + "\n\n" + a.capsBehavior.View()
		out += "\n\n" + a.numInitial.View() + "\n\n" + a.numBehavior.View()
		out += "\n\n" + a.scrollInitial.View() + "\n\n" + a.scrollBehavior.View()
	}
	out += "\n\n" + a.bar.View()
	return out
}
