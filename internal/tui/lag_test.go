package tui

import (
	"testing"
	"time"

	"github.com/FlexEbat/unattend-gen/internal/tui/screens"
)

// TestTypingDoesNotBlockUpdate guards against Update synchronously executing
// arbitrary commands (e.g. the text cursor's blink command, which sleeps for
// ~half a second), which made typing in the TUI nearly impossible.
func TestTypingDoesNotBlockUpdate(t *testing.T) {
	m := modelAt(screens.ScreenAccounts)
	m.Init()
	for _, r := range "hello" {
		start := time.Now()
		next, _ := m.Update(keyMsg(string(r)))
		m = next.(Model)
		if d := time.Since(start); d > 100*time.Millisecond {
			t.Fatalf("Update took %v for key %q; it must not run commands synchronously", d, string(r))
		}
	}
}
