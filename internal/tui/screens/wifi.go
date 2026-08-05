package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
)

// Wifi is a stub screen for slice 5: it only toggles whether a Wi-Fi profile
// is configured at all (profile.Wifi nil vs non-nil). The SSID/authentication/
// password fields and their validation are added in slice 6.
type Wifi struct {
	profile *profile.Profile
	enabled checkbox
	bar     widgets.ConfirmBar
}

// NewWifi builds the wifi screen backed by profile.
func NewWifi(p *profile.Profile) Wifi {
	return Wifi{
		profile: p,
		enabled: checkbox{Label: "Configure Wi-Fi (fields added in a later slice)", Checked: p.Wifi != nil},
		bar:     widgets.NewConfirmBar("Space: toggle", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
	}
}

// Init is a no-op.
func (w Wifi) Init() tea.Cmd {
	return nil
}

func (w *Wifi) sync() {
	if w.enabled.Checked {
		if w.profile.Wifi == nil {
			w.profile.Wifi = &profile.WifiSettings{Authentication: profile.WifiOpen}
		}
	} else {
		w.profile.Wifi = nil
	}
}

// Update toggles the checkbox and handles screen navigation.
func (w Wifi) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case " ", "enter":
			w.enabled.Checked = !w.enabled.Checked
			w.sync()
			return w, nil
		case "ctrl+n":
			w.sync()
			return w, Navigate(ScreenReview)
		case "esc":
			w.sync()
			return w, Navigate(ScreenTweaks)
		case "ctrl+r":
			w.sync()
			return w, Navigate(ScreenReview)
		}
	}
	return w, nil
}

// View renders the single checkbox.
func (w Wifi) View() string {
	return w.enabled.View(true) + "\n\n" + w.bar.View()
}
