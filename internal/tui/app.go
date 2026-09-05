// Package tui wires the individual screens into a single bubbletea
// application. State lives in one *profile.Profile shared by every screen,
// so switching screens (including jumping straight to review) never loses
// already-entered data.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/screens"
)

// Model is the root bubbletea model: it holds the shared profile, the set of
// screens and which one is currently active.
type Model struct {
	profile *profile.Profile
	current screens.ID
	screens map[screens.ID]tea.Model
	err     string
}

// NewModel builds the root model. If initial is nil, a fresh profile with
// the same defaults as `profile init` (no preset) is used; its Name is set
// later, when the welcome screen collects one.
func NewModel(initial *profile.Profile) Model {
	if initial == nil {
		initial = profile.Default("")
	}

	return Model{
		profile: initial,
		current: screens.ScreenWelcome,
		screens: map[screens.ID]tea.Model{
			screens.ScreenWelcome:         screens.NewWelcome(initial),
			screens.ScreenLanguage:        screens.NewLanguage(initial),
			screens.ScreenAccounts:        screens.NewAccounts(initial),
			screens.ScreenTweaks:          screens.NewTweaks(initial),
			screens.ScreenWifi:            screens.NewWifi(initial),
			screens.ScreenApps:            screens.NewApps(initial),
			screens.ScreenPersonalization: screens.NewPersonalization(initial),
			screens.ScreenAccessibility:   screens.NewAccessibility(initial),
			screens.ScreenDesktop:         screens.NewDesktop(initial),
			screens.ScreenScripts:         screens.NewScripts(initial),
			screens.ScreenReview:          screens.NewReview(initial),
		},
	}
}

// Init starts the initial screen.
func (m Model) Init() tea.Cmd {
	return m.screens[m.current].Init()
}

// Update dispatches messages to the active screen and handles screen
// switches. A switch to Review is gated by profile.ValidateProfile: an
// invalid profile stays on the current screen instead.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	active := m.screens[m.current]
	updated, cmd := active.Update(msg)
	m.screens[m.current] = updated

	if nav, ok := extractNavigate(cmd); ok {
		if nav.To == screens.ScreenReview {
			if errs := screens.ValidateForReview(m.profile); len(errs) > 0 {
				m.err = errs[0]
				if er, ok := m.screens[m.current].(screens.ErrorReceiver); ok {
					m.screens[m.current] = er.SetErrors(errs)
				}
				m.screens[screens.ScreenReview] = screens.NewReview(m.profile)
				return m, nil
			}
		}
		m.err = ""
		m.current = nav.To
		m.screens[m.current] = rebuildScreen(m.current, m.profile)
		return m, m.screens[m.current].Init()
	}

	return m, cmd
}

// rebuildScreen reconstructs a screen from the current profile so it always
// reflects the latest data, including changes made on other screens.
func rebuildScreen(id screens.ID, p *profile.Profile) tea.Model {
	switch id {
	case screens.ScreenWelcome:
		return screens.NewWelcome(p)
	case screens.ScreenLanguage:
		return screens.NewLanguage(p)
	case screens.ScreenAccounts:
		return screens.NewAccounts(p)
	case screens.ScreenTweaks:
		return screens.NewTweaks(p)
	case screens.ScreenWifi:
		return screens.NewWifi(p)
	case screens.ScreenApps:
		return screens.NewApps(p)
	case screens.ScreenPersonalization:
		return screens.NewPersonalization(p)
	case screens.ScreenAccessibility:
		return screens.NewAccessibility(p)
	case screens.ScreenDesktop:
		return screens.NewDesktop(p)
	case screens.ScreenScripts:
		return screens.NewScripts(p)
	default:
		return screens.NewReview(p)
	}
}

// extractNavigate runs cmd, if any, and reports whether it produced a
// screens.NavigateMsg.
func extractNavigate(cmd tea.Cmd) (screens.NavigateMsg, bool) {
	if cmd == nil {
		return screens.NavigateMsg{}, false
	}
	msg := cmd()
	nav, ok := msg.(screens.NavigateMsg)
	return nav, ok
}

// View renders the active screen, plus a validation banner when the last
// attempt to reach Review failed.
func (m Model) View() string {
	out := m.screens[m.current].View()
	if m.err != "" {
		out += "\n\nCannot open review yet: " + m.err
	}
	return out
}
