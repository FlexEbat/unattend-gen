// Package tui wires the individual screens into a single bubbletea
// application. State lives in one *profile.Profile shared by every screen,
// so switching screens (including jumping straight to review) never loses
// already-entered data.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/i18n"
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
	size    tea.WindowSizeMsg
	hasSize bool
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
			screens.ScreenVisualEffects:   screens.NewVisualEffects(initial),
			screens.ScreenTaskbar:         screens.NewTaskbar(initial),
			screens.ScreenAdvanced:        screens.NewAdvanced(initial),
			screens.ScreenScripts:         screens.NewScripts(initial),
			screens.ScreenReview:          screens.NewReview(initial),
		},
	}
}

// Init starts the initial screen.
func (m Model) Init() tea.Cmd {
	return m.screens[m.current].Init()
}

// Update dispatches messages to the active screen. Screen switches arrive as
// screens.NavigateMsg through the normal bubbletea message loop; commands
// returned by screens (cursor blink, ...) are handed back to bubbletea to run
// asynchronously and are never executed here, because some of them block
// (the text cursor's blink command sleeps ~500ms) and would freeze typing.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		// Remember it: only the active screen receives it below, so a screen
		// shown later (Review) is told the size when it is created.
		m.size, m.hasSize = msg, true
	case screens.NavigateMsg:
		return m.navigate(msg)
	}

	updated, cmd := m.screens[m.current].Update(msg)
	m.screens[m.current] = updated
	return m, cmd
}

// navigate switches the active screen. A switch to Review is gated by
// profile.ValidateProfile: an invalid profile stays on the current screen
// instead.
func (m Model) navigate(nav screens.NavigateMsg) (tea.Model, tea.Cmd) {
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
	if m.hasSize {
		m.screens[m.current], _ = m.screens[m.current].Update(m.size)
	}
	return m, m.screens[m.current].Init()
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
	case screens.ScreenVisualEffects:
		return screens.NewVisualEffects(p)
	case screens.ScreenTaskbar:
		return screens.NewTaskbar(p)
	case screens.ScreenAdvanced:
		return screens.NewAdvanced(p)
	case screens.ScreenScripts:
		return screens.NewScripts(p)
	default:
		return screens.NewReview(p)
	}
}

// View renders the active screen, plus a validation banner when the last
// attempt to reach Review failed.
func (m Model) View() string {
	out := m.screens[m.current].View()
	if m.err != "" {
		out += "\n\n" + i18n.Tf("Cannot open review yet: %s", i18n.Display(m.err))
	}
	return out
}
