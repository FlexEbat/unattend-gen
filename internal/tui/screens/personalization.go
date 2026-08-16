package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/widgets"
)

var colorThemeOptions = []widgets.SelectOption{
	{Value: "", Label: "Windows default"},
	{Value: string(profile.ColorThemeLight), Label: "Light"},
	{Value: string(profile.ColorThemeDark), Label: "Dark"},
}

// Personalization is the colors screen: system/apps theme, accent color and
// where it shows, transparency. Wallpaper and lock screen image need actual
// image bytes and aren't covered yet (see tech notes).
type Personalization struct {
	profile *profile.Profile

	systemTheme widgets.LabeledSelect
	appsTheme   widgets.LabeledSelect
	accentColor widgets.LabeledInput

	showAccentOnStartTaskbar widgets.Checkbox
	showAccentOnTitleBars    widgets.Checkbox
	disableTransparency      widgets.Checkbox

	focus int
	bar   widgets.ConfirmBar
}

const personalizationFieldCount = 6

// NewPersonalization builds the personalization screen backed by profile.
func NewPersonalization(p *profile.Profile) Personalization {
	s := Personalization{
		profile:                  p,
		systemTheme:              widgets.NewLabeledSelect("System theme (taskbar, Start)", colorThemeOptions),
		appsTheme:                widgets.NewLabeledSelect("Apps theme", colorThemeOptions),
		accentColor:              widgets.NewLabeledInput("Accent color (hex RRGGBB, optional)", "0078D4"),
		showAccentOnStartTaskbar: widgets.Checkbox{Label: "Show accent color on Start and taskbar"},
		showAccentOnTitleBars:    widgets.Checkbox{Label: "Show accent color on title bars"},
		disableTransparency:      widgets.Checkbox{Label: "Disable transparency effects"},
		bar:                      widgets.NewConfirmBar("Tab: focus", "Space: toggle", "Ctrl+N: next", "Esc: back", "Ctrl+R: review"),
	}
	s.systemTheme.SetValue(string(p.Personalization.SystemTheme))
	s.appsTheme.SetValue(string(p.Personalization.AppsTheme))
	if p.Personalization.AccentColor != nil {
		s.accentColor.SetValue(*p.Personalization.AccentColor)
	}
	s.showAccentOnStartTaskbar.Checked = p.Personalization.ShowAccentOnStartTaskbar
	s.showAccentOnTitleBars.Checked = p.Personalization.ShowAccentOnTitleBars
	s.disableTransparency.Checked = p.Personalization.DisableTransparency
	return s
}

// Init is a no-op.
func (s Personalization) Init() tea.Cmd {
	return nil
}

func (s *Personalization) sync() {
	s.profile.Personalization = profile.PersonalizationSettings{
		SystemTheme:              profile.ColorTheme(s.systemTheme.Value()),
		AppsTheme:                profile.ColorTheme(s.appsTheme.Value()),
		ShowAccentOnStartTaskbar: s.showAccentOnStartTaskbar.Checked,
		ShowAccentOnTitleBars:    s.showAccentOnTitleBars.Checked,
		DisableTransparency:      s.disableTransparency.Checked,
	}
	if v := s.accentColor.Value(); v != "" {
		s.profile.Personalization.AccentColor = &v
	}
}

// Update handles focus cycling, select/text input, checkbox toggling and
// screen navigation.
func (s Personalization) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			s.focus = (s.focus + 1) % personalizationFieldCount
			s.sync()
			return s, nil
		case "shift+tab":
			s.focus = (s.focus - 1 + personalizationFieldCount) % personalizationFieldCount
			s.sync()
			return s, nil
		case " ":
			switch s.focus {
			case 3:
				s.showAccentOnStartTaskbar.Checked = !s.showAccentOnStartTaskbar.Checked
			case 4:
				s.showAccentOnTitleBars.Checked = !s.showAccentOnTitleBars.Checked
			case 5:
				s.disableTransparency.Checked = !s.disableTransparency.Checked
			}
			s.sync()
			return s, nil
		case "ctrl+n":
			s.sync()
			return s, Navigate(ScreenScripts)
		case "esc":
			s.sync()
			return s, Navigate(ScreenApps)
		case "ctrl+r":
			s.sync()
			return s, Navigate(ScreenReview)
		}
	}

	var cmd tea.Cmd
	switch s.focus {
	case 0:
		s.systemTheme, cmd = s.systemTheme.Update(msg)
	case 1:
		s.appsTheme, cmd = s.appsTheme.Update(msg)
	case 2:
		s.accentColor, cmd = s.accentColor.Update(msg)
	}
	s.sync()
	return s, cmd
}

// View renders the theme selectors, accent color field and checkboxes.
func (s Personalization) View() string {
	out := s.systemTheme.View() + "\n\n" + s.appsTheme.View() + "\n\n" + s.accentColor.View()
	out += "\n\n" + s.showAccentOnStartTaskbar.View(s.focus == 3)
	out += "\n" + s.showAccentOnTitleBars.View(s.focus == 4)
	out += "\n" + s.disableTransparency.View(s.focus == 5)
	out += "\n\n" + s.bar.View()
	return out
}
