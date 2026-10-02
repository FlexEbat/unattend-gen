package tui

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/tui/screens"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// taskbarIconsModeFocus is the Tab count from the first field of the Taskbar
// screen to the taskbar-icons mode select when no custom text area is open:
// 5 checkboxes, search select, Start pins mode, Start tiles mode.
const taskbarIconsModeFocus = 8

func pressN(t *testing.T, m Model, key string, n int) Model {
	t.Helper()
	for i := 0; i < n; i++ {
		m = press(t, m, key)
	}
	return m
}

// TestTaskbarIconsCustomWritesProfile covers the taskbar icons fields end to end:
// choosing "custom" reveals the XML text area and writes both fields to the
// shared profile; switching back to "default" clears the XML again.
func TestTaskbarIconsCustomWritesProfile(t *testing.T) {
	m := modelAt(screens.ScreenTaskbar)
	if strings.Contains(m.View(), "Taskbar layout XML") {
		t.Fatal("XML text area must be hidden until mode=custom")
	}

	m = pressN(t, m, "tab", taskbarIconsModeFocus)
	m = pressN(t, m, "down", 2) // default -> empty -> custom
	if m.profile.TaskbarIcons.Mode != profile.TaskbarIconsModeCustom {
		t.Fatalf("Mode = %q, want custom", m.profile.TaskbarIcons.Mode)
	}
	if m.profile.TaskbarIcons.XML == nil {
		t.Fatal("XML must be non-nil (empty) while mode=custom")
	}
	if !strings.Contains(m.View(), "Taskbar layout XML") {
		t.Fatal("XML text area must be visible for mode=custom")
	}

	m = press(t, m, "tab") // focus the XML text area
	m = press(t, m, "<a/>")
	if got := m.profile.TaskbarIcons.XML; got == nil || *got != "<a/>" {
		t.Fatalf("XML = %v, want <a/>", got)
	}

	m = press(t, m, "shift+tab") // back to the mode select
	m = pressN(t, m, "up", 2)    // custom -> empty -> default
	if m.profile.TaskbarIcons.Mode != profile.TaskbarIconsModeDefault {
		t.Fatalf("Mode = %q, want default", m.profile.TaskbarIcons.Mode)
	}
	if m.profile.TaskbarIcons.XML != nil {
		t.Fatalf("XML = %q, want nil after leaving custom", *m.profile.TaskbarIcons.XML)
	}
}

// TestTaskbarIconsSurvivesNavigation checks the profile -> screen direction:
// values set on the profile are shown again when the screen is rebuilt.
func TestTaskbarIconsSurvivesNavigation(t *testing.T) {
	m := modelAt(screens.ScreenTaskbar)
	xml := "<LayoutModificationTemplate/>"
	m.profile.TaskbarIcons = profile.TaskbarIconsSettings{Mode: profile.TaskbarIconsModeCustom, XML: &xml}
	m.screens[screens.ScreenTaskbar] = rebuildScreen(screens.ScreenTaskbar, m.profile)

	m = press(t, m, "ctrl+n")
	m = press(t, m, "esc")
	if m.current != screens.ScreenTaskbar {
		t.Fatalf("current = %d, want Taskbar", m.current)
	}
	if !strings.Contains(m.View(), "<LayoutModificationTemplate/>") {
		t.Errorf("rebuilt Taskbar screen does not show the stored XML:\n%s", m.View())
	}
	if got := m.profile.TaskbarIcons; got.Mode != profile.TaskbarIconsModeCustom || got.XML == nil || *got.XML != xml {
		t.Errorf("profile TaskbarIcons = %+v, want it unchanged", got)
	}
}

// TestTaskbarCustomTextAreasAcceptInput is a regression test: the Start pins
// and Start tiles text areas were never focused, so typing into them was
// silently dropped.
func TestTaskbarCustomTextAreasAcceptInput(t *testing.T) {
	m := modelAt(screens.ScreenTaskbar)

	m = pressN(t, m, "tab", 6) // Start pins mode
	m = pressN(t, m, "down", 2)
	m = press(t, m, "tab") // pins JSON
	m = press(t, m, "{}")
	if got := m.profile.StartPins.JSON; got == nil || *got != "{}" {
		t.Errorf("StartPins.JSON = %v, want {}", got)
	}

	m = press(t, m, "tab") // Start tiles mode
	m = pressN(t, m, "down", 2)
	m = press(t, m, "tab") // tiles XML
	m = press(t, m, "<t/>")
	if got := m.profile.StartTiles.XML; got == nil || *got != "<t/>" {
		t.Errorf("StartTiles.XML = %v, want <t/>", got)
	}
}
