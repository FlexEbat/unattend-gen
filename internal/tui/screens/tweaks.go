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

// tweakLabels is the fixed display order for the SystemTweaks checkboxes.
// tweaksToValues/valuesToTweaks below must list the fields in this same
// order.
var tweakLabels = [tweaksCount]string{
	"Disable Windows Update",
	"Disable UAC",
	"Bypass Windows 11 hardware requirements",
	"Disable Smart App Control",
	"Disable SmartScreen (Windows and Edge)",
	"Disable Fast Startup",
	"Disable System Restore",
	"Enable long paths",
	"Enable Remote Desktop (RDP)",
	"Allow PowerShell script execution",
	"Disable last access timestamp",
	"Prevent automatic device encryption",
	"Disable auto sign-on of last user after restart",
	"Disable WPBT execution",
	"Audit process creation (with command line)",
	"Hide Edge first run experience",
	"Disable Edge startup boost / background mode",
}

const tweaksCount = 17

// Tweaks is the express settings / system tweaks screen.
type Tweaks struct {
	profile     *profile.Profile
	expressMode widgets.LabeledSelect
	checks      [tweaksCount]widgets.Checkbox
	focus       int
	bar         widgets.ConfirmBar
}

// NewTweaks builds the tweaks screen backed by profile.
func NewTweaks(p *profile.Profile) Tweaks {
	t := Tweaks{
		profile:     p,
		expressMode: widgets.NewLabeledSelect("Express settings", expressModeOptions),
		bar:         widgets.NewConfirmBar("Tab: focus", "Space: toggle", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
	}
	t.expressMode.SetValue(string(p.ExpressSettings.Mode))
	values := tweaksToValues(p.SystemTweaks)
	for i, label := range tweakLabels {
		t.checks[i] = widgets.Checkbox{Label: label, Checked: values[i]}
	}
	return t
}

func tweaksToValues(tw profile.SystemTweaks) [tweaksCount]bool {
	return [tweaksCount]bool{
		tw.DisableWindowsUpdate,
		tw.DisableUAC,
		tw.BypassWin11Requirements,
		tw.DisableSmartAppControl,
		tw.DisableSmartScreen,
		tw.DisableFastStartup,
		tw.DisableSystemRestore,
		tw.EnableLongPaths,
		tw.EnableRemoteDesktop,
		tw.AllowPowerShellScripts,
		tw.DisableLastAccessTimestamp,
		tw.PreventDeviceEncryption,
		tw.DisableAutoSignOnLastUser,
		tw.DisableWPBT,
		tw.AuditProcessCreation,
		tw.HideEdgeFirstRun,
		tw.DisableEdgeStartupBoost,
	}
}

func valuesToTweaks(v [tweaksCount]bool) profile.SystemTweaks {
	return profile.SystemTweaks{
		DisableWindowsUpdate:       v[0],
		DisableUAC:                 v[1],
		BypassWin11Requirements:    v[2],
		DisableSmartAppControl:     v[3],
		DisableSmartScreen:         v[4],
		DisableFastStartup:         v[5],
		DisableSystemRestore:       v[6],
		EnableLongPaths:            v[7],
		EnableRemoteDesktop:        v[8],
		AllowPowerShellScripts:     v[9],
		DisableLastAccessTimestamp: v[10],
		PreventDeviceEncryption:    v[11],
		DisableAutoSignOnLastUser:  v[12],
		DisableWPBT:                v[13],
		AuditProcessCreation:       v[14],
		HideEdgeFirstRun:           v[15],
		DisableEdgeStartupBoost:    v[16],
	}
}

// Init is a no-op: the select list handles its own focus internally.
func (t Tweaks) Init() tea.Cmd {
	return nil
}

// fieldCount: field 0 is the express-settings select, fields 1..tweaksCount
// are the checkboxes.
const tweaksFieldCount = 1 + tweaksCount

func (t *Tweaks) sync() {
	t.profile.ExpressSettings.Mode = profile.ExpressSettingsMode(t.expressMode.Value())
	var values [tweaksCount]bool
	for i, c := range t.checks {
		values[i] = c.Checked
	}
	t.profile.SystemTweaks = valuesToTweaks(values)
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
			if t.focus >= 1 {
				idx := t.focus - 1
				t.checks[idx].Checked = !t.checks[idx].Checked
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

// View renders the express settings selector and every tweak checkbox.
func (t Tweaks) View() string {
	out := t.expressMode.View() + "\n\n"
	for i, c := range t.checks {
		out += c.View(t.focus == i+1) + "\n"
	}
	out += "\n" + t.bar.View()
	return out
}
