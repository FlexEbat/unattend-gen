package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
)

// Apps is the "remove preinstalled apps and Windows features" screen: one
// checkbox per profile.RemovableApp (Appx packages), one per
// profile.RemovableFeature (DISM capabilities), then one per
// profile.RemovableOptionalFeature (legacy optional features, slice 17) —
// three different removal mechanisms, shown together since to the user
// they're all just "things to remove".
type Apps struct {
	profile               *profile.Profile
	appChecks             []widgets.Checkbox
	featureChecks         []widgets.Checkbox
	optionalFeatureChecks []widgets.Checkbox
	focus                 int
	bar                   widgets.ConfirmBar
}

// NewApps builds the apps screen backed by profile.
func NewApps(p *profile.Profile) Apps {
	a := Apps{
		profile:               p,
		appChecks:             make([]widgets.Checkbox, len(profile.RemovableApps)),
		featureChecks:         make([]widgets.Checkbox, len(profile.RemovableFeatures)),
		optionalFeatureChecks: make([]widgets.Checkbox, len(profile.RemovableOptionalFeatures)),
		bar:                   widgets.NewConfirmBar("Tab: focus", "Space: toggle", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
	}
	selectedApps := make(map[profile.RemovableApp]bool, len(p.RemoveApps))
	for _, app := range p.RemoveApps {
		selectedApps[app] = true
	}
	for i, app := range profile.RemovableApps {
		a.appChecks[i] = widgets.Checkbox{Label: string(app), Checked: selectedApps[app]}
	}
	selectedFeatures := make(map[profile.RemovableFeature]bool, len(p.RemoveFeatures))
	for _, f := range p.RemoveFeatures {
		selectedFeatures[f] = true
	}
	for i, f := range profile.RemovableFeatures {
		a.featureChecks[i] = widgets.Checkbox{Label: string(f), Checked: selectedFeatures[f]}
	}
	selectedOptionalFeatures := make(map[profile.RemovableOptionalFeature]bool, len(p.RemoveOptionalFeatures))
	for _, f := range p.RemoveOptionalFeatures {
		selectedOptionalFeatures[f] = true
	}
	for i, f := range profile.RemovableOptionalFeatures {
		a.optionalFeatureChecks[i] = widgets.Checkbox{Label: string(f), Checked: selectedOptionalFeatures[f]}
	}
	return a
}

// Init is a no-op.
func (a Apps) Init() tea.Cmd {
	return nil
}

func (a Apps) fieldCount() int {
	return len(a.appChecks) + len(a.featureChecks) + len(a.optionalFeatureChecks)
}

// checkboxAt returns a pointer to the checkbox at overall focus index i,
// wherever it lives (apps, features, or optional features).
func (a *Apps) checkboxAt(i int) *widgets.Checkbox {
	if i < len(a.appChecks) {
		return &a.appChecks[i]
	}
	i -= len(a.appChecks)
	if i < len(a.featureChecks) {
		return &a.featureChecks[i]
	}
	return &a.optionalFeatureChecks[i-len(a.featureChecks)]
}

func (a *Apps) sync() {
	var apps []profile.RemovableApp
	for i, c := range a.appChecks {
		if c.Checked {
			apps = append(apps, profile.RemovableApps[i])
		}
	}
	a.profile.RemoveApps = apps

	var features []profile.RemovableFeature
	for i, c := range a.featureChecks {
		if c.Checked {
			features = append(features, profile.RemovableFeatures[i])
		}
	}
	a.profile.RemoveFeatures = features

	var optionalFeatures []profile.RemovableOptionalFeature
	for i, c := range a.optionalFeatureChecks {
		if c.Checked {
			optionalFeatures = append(optionalFeatures, profile.RemovableOptionalFeatures[i])
		}
	}
	a.profile.RemoveOptionalFeatures = optionalFeatures
}

// Update handles focus cycling, checkbox toggling and screen navigation.
func (a Apps) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			a.focus = (a.focus + 1) % a.fieldCount()
			return a, nil
		case "shift+tab":
			a.focus = (a.focus - 1 + a.fieldCount()) % a.fieldCount()
			return a, nil
		case " ", "enter":
			c := a.checkboxAt(a.focus)
			c.Checked = !c.Checked
			a.sync()
			return a, nil
		case "ctrl+n":
			a.sync()
			return a, Navigate(ScreenPersonalization)
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

// View renders every app, feature and optional-feature checkbox.
func (a Apps) View() string {
	out := "Remove preinstalled apps\n\n"
	for i, c := range a.appChecks {
		out += c.View(a.focus == i) + "\n"
	}
	out += "\nRemove Windows features\n\n"
	for i, c := range a.featureChecks {
		out += c.View(a.focus == len(a.appChecks)+i) + "\n"
	}
	out += "\nRemove legacy optional features\n\n"
	for i, c := range a.optionalFeatureChecks {
		out += c.View(a.focus == len(a.appChecks)+len(a.featureChecks)+i) + "\n"
	}
	out += "\n" + a.bar.View()
	return out
}
