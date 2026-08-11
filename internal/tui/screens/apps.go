package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
)

// Apps is the "remove preinstalled apps" screen: one checkbox per
// profile.RemovableApp, in profile.RemovableApps order.
type Apps struct {
	profile *profile.Profile
	checks  []widgets.Checkbox
	focus   int
	bar     widgets.ConfirmBar
}

// NewApps builds the apps screen backed by profile.
func NewApps(p *profile.Profile) Apps {
	a := Apps{
		profile: p,
		checks:  make([]widgets.Checkbox, len(profile.RemovableApps)),
		bar:     widgets.NewConfirmBar("Tab: focus", "Space: toggle", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
	}
	selected := make(map[profile.RemovableApp]bool, len(p.RemoveApps))
	for _, app := range p.RemoveApps {
		selected[app] = true
	}
	for i, app := range profile.RemovableApps {
		a.checks[i] = widgets.Checkbox{Label: string(app), Checked: selected[app]}
	}
	return a
}

// Init is a no-op.
func (a Apps) Init() tea.Cmd {
	return nil
}

func (a *Apps) sync() {
	var apps []profile.RemovableApp
	for i, c := range a.checks {
		if c.Checked {
			apps = append(apps, profile.RemovableApps[i])
		}
	}
	a.profile.RemoveApps = apps
}

// Update handles focus cycling, checkbox toggling and screen navigation.
func (a Apps) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			a.focus = (a.focus + 1) % len(a.checks)
			return a, nil
		case "shift+tab":
			a.focus = (a.focus - 1 + len(a.checks)) % len(a.checks)
			return a, nil
		case " ", "enter":
			a.checks[a.focus].Checked = !a.checks[a.focus].Checked
			a.sync()
			return a, nil
		case "ctrl+n":
			a.sync()
			return a, Navigate(ScreenScripts)
		case "esc":
			a.sync()
			return a, Navigate(ScreenWifi)
		case "ctrl+r":
			a.sync()
			return a, Navigate(ScreenReview)
		}
	}
	return a, nil
}

// View renders every app checkbox.
func (a Apps) View() string {
	out := "Remove preinstalled apps\n\n"
	for i, c := range a.checks {
		out += c.View(a.focus == i) + "\n"
	}
	out += "\n" + a.bar.View()
	return out
}
