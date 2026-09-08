package screens

import (
	"strings"

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
// all, and then either the manual fields (SSID/authentication/password/
// visibility) or, if rawMode is checked, a raw WLAN profile XML pasted in
// directly (slice 22) — exported via `netsh wlan export profile
// key=clear`. rawMode replaces the manual fields entirely rather than
// combining with them, matching Profile.WifiSettings.RawProfileXML's
// contract (set = used verbatim, manual fields ignored).
type Wifi struct {
	profile  *profile.Profile
	enabled  widgets.Checkbox
	rawMode  widgets.Checkbox
	rawXML   widgets.LabeledTextArea
	ssid     widgets.LabeledInput
	auth     widgets.LabeledSelect
	password widgets.PasswordInput
	hidden   widgets.Checkbox
	focus    int
	bar      widgets.ConfirmBar
}

// NewWifi builds the wifi screen backed by profile.
func NewWifi(p *profile.Profile) Wifi {
	w := Wifi{
		profile:  p,
		enabled:  widgets.Checkbox{Label: "Configure Wi-Fi"},
		rawMode:  widgets.Checkbox{Label: "Use a raw exported WLAN profile XML instead"},
		rawXML:   widgets.NewLabeledTextArea("WLAN profile XML (netsh wlan export profile key=clear)", ""),
		ssid:     widgets.NewLabeledInput("SSID", "MyNetwork"),
		auth:     widgets.NewLabeledSelect("Authentication", wifiAuthOptions),
		password: widgets.NewPasswordInput("Password (min 8 chars, not used for Open)"),
		hidden:   widgets.Checkbox{Label: "Hidden network"},
		bar:      widgets.NewConfirmBar("Tab: focus", "Space: toggle", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
	}
	if p.Wifi != nil {
		w.enabled.Checked = true
		if p.Wifi.RawProfileXML != nil {
			w.rawMode.Checked = true
			w.rawXML.SetValue(*p.Wifi.RawProfileXML)
		} else {
			w.ssid.SetValue(p.Wifi.SSID)
			w.auth.SetValue(string(p.Wifi.Authentication))
			if p.Wifi.Password != nil {
				w.password.SetValue(*p.Wifi.Password)
			}
			w.hidden.Checked = p.Wifi.ConnectHidden
		}
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

// fieldCount: 0=enabled; if !enabled, that's it. If enabled: 1=rawMode;
// if rawMode, 2=rawXML. Otherwise 2=SSID, 3=auth, 4=password, 5=hidden.
func (w Wifi) fieldCount() int {
	if !w.enabled.Checked {
		return 1
	}
	if w.rawMode.Checked {
		return 3
	}
	return 6
}

func (w *Wifi) sync() {
	if !w.enabled.Checked {
		w.profile.Wifi = nil
		return
	}
	if w.rawMode.Checked {
		raw := w.rawXML.Value()
		w.profile.Wifi = &profile.WifiSettings{RawProfileXML: &raw}
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
	fieldCount := w.fieldCount()
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			w.focus = (w.focus + 1) % fieldCount
			w.sync()
			return w, nil
		case "shift+tab":
			w.focus = (w.focus - 1 + fieldCount) % fieldCount
			w.sync()
			return w, nil
		case " ":
			switch {
			case w.focus == 0:
				w.enabled.Checked = !w.enabled.Checked
			case w.focus == 1 && w.enabled.Checked:
				w.rawMode.Checked = !w.rawMode.Checked
			case w.focus == 5 && w.enabled.Checked && !w.rawMode.Checked:
				w.hidden.Checked = !w.hidden.Checked
			}
			w.sync()
			if w.focus >= w.fieldCount() {
				w.focus = w.fieldCount() - 1
			}
			return w, nil
		case "ctrl+n":
			w.sync()
			return w, Navigate(ScreenApps)
		case "esc":
			w.sync()
			return w, Navigate(ScreenTweaks)
		case "ctrl+r":
			w.sync()
			return w, Navigate(ScreenReview)
		}
	}

	var cmd tea.Cmd
	if w.enabled.Checked {
		switch {
		case w.rawMode.Checked && w.focus == 2:
			w.rawXML, cmd = w.rawXML.Update(msg)
		case !w.rawMode.Checked && w.focus == 2:
			w.ssid, cmd = w.ssid.Update(msg)
		case !w.rawMode.Checked && w.focus == 3:
			w.auth, cmd = w.auth.Update(msg)
		case !w.rawMode.Checked && w.focus == 4:
			w.password, cmd = w.password.Update(msg)
		}
	}
	w.sync()
	return w, cmd
}

// View renders the checkbox and, once enabled, either the raw XML text
// area or the manual network fields.
func (w Wifi) View() string {
	out := w.enabled.View(w.focus == 0)
	if w.enabled.Checked {
		out += "\n\n" + w.rawMode.View(w.focus == 1)
		if w.rawMode.Checked {
			out += "\n\n" + w.rawXML.View()
		} else {
			out += "\n\n" + w.ssid.View() + "\n\n" + w.auth.View() + "\n\n" + w.password.View() + "\n\n" + w.hidden.View(w.focus == 5)
		}
	}
	out += "\n\n" + w.bar.View()
	return out
}

// SetErrors routes validate.go's messages to the field each one names.
// Implements screens.ErrorReceiver.
func (w Wifi) SetErrors(errs []string) tea.Model {
	w.ssid.Err = ""
	w.password.Err = ""
	for _, e := range errs {
		switch {
		case strings.Contains(e, "SSID"):
			w.ssid.Err = e
		case strings.Contains(e, "Пароль Wi-Fi"):
			w.password.Err = e
		}
	}
	return w
}
