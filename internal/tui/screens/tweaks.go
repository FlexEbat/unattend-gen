package screens

import (
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
)

var expressModeOptions = []widgets.SelectOption{
	{Value: string(profile.ExpressInteractive), Label: "Interactive (ask during setup)"},
	{Value: string(profile.ExpressAllEnabled), Label: "Enable all categories"},
	{Value: string(profile.ExpressAllDisabled), Label: "Disable all categories"},
}

var passwordExpirationOptions = []widgets.SelectOption{
	{Value: string(profile.PasswordExpirationDefault), Label: "Windows default (42 days)"},
	{Value: string(profile.PasswordExpirationNever), Label: "Never expires"},
	{Value: string(profile.PasswordExpirationCustom), Label: "Custom (days)"},
}

var accountLockoutOptions = []widgets.SelectOption{
	{Value: string(profile.AccountLockoutDefault), Label: "Windows default (10 attempts / 10 min)"},
	{Value: string(profile.AccountLockoutDisabled), Label: "Disabled"},
	{Value: string(profile.AccountLockoutCustom), Label: "Custom"},
}

var hiddenFilesOptions = []widgets.SelectOption{
	{Value: string(profile.HiddenFilesDefault), Label: "File Explorer default (don't show hidden files)"},
	{Value: string(profile.HiddenFilesShowHidden), Label: "Show hidden files"},
	{Value: string(profile.HiddenFilesShowAll), Label: "Show hidden and protected OS files"},
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

// Tweaks is the express settings / system tweaks / account policy screen.
type Tweaks struct {
	profile     *profile.Profile
	expressMode widgets.LabeledSelect
	checks      [tweaksCount]widgets.Checkbox

	passwordExpirationMode widgets.LabeledSelect
	passwordExpirationDays widgets.LabeledInput

	accountLockoutMode      widgets.LabeledSelect
	accountLockoutThreshold widgets.LabeledInput
	accountLockoutWindow    widgets.LabeledInput
	accountLockoutDuration  widgets.LabeledInput

	fileExplorerHiddenFiles  widgets.LabeledSelect
	fileExplorerShowExt      widgets.Checkbox
	fileExplorerClassicMenu  widgets.Checkbox
	fileExplorerHideTooltips widgets.Checkbox
	fileExplorerOpenThisPC   widgets.Checkbox
	fileExplorerShowEndTask  widgets.Checkbox

	focus int
	bar   widgets.ConfirmBar
}

// NewTweaks builds the tweaks screen backed by profile.
func NewTweaks(p *profile.Profile) Tweaks {
	t := Tweaks{
		profile:                  p,
		expressMode:              widgets.NewLabeledSelect("Express settings", expressModeOptions),
		passwordExpirationMode:   widgets.NewLabeledSelect("Password expiration", passwordExpirationOptions),
		passwordExpirationDays:   widgets.NewLabeledInput("Days", "90"),
		accountLockoutMode:       widgets.NewLabeledSelect("Account lockout policy", accountLockoutOptions),
		accountLockoutThreshold:  widgets.NewLabeledInput("Failed attempts", "10"),
		accountLockoutWindow:     widgets.NewLabeledInput("Window (minutes)", "10"),
		accountLockoutDuration:   widgets.NewLabeledInput("Duration (minutes)", "10"),
		fileExplorerHiddenFiles:  widgets.NewLabeledSelect("Hidden files", hiddenFilesOptions),
		fileExplorerShowExt:      widgets.Checkbox{Label: "Always show file extensions"},
		fileExplorerClassicMenu:  widgets.Checkbox{Label: "Classic right-click context menu (Windows 11)"},
		fileExplorerHideTooltips: widgets.Checkbox{Label: "Hide folder/desktop item tooltips"},
		fileExplorerOpenThisPC:   widgets.Checkbox{Label: "Open File Explorer to This PC (not Quick access)"},
		fileExplorerShowEndTask:  widgets.Checkbox{Label: "Show End task in the taskbar right-click menu"},
		bar:                      widgets.NewConfirmBar("Tab: focus", "Space: toggle", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
	}
	t.expressMode.SetValue(string(p.ExpressSettings.Mode))
	values := tweaksToValues(p.SystemTweaks)
	for i, label := range tweakLabels {
		t.checks[i] = widgets.Checkbox{Label: label, Checked: values[i]}
	}

	t.passwordExpirationMode.SetValue(string(profile.PasswordExpirationDefault))
	if p.PasswordExpiration.Mode != "" {
		t.passwordExpirationMode.SetValue(string(p.PasswordExpiration.Mode))
	}
	if p.PasswordExpiration.Days != nil {
		t.passwordExpirationDays.SetValue(strconv.Itoa(*p.PasswordExpiration.Days))
	}

	t.accountLockoutMode.SetValue(string(profile.AccountLockoutDefault))
	if p.AccountLockout.Mode != "" {
		t.accountLockoutMode.SetValue(string(p.AccountLockout.Mode))
	}
	if p.AccountLockout.Threshold != nil {
		t.accountLockoutThreshold.SetValue(strconv.Itoa(*p.AccountLockout.Threshold))
	}
	if p.AccountLockout.WindowMinutes != nil {
		t.accountLockoutWindow.SetValue(strconv.Itoa(*p.AccountLockout.WindowMinutes))
	}
	if p.AccountLockout.DurationMinutes != nil {
		t.accountLockoutDuration.SetValue(strconv.Itoa(*p.AccountLockout.DurationMinutes))
	}

	t.fileExplorerHiddenFiles.SetValue(string(profile.HiddenFilesDefault))
	if p.FileExplorer.HiddenFiles != "" {
		t.fileExplorerHiddenFiles.SetValue(string(p.FileExplorer.HiddenFiles))
	}
	t.fileExplorerShowExt.Checked = p.FileExplorer.ShowFileExtensions
	t.fileExplorerClassicMenu.Checked = p.FileExplorer.ClassicContextMenu
	t.fileExplorerHideTooltips.Checked = p.FileExplorer.HideFolderTooltips
	t.fileExplorerOpenThisPC.Checked = p.FileExplorer.OpenToThisPC
	t.fileExplorerShowEndTask.Checked = p.FileExplorer.ShowEndTaskInTaskbar
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

// Init returns the cursor-blink command for the text inputs; the focus
// state itself doesn't need setting here (default focus 0 is the
// express-settings select, which doesn't need an explicit Focus() call).
func (t Tweaks) Init() tea.Cmd {
	return textinput.Blink
}

// Field layout (indices, computed since two groups are conditional):
//
//	0                                  expressMode
//	1..tweaksCount                     checks
//	pwModeIdx                          passwordExpirationMode
//	pwDaysIdx (only if mode==custom)   passwordExpirationDays
//	lockoutModeIdx                     accountLockoutMode
//	lockoutModeIdx+1..+3 (if custom)   threshold, window, duration
const pwModeIdx = 1 + tweaksCount

func (t Tweaks) hasPasswordDays() bool {
	return profile.PasswordExpirationMode(t.passwordExpirationMode.Value()) == profile.PasswordExpirationCustom
}

func (t Tweaks) hasLockoutCustomFields() bool {
	return profile.AccountLockoutMode(t.accountLockoutMode.Value()) == profile.AccountLockoutCustom
}

func (t Tweaks) lockoutModeIdx() int {
	idx := pwModeIdx + 1
	if t.hasPasswordDays() {
		idx++
	}
	return idx
}

// fileExplorerBaseIdx is the index of the first File Explorer field
// (the HiddenFiles select); the 5 checkboxes follow it at +1..+5.
func (t Tweaks) fileExplorerBaseIdx() int {
	idx := t.lockoutModeIdx() + 1
	if t.hasLockoutCustomFields() {
		idx += 3
	}
	return idx
}

func (t Tweaks) fieldCount() int {
	return t.fileExplorerBaseIdx() + 6
}

func (t *Tweaks) sync() {
	t.profile.ExpressSettings.Mode = profile.ExpressSettingsMode(t.expressMode.Value())
	var values [tweaksCount]bool
	for i, c := range t.checks {
		values[i] = c.Checked
	}
	t.profile.SystemTweaks = valuesToTweaks(values)

	t.profile.PasswordExpiration = profile.PasswordExpirationSettings{
		Mode: profile.PasswordExpirationMode(t.passwordExpirationMode.Value()),
	}
	if t.hasPasswordDays() {
		if days, err := strconv.Atoi(t.passwordExpirationDays.Value()); err == nil {
			t.profile.PasswordExpiration.Days = &days
		}
	}

	t.profile.AccountLockout = profile.AccountLockoutSettings{
		Mode: profile.AccountLockoutMode(t.accountLockoutMode.Value()),
	}
	if t.hasLockoutCustomFields() {
		if v, err := strconv.Atoi(t.accountLockoutThreshold.Value()); err == nil {
			t.profile.AccountLockout.Threshold = &v
		}
		if v, err := strconv.Atoi(t.accountLockoutWindow.Value()); err == nil {
			t.profile.AccountLockout.WindowMinutes = &v
		}
		if v, err := strconv.Atoi(t.accountLockoutDuration.Value()); err == nil {
			t.profile.AccountLockout.DurationMinutes = &v
		}
	}

	t.profile.FileExplorer = profile.FileExplorerSettings{
		HiddenFiles:          profile.HiddenFilesMode(t.fileExplorerHiddenFiles.Value()),
		ShowFileExtensions:   t.fileExplorerShowExt.Checked,
		ClassicContextMenu:   t.fileExplorerClassicMenu.Checked,
		HideFolderTooltips:   t.fileExplorerHideTooltips.Checked,
		OpenToThisPC:         t.fileExplorerOpenThisPC.Checked,
		ShowEndTaskInTaskbar: t.fileExplorerShowEndTask.Checked,
	}
}

// fileExplorerCheckboxAt returns a pointer to the File Explorer checkbox at
// overall focus index i, or nil if i isn't one of them.
func (t *Tweaks) fileExplorerCheckboxAt(i int) *widgets.Checkbox {
	base := t.fileExplorerBaseIdx()
	switch i {
	case base + 1:
		return &t.fileExplorerShowExt
	case base + 2:
		return &t.fileExplorerClassicMenu
	case base + 3:
		return &t.fileExplorerHideTooltips
	case base + 4:
		return &t.fileExplorerOpenThisPC
	case base + 5:
		return &t.fileExplorerShowEndTask
	default:
		return nil
	}
}

// Update handles focus cycling, checkbox toggling, select/text input and
// screen navigation.
func (t Tweaks) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			t.focus = (t.focus + 1) % t.fieldCount()
			return t, nil
		case "shift+tab":
			t.focus = (t.focus - 1 + t.fieldCount()) % t.fieldCount()
			return t, nil
		case " ":
			if t.focus >= 1 && t.focus <= tweaksCount {
				idx := t.focus - 1
				t.checks[idx].Checked = !t.checks[idx].Checked
				t.sync()
				return t, nil
			}
			if box := t.fileExplorerCheckboxAt(t.focus); box != nil {
				box.Checked = !box.Checked
				t.sync()
				return t, nil
			}
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

	lockoutModeIdx := t.lockoutModeIdx()
	var cmd tea.Cmd
	switch {
	case t.focus == 0:
		t.expressMode, cmd = t.expressMode.Update(msg)
	case t.focus == pwModeIdx:
		t.passwordExpirationMode, cmd = t.passwordExpirationMode.Update(msg)
	case t.hasPasswordDays() && t.focus == pwModeIdx+1:
		t.passwordExpirationDays, cmd = t.passwordExpirationDays.Update(msg)
	case t.focus == lockoutModeIdx:
		t.accountLockoutMode, cmd = t.accountLockoutMode.Update(msg)
	case t.hasLockoutCustomFields() && t.focus == lockoutModeIdx+1:
		t.accountLockoutThreshold, cmd = t.accountLockoutThreshold.Update(msg)
	case t.hasLockoutCustomFields() && t.focus == lockoutModeIdx+2:
		t.accountLockoutWindow, cmd = t.accountLockoutWindow.Update(msg)
	case t.hasLockoutCustomFields() && t.focus == lockoutModeIdx+3:
		t.accountLockoutDuration, cmd = t.accountLockoutDuration.Update(msg)
	case t.focus == t.fileExplorerBaseIdx():
		t.fileExplorerHiddenFiles, cmd = t.fileExplorerHiddenFiles.Update(msg)
	}
	t.sync()
	return t, cmd
}

// View renders the express settings selector, every tweak checkbox and the
// password/lockout policy fields.
func (t Tweaks) View() string {
	out := t.expressMode.View() + "\n\n"
	for i, c := range t.checks {
		out += c.View(t.focus == i+1) + "\n"
	}
	out += "\n" + t.passwordExpirationMode.View()
	if t.hasPasswordDays() {
		out += "\n" + t.passwordExpirationDays.View()
	}
	out += "\n\n" + t.accountLockoutMode.View()
	if t.hasLockoutCustomFields() {
		out += "\n" + t.accountLockoutThreshold.View()
		out += "\n" + t.accountLockoutWindow.View()
		out += "\n" + t.accountLockoutDuration.View()
	}
	base := t.fileExplorerBaseIdx()
	out += "\n\n" + t.fileExplorerHiddenFiles.View()
	out += "\n" + t.fileExplorerShowExt.View(t.focus == base+1)
	out += "\n" + t.fileExplorerClassicMenu.View(t.focus == base+2)
	out += "\n" + t.fileExplorerHideTooltips.View(t.focus == base+3)
	out += "\n" + t.fileExplorerOpenThisPC.View(t.focus == base+4)
	out += "\n" + t.fileExplorerShowEndTask.View(t.focus == base+5)
	out += "\n\n" + t.bar.View()
	return out
}
