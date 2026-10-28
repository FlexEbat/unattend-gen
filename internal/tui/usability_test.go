package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/screens"
)

func resize(t *testing.T, m Model, w, h int) Model {
	t.Helper()
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next.(Model)
}

func lines(s string) int { return strings.Count(s, "\n") + 1 }

// The XML preview used to be printed in full, so on a normal terminal its top
// was cut off and unreachable.
func TestReviewPreviewFitsAndScrolls(t *testing.T) {
	m := resize(t, modelAt(screens.ScreenReview), 100, 24)
	first := m.View()
	if got := lines(first); got > 24 {
		t.Fatalf("review view is %d lines on a 24-row terminal; it must fit", got)
	}
	if !strings.Contains(first, "<?xml") {
		t.Fatalf("the top of the XML must be visible:\n%s", first)
	}
	for i := 0; i < 3; i++ {
		m = press(t, m, "pgdown")
	}
	if scrolled := m.View(); scrolled == first || strings.Contains(scrolled, "<?xml") {
		t.Fatal("PgDn did not scroll the preview")
	}
	if got := lines(m.View()); got > 24 {
		t.Fatalf("scrolled view is %d lines; it must still fit", got)
	}
}

// Only the active screen receives WindowSizeMsg, so Review must be told the
// size when it is created by navigation.
func TestReviewKnowsTerminalSizeAfterNavigation(t *testing.T) {
	m := resize(t, modelAt(screens.ScreenLanguage), 100, 24)
	m = press(t, m, "ctrl+r")
	if m.current != screens.ScreenReview {
		t.Fatalf("ctrl+r went to screen %d, want Review", m.current)
	}
	if got := lines(m.View()); got > 24 {
		t.Fatalf("review reached by navigation is %d lines on a 24-row terminal", got)
	}
}

func TestExportAsksBeforeOverwriting(t *testing.T) {
	defer chdir(t, t.TempDir())()
	if err := os.WriteFile("autounattend.xml", []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	read := func() string {
		b, err := os.ReadFile("autounattend.xml")
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	m := modelAt(screens.ScreenReview)

	m = press(t, m, "x")
	if read() != "old" || !strings.Contains(m.View(), "already exists") {
		t.Fatalf("first x must only ask; file=%q", read())
	}
	m = press(t, m, "down") // any other key cancels
	m = press(t, m, "x")
	if read() != "old" {
		t.Fatal("a cancelled prompt must not be confirmed by a later x")
	}
	m = press(t, m, "x") // second x in a row confirms
	if read() == "old" {
		t.Fatal("x, x should overwrite the file")
	}
	_ = m
}

func TestWelcomeNameChecksAndEscape(t *testing.T) {
	defer chdir(t, t.TempDir())()
	if err := profile.SaveProfile(profile.Default("demo"), filepath.Join(profile.ProfilesDir, "demo.json")); err != nil {
		t.Fatal(err)
	}
	typeName := func(m Model, name string) Model {
		for _, r := range name {
			m = press(t, m, string(r))
		}
		return press(t, m, "enter")
	}

	m := modelAt(screens.ScreenWelcome)
	m = press(t, m, "enter") // "New profile"

	m = typeName(m, "demo")
	if m.current != screens.ScreenWelcome || !strings.Contains(m.View(), "already exists") {
		t.Fatalf("an existing name must be refused on Welcome:\n%s", m.View())
	}
	m = press(t, m, "esc") // leave the name field
	if strings.Contains(m.View(), "New profile name") {
		t.Fatal("esc should return from the name field to the list")
	}
	m = press(t, m, "enter") // New profile again
	m = typeName(m, "a/b")
	if m.current != screens.ScreenWelcome || !strings.Contains(m.View(), "path separators") {
		t.Fatalf("an invalid name must be refused on Welcome:\n%s", m.View())
	}
}
