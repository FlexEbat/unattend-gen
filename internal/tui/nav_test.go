package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/screens"
)

// screenNames labels every screens.ID for failure messages. Its length also
// guards the loops below: adding a screen without listing it here fails
// TestScreenNamesCoverEveryID, so a new screen cannot silently skip the
// wiring checks.
var screenNames = []string{
	"Welcome", "Language", "Accounts", "Tweaks", "Wifi", "Apps",
	"Personalization", "Accessibility", "Desktop", "VisualEffects",
	"Taskbar", "Advanced", "Scripts", "Review",
}

func allIDs() []screens.ID {
	ids := make([]screens.ID, 0, len(screenNames))
	for id := screens.ScreenWelcome; id <= screens.ScreenReview; id++ {
		ids = append(ids, id)
	}
	return ids
}

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "ctrl+n":
		return tea.KeyMsg{Type: tea.KeyCtrlN}
	case "ctrl+r":
		return tea.KeyMsg{Type: tea.KeyCtrlR}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func press(t *testing.T, m Model, key string) Model {
	t.Helper()
	next, _ := m.Update(keyMsg(key))
	got, ok := next.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want tui.Model", next)
	}
	return got
}

// modelAt returns a model whose active screen is id, backed by a valid
// default profile.
func modelAt(id screens.ID) Model {
	m := NewModel(profile.Default("demo"))
	m.current = id
	m.screens[id] = rebuildScreen(id, m.profile)
	return m
}

func typeName(v tea.Model) string {
	return fmt.Sprintf("%T", v)
}

// TestScreenNamesCoverEveryID fails when a screens.ID exists beyond the list
// above, i.e. a screen was added after Review or the list went stale.
func TestScreenNamesCoverEveryID(t *testing.T) {
	if got := int(screens.ScreenReview) + 1; got != len(screenNames) {
		t.Fatalf("screens.ID range has %d entries, screenNames has %d; update screenNames", got, len(screenNames))
	}
}

// TestEveryScreenIDIsWired is the regression test for a bug where ScreenAdvanced
// was missing from rebuildScreen and fell through to Review: every
// ID must be registered in NewModel and rebuilt to the same distinct type.
func TestEveryScreenIDIsWired(t *testing.T) {
	p := profile.Default("demo")
	m := NewModel(p)

	seen := map[string]string{}
	for _, id := range allIDs() {
		name := screenNames[id]

		registered, ok := m.screens[id]
		if !ok || registered == nil {
			t.Fatalf("%s: not registered in NewModel", name)
		}
		rebuilt := rebuildScreen(id, p)
		if typeName(rebuilt) != typeName(registered) {
			t.Errorf("%s: rebuildScreen returns %s, NewModel registers %s", name, typeName(rebuilt), typeName(registered))
		}
		if prev, dup := seen[typeName(rebuilt)]; dup {
			t.Errorf("%s and %s share screen type %s; rebuildScreen likely falls through to default", name, prev, typeName(rebuilt))
		}
		seen[typeName(rebuilt)] = name
	}
}

func TestEveryScreenRenders(t *testing.T) {
	for _, id := range allIDs() {
		name := screenNames[id]
		m := modelAt(id)
		if cmd := m.Init(); cmd != nil {
			_ = cmd // Init commands (cursor blink) are not executed here.
		}
		next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		m = next.(Model)
		if strings.TrimSpace(m.View()) == "" {
			t.Errorf("%s: View() is empty", name)
		}
	}
}

// TestNavigationForward checks Ctrl+N walks the whole linear flow in order,
// from Language to Review.
func TestNavigationForward(t *testing.T) {
	for id := screens.ScreenLanguage; id < screens.ScreenReview; id++ {
		m := press(t, modelAt(id), "ctrl+n")
		if want := id + 1; m.current != want {
			t.Errorf("Ctrl+N on %s: got %s, want %s", screenNames[id], screenNames[m.current], screenNames[want])
		}
	}
}

// TestNavigationBack checks Esc walks the flow backwards, including
// Review -> Scripts and Language -> Welcome.
func TestNavigationBack(t *testing.T) {
	for id := screens.ScreenLanguage; id <= screens.ScreenReview; id++ {
		m := press(t, modelAt(id), "esc")
		if want := id - 1; m.current != want {
			t.Errorf("Esc on %s: got %s, want %s", screenNames[id], screenNames[m.current], screenNames[want])
		}
	}
}

// TestCtrlRReachesReviewFromEveryScreen covers the "jump to review" shortcut
// on every screen that offers it (all except Review itself).
func TestCtrlRReachesReviewFromEveryScreen(t *testing.T) {
	for id := screens.ScreenWelcome; id < screens.ScreenReview; id++ {
		m := press(t, modelAt(id), "ctrl+r")
		if m.current != screens.ScreenReview {
			t.Errorf("Ctrl+R on %s: got %s, want Review", screenNames[id], screenNames[m.current])
		}
	}
}

// TestInvalidProfileBlocksReviewFromEveryScreen checks the validation gate:
// an invalid profile keeps the user where they are, whichever screen they
// press Ctrl+R on. Accounts is skipped: its sync() rewrites ComputerName from
// its own text field before validation, so injecting the bad value into the
// profile cannot work there; TestAppInvalidValueBlocksReview covers it by
// typing into the field instead.
func TestInvalidProfileBlocksReviewFromEveryScreen(t *testing.T) {
	for id := screens.ScreenLanguage; id < screens.ScreenReview; id++ {
		if id == screens.ScreenAccounts {
			continue
		}
		m := modelAt(id)
		bad := "this-computer-name-is-too-long"
		m.profile.ComputerName = &bad

		m = press(t, m, "ctrl+r")
		if m.current != id {
			t.Errorf("Ctrl+R on %s with invalid profile: moved to %s, want to stay", screenNames[id], screenNames[m.current])
		}
		if m.err == "" {
			t.Errorf("Ctrl+R on %s with invalid profile: no error recorded", screenNames[id])
		}
	}
}

// TestAppsSelectionSurvivesNavigation covers the screen <-> shared profile
// round trip: toggling a checkbox writes to the profile, and the rebuilt
// screen shows it after navigating away and back.
func TestAppsSelectionSurvivesNavigation(t *testing.T) {
	m := modelAt(screens.ScreenApps)

	m = press(t, m, " ") // first checkbox has focus
	m = press(t, m, "ctrl+n")
	if m.current != screens.ScreenPersonalization {
		t.Fatalf("Ctrl+N: got %s, want Personalization", screenNames[m.current])
	}

	first := profile.RemovableApps[0]
	if len(m.profile.RemoveApps) != 1 || m.profile.RemoveApps[0] != first {
		t.Fatalf("RemoveApps = %v, want [%s]", m.profile.RemoveApps, first)
	}

	m = press(t, m, "esc")
	if m.current != screens.ScreenApps {
		t.Fatalf("Esc: got %s, want Apps", screenNames[m.current])
	}
	if want := "[x] " + string(first); !strings.Contains(m.View(), want) {
		t.Errorf("Apps view does not show %q after returning:\n%s", want, m.View())
	}
}

// TestDesktopCustomizeIconsWritesProfile covers a conditional-field screen:
// enabling "customize icons" turns nil DesktopIcons into a full map, and
// disabling it again resets to nil.
func TestDesktopCustomizeIconsWritesProfile(t *testing.T) {
	m := modelAt(screens.ScreenDesktop)
	if m.profile.DesktopIcons != nil {
		t.Fatalf("precondition: DesktopIcons = %v, want nil", m.profile.DesktopIcons)
	}

	m = press(t, m, " ")
	if got := len(m.profile.DesktopIcons); got != len(profile.DesktopIcons) {
		t.Errorf("after enabling: DesktopIcons has %d entries, want %d", got, len(profile.DesktopIcons))
	}

	m = press(t, m, " ")
	if m.profile.DesktopIcons != nil {
		t.Errorf("after disabling: DesktopIcons = %v, want nil", m.profile.DesktopIcons)
	}
}
