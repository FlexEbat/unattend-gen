package screens

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
)

var wifiAuthOptions = []widgets.SelectOption{
	{Value: string(profile.WifiOpen), Label: "Open (no password)"},
	{Value: string(profile.WifiWPA2Personal), Label: "WPA2 Personal"},
	{Value: string(profile.WifiWPA3Personal), Label: "WPA3 Personal"},
}

// Wifi is the Wi-Fi configuration screen: whether to configure a network at
// all, and if so its SSID, authentication, password and visibility.
type Wifi struct {
	profile  *profile.Profile
	enabled  checkbox
	ssid     widgets.LabeledInput
	auth     widgets.LabeledSelect
	password widgets.PasswordInput
	hidden   checkbox
	focus    int
	bar      widgets.ConfirmBar
}

// NewWifi builds the wifi screen backed by profile.
func NewWifi(p *profile.Profile) Wifi {
	w := Wifi{
		profile:  p,
		enabled:  checkbox{Label: "Configure Wi-Fi"},
		ssid:     widgets.NewLabeledInput("SSID", "MyNetwork"),
		auth:     widgets.NewLabeledSelect("Authentication", wifiAuthOptions),
		password: widgets.NewPasswordInput("Password (min 8 chars, not used for Open)"),
		hidden:   checkbox{Label: "Hidden network"},
		bar:      widgets.NewConfirmBar("Tab: focus", "Space: toggle", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
	}
	if p.Wifi != nil {
		w.enabled.Checked = true
		w.ssid.SetValue(p.Wifi.SSID)
		w.auth.SetValue(string(p.Wifi.Authentication))
		if p.Wifi.Password != nil {
			w.password.SetValue(*p.Wifi.Password)
		}
		w.hidden.Checked = p.Wifi.ConnectHidden
	} else {
		w.auth.SetValue(string(profile.WifiOpen))
	}
	return w
}

// Init returns the cursor-blink command; the focus state itself was already
// set (or not needed) at construction time.
func (w Wifi) Init() tea.Cmd {
	return textinput.Blink
}

// fieldCount: 0=enabled checkbox, 1=SSID, 2=auth, 3=password, 4=hidden.
const wifiFieldCount = 5

func (w *Wifi) sync() {
	if !w.enabled.Checked {
		w.profile.Wifi = nil
		return
	}
	var password *string
	if v := w.password.Value(); v != "" {
		password = &v
	}
	w.profile.Wifi = &profile.WifiSettings{
		SSID:           w.ssid.Value(),
		Authentication: profile.WifiAuthentication(w.auth.Value()),
		Password:       password,
		ConnectHidden:  w.hidden.Checked,
	}
}

// Update handles focus cycling, checkbox/select input and screen navigation.
func (w Wifi) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			w.focus = (w.focus + 1) % wifiFieldCount
			w.sync()
			return w, nil
		case "shift+tab":
			w.focus = (w.focus - 1 + wifiFieldCount) % wifiFieldCount
			w.sync()
			return w, nil
		case " ":
			switch w.focus {
			case 0:
				w.enabled.Checked = !w.enabled.Checked
			case 4:
				w.hidden.Checked = !w.hidden.Checked
			}
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

	var cmd tea.Cmd
	switch w.focus {
	case 1:
		w.ssid, cmd = w.ssid.Update(msg)
	case 2:
		w.auth, cmd = w.auth.Update(msg)
	case 3:
		w.password, cmd = w.password.Update(msg)
	}
	w.sync()
	return w, cmd
}

// View renders the checkbox and, once enabled, the network fields.
func (w Wifi) View() string {
	out := w.enabled.View(w.focus == 0)
	if w.enabled.Checked {
		out += "\n\n" + w.ssid.View() + "\n\n" + w.auth.View() + "\n\n" + w.password.View() + "\n\n" + w.hidden.View(w.focus == 4)
	}
	out += "\n\n" + w.bar.View()
	return out
}
