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

// Advanced is the VM guest tools / AppLocker screen: a checkbox per
// profile.VMGuestTool (each silently installs from its guest-tools ISO if
// attached, no-ops otherwise) plus a raw AppLocker policy XML text area
// (empty = skip, matching Profile.AppLockerPolicyXML's nil-means-skip
// contract). Slice 21 (tech.md backlog group C).
type Advanced struct {
	profile *profile.Profile

	vmToolChecks []widgets.Checkbox
	appLocker    widgets.LabeledTextArea

	focus int
	bar   widgets.ConfirmBar
}

var advancedFieldCount = len(profile.VMGuestTools) + 1 // + AppLocker text area

// NewAdvanced builds the advanced screen backed by profile.
func NewAdvanced(p *profile.Profile) Advanced {
	a := Advanced{
		profile:      p,
		vmToolChecks: make([]widgets.Checkbox, len(profile.VMGuestTools)),
		appLocker:    widgets.NewLabeledTextArea("AppLocker policy XML (optional, raw XML)", ""),
		bar:          widgets.NewConfirmBar("Tab: focus", "Space: toggle", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
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
	return a
}

// Init is a no-op.
func (a Advanced) Init() tea.Cmd {
	return nil
}

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
			if a.focus < len(a.vmToolChecks) {
				a.vmToolChecks[a.focus].Checked = !a.vmToolChecks[a.focus].Checked
				a.sync()
				return a, nil
			}
		case "ctrl+n":
			a.sync()
			return a, Navigate(ScreenScripts)
		case "esc":
			a.sync()
			return a, Navigate(ScreenDesktop)
		case "ctrl+r":
			a.sync()
			return a, Navigate(ScreenReview)
		}
	}

	var cmd tea.Cmd
	if a.focus == len(a.vmToolChecks) {
		a.appLocker, cmd = a.appLocker.Update(msg)
	}
	a.sync()
	return a, cmd
}

// View renders the VM guest tools checkboxes and AppLocker text area.
func (a Advanced) View() string {
	out := "Install VM guest tools\n\n"
	for i, c := range a.vmToolChecks {
		out += c.View(a.focus == i) + "\n"
	}
	out += "\n" + a.appLocker.View()
	out += "\n\n" + a.bar.View()
	return out
}
