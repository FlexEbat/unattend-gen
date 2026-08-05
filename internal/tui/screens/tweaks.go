package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
)

var expressModeOptions = []widgets.SelectOption{
	{Value: string(profile.ExpressInteractive), Label: "Interactive (ask during setup)"},
	{Value: string(profile.ExpressAllEnabled), Label: "Enable all categories"},
	{Value: string(profile.ExpressAllDisabled), Label: "Disable all categories"},
}

// checkbox is a boolean field rendered as [x]/[ ], toggled by space or
// enter while focused.
type checkbox struct {
	Label   string
	Checked bool
}

func (c checkbox) View(focused bool) string {
	box := "[ ]"
	if c.Checked {
		box = "[x]"
	}
	if focused {
		return "> " + box + " " + c.Label
	}
	return "  " + box + " " + c.Label
}

// Tweaks is the express settings / system tweaks screen.
type Tweaks struct {
	profile              *profile.Profile
	expressMode          widgets.LabeledSelect
	disableWindowsUpdate checkbox
	disableUAC           checkbox
	bypassWin11          checkbox
	focus                int
	bar                  widgets.ConfirmBar
}

// NewTweaks builds the tweaks screen backed by profile.
func NewTweaks(p *profile.Profile) Tweaks {
	t := Tweaks{
		profile:              p,
		expressMode:          widgets.NewLabeledSelect("Express settings", expressModeOptions),
		disableWindowsUpdate: checkbox{Label: "Disable Windows Update"},
		disableUAC:           checkbox{Label: "Disable UAC"},
		bypassWin11:          checkbox{Label: "Bypass Windows 11 hardware requirements"},
		bar:                  widgets.NewConfirmBar("Tab: focus", "Space: toggle", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
	}
	t.expressMode.SetValue(string(p.ExpressSettings.Mode))
	t.disableWindowsUpdate.Checked = p.SystemTweaks.DisableWindowsUpdate
	t.disableUAC.Checked = p.SystemTweaks.DisableUAC
	t.bypassWin11.Checked = p.SystemTweaks.BypassWin11Requirements
	return t
}

// Init is a no-op: the select list handles its own focus internally.
func (t Tweaks) Init() tea.Cmd {
	return nil
}

const tweaksFieldCount = 4

func (t *Tweaks) sync() {
	t.profile.ExpressSettings.Mode = profile.ExpressSettingsMode(t.expressMode.Value())
	t.profile.SystemTweaks = profile.SystemTweaks{
		DisableWindowsUpdate:    t.disableWindowsUpdate.Checked,
		DisableUAC:              t.disableUAC.Checked,
		BypassWin11Requirements: t.bypassWin11.Checked,
	}
}

// Update handles focus cycling, checkbox toggling and screen navigation.
func (t Tweaks) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			t.focus = (t.focus + 1) % tweaksFieldCount
			return t, nil
		case "shift+tab":
			t.focus = (t.focus - 1 + tweaksFieldCount) % tweaksFieldCount
			return t, nil
		case " ", "enter":
			switch t.focus {
			case 1:
				t.disableWindowsUpdate.Checked = !t.disableWindowsUpdate.Checked
			case 2:
				t.disableUAC.Checked = !t.disableUAC.Checked
			case 3:
				t.bypassWin11.Checked = !t.bypassWin11.Checked
			}
			t.sync()
			return t, nil
		case "ctrl+n":
			t.sync()
			return t, Navigate(ScreenWifi)
		case "esc":
			t.sync()
			return t, Navigate(ScreenAccounts)
		case "ctrl+r":
			t.sync()
			return t, Navigate(ScreenReview)
		}
	}

	var cmd tea.Cmd
	if t.focus == 0 {
		t.expressMode, cmd = t.expressMode.Update(msg)
	}
	t.sync()
	return t, cmd
}

// View renders the express settings selector and the three checkboxes.
func (t Tweaks) View() string {
	out := t.expressMode.View() + "\n\n"
	out += t.disableWindowsUpdate.View(t.focus == 1) + "\n"
	out += t.disableUAC.View(t.focus == 2) + "\n"
	out += t.bypassWin11.View(t.focus == 3)
	out += "\n\n" + t.bar.View()
	return out
}
