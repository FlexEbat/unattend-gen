package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
)

var vmGuestToolLabels = map[profile.VMGuestTool]string{
	profile.VMGuestToolVBoxGuestAdditions: "VirtualBox Guest Additions",
	profile.VMGuestToolVMwareTools:        "VMware Tools",
	profile.VMGuestToolVirtIOGuestTools:   "VirtIO Guest Tools",
	profile.VMGuestToolParallelsTools:     "Parallels Tools",
}

// Advanced is the catch-all screen for niche settings (tech.md backlog
// groups C and E): VM guest tools, AppLocker policy, dynamic computer
// name, and three small setup-stage checkboxes. Fields, in focus order:
// vmToolChecks (0..3), appLocker textarea (4), computerNameScript textarea
// (5), keepSensitiveFiles/useNarrator/obscurePasswords checkboxes (6..8).
type Advanced struct {
	profile *profile.Profile

	vmToolChecks          []widgets.Checkbox
	appLocker             widgets.LabeledTextArea
	computerNameScript    widgets.LabeledTextArea
	keepSensitiveFiles    widgets.Checkbox
	useNarrator           widgets.Checkbox
	obscurePasswords      widgets.Checkbox
	hidePowerShellWindows widgets.Checkbox

	focus int
	bar   widgets.ConfirmBar
}

var advancedFieldCount = len(profile.VMGuestTools) + 6 // + 2 text areas + 4 checkboxes

// NewAdvanced builds the advanced screen backed by profile.
func NewAdvanced(p *profile.Profile) Advanced {
	a := Advanced{
		profile:               p,
		vmToolChecks:          make([]widgets.Checkbox, len(profile.VMGuestTools)),
		appLocker:             widgets.NewLabeledTextArea("AppLocker policy XML (optional, raw XML)", ""),
		computerNameScript:    widgets.NewLabeledTextArea("Computer name script (PowerShell, outputs the new name; overrides a static computer name)", ""),
		keepSensitiveFiles:    widgets.Checkbox{Label: "Keep unattend.xml/Wi-Fi profile after setup (default: deleted)"},
		useNarrator:           widgets.Checkbox{Label: "Start Narrator automatically during setup and first logon"},
		obscurePasswords:      widgets.Checkbox{Label: "Obscure account passwords in the generated XML (Base64, not encryption)"},
		hidePowerShellWindows: widgets.Checkbox{Label: "Hide PowerShell windows during setup"},
		bar:                   widgets.NewConfirmBar("Tab: focus", "Space: toggle", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
	}
	selected := make(map[profile.VMGuestTool]bool, len(p.InstallVMGuestTools))
	for _, t := range p.InstallVMGuestTools {
		selected[t] = true
	}
	for i, t := range profile.VMGuestTools {
		a.vmToolChecks[i] = widgets.Checkbox{Label: vmGuestToolLabels[t], Checked: selected[t]}
	}
	if p.AppLockerPolicyXML != nil {
		a.appLocker.SetValue(*p.AppLockerPolicyXML)
	}
	if p.ComputerNameScript != nil {
		a.computerNameScript.SetValue(*p.ComputerNameScript)
	}
	a.keepSensitiveFiles.Checked = p.KeepSensitiveFiles
	a.useNarrator.Checked = p.UseNarrator
	a.obscurePasswords.Checked = p.ObscurePasswords
	a.hidePowerShellWindows.Checked = p.HidePowerShellWindows
	return a
}

// Init is a no-op.
func (a Advanced) Init() tea.Cmd {
	return nil
}

var (
	advancedAppLockerIdx             = len(profile.VMGuestTools)
	advancedComputerNameScriptIdx    = advancedAppLockerIdx + 1
	advancedKeepSensitiveFilesIdx    = advancedComputerNameScriptIdx + 1
	advancedUseNarratorIdx           = advancedKeepSensitiveFilesIdx + 1
	advancedObscurePasswordsIdx      = advancedUseNarratorIdx + 1
	advancedHidePowerShellWindowsIdx = advancedObscurePasswordsIdx + 1
)

func (a *Advanced) sync() {
	var tools []profile.VMGuestTool
	for i, c := range a.vmToolChecks {
		if c.Checked {
			tools = append(tools, profile.VMGuestTools[i])
		}
	}
	a.profile.InstallVMGuestTools = tools

	if v := a.appLocker.Value(); v != "" {
		a.profile.AppLockerPolicyXML = &v
	} else {
		a.profile.AppLockerPolicyXML = nil
	}

	if v := a.computerNameScript.Value(); v != "" {
		a.profile.ComputerNameScript = &v
	} else {
		a.profile.ComputerNameScript = nil
	}

	a.profile.KeepSensitiveFiles = a.keepSensitiveFiles.Checked
	a.profile.UseNarrator = a.useNarrator.Checked
	a.profile.ObscurePasswords = a.obscurePasswords.Checked
	a.profile.HidePowerShellWindows = a.hidePowerShellWindows.Checked
}

// Update handles focus cycling, checkbox toggling, text input and screen
// navigation.
func (a Advanced) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			a.focus = (a.focus + 1) % advancedFieldCount
			return a, nil
		case "shift+tab":
			a.focus = (a.focus - 1 + advancedFieldCount) % advancedFieldCount
			return a, nil
		case " ":
			switch {
			case a.focus < len(a.vmToolChecks):
				a.vmToolChecks[a.focus].Checked = !a.vmToolChecks[a.focus].Checked
				a.sync()
				return a, nil
			case a.focus == advancedKeepSensitiveFilesIdx:
				a.keepSensitiveFiles.Checked = !a.keepSensitiveFiles.Checked
				a.sync()
				return a, nil
			case a.focus == advancedUseNarratorIdx:
				a.useNarrator.Checked = !a.useNarrator.Checked
				a.sync()
				return a, nil
			case a.focus == advancedObscurePasswordsIdx:
				a.obscurePasswords.Checked = !a.obscurePasswords.Checked
				a.sync()
				return a, nil
			case a.focus == advancedHidePowerShellWindowsIdx:
				a.hidePowerShellWindows.Checked = !a.hidePowerShellWindows.Checked
				a.sync()
				return a, nil
			}
		case "ctrl+n":
			a.sync()
			return a, Navigate(ScreenScripts)
		case "esc":
			a.sync()
			return a, Navigate(ScreenTaskbar)
		case "ctrl+r":
			a.sync()
			return a, Navigate(ScreenReview)
		}
	}

	var cmd tea.Cmd
	switch a.focus {
	case advancedAppLockerIdx:
		a.appLocker, cmd = a.appLocker.Update(msg)
	case advancedComputerNameScriptIdx:
		a.computerNameScript, cmd = a.computerNameScript.Update(msg)
	}
	a.sync()
	return a, cmd
}

// View renders the VM guest tools checkboxes, AppLocker/computer-name text
// areas, and the three small setup-stage checkboxes.
func (a Advanced) View() string {
	out := "Install VM guest tools\n\n"
	for i, c := range a.vmToolChecks {
		out += c.View(a.focus == i) + "\n"
	}
	out += "\n" + a.appLocker.View()
	out += "\n\n" + a.computerNameScript.View()
	out += "\n\n" + a.keepSensitiveFiles.View(a.focus == advancedKeepSensitiveFilesIdx)
	out += "\n" + a.useNarrator.View(a.focus == advancedUseNarratorIdx)
	out += "\n" + a.obscurePasswords.View(a.focus == advancedObscurePasswordsIdx)
	out += "\n" + a.hidePowerShellWindows.View(a.focus == advancedHidePowerShellWindowsIdx)
	out += "\n\n" + a.bar.View()
	return out
}
