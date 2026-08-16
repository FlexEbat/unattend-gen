// Package screens holds the individual TUI screens (welcome, language,
// accounts, tweaks, wifi, apps, personalization, scripts, review). Each screen is a bubbletea.Model that
// reads and writes the shared *profile.Profile passed in at construction, so
// switching screens never loses already-entered data.
package screens

import (
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// ID identifies one of the linear screens.
type ID int

const (
	ScreenWelcome ID = iota
	ScreenLanguage
	ScreenAccounts
	ScreenTweaks
	ScreenWifi
	ScreenApps
	ScreenPersonalization
	ScreenScripts
	ScreenReview
)

// NavigateMsg asks the root app model to switch the active screen.
type NavigateMsg struct {
	To ID
}

// Navigate returns a tea.Cmd that requests a screen switch.
func Navigate(to ID) tea.Cmd {
	return func() tea.Msg { return NavigateMsg{To: to} }
}

// ErrorReceiver is implemented by screens that can route ValidateForReview's
// error strings to the specific field widgets they concern, instead of only
// a generic banner. Matching is done on the exact Russian messages
// validate.go produces (they're our own constants, not third-party text, so
// substring matching on them is precise rather than a fragile heuristic).
type ErrorReceiver interface {
	SetErrors(errs []string) tea.Model
}

// ValidateForReview re-validates the profile the same way profile.ValidateProfile
// does for the CLI, so the review screen and `unattend-gen generate` never
// disagree about what is valid. Returns nil when the profile is valid.
func ValidateForReview(p *profile.Profile) []string {
	data, err := json.Marshal(p)
	if err != nil {
		return []string{err.Error()}
	}
	result := profile.ValidateProfile(data)
	return result.Errors
}
