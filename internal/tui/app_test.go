package tui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/tui/screens"
	"github.com/FlexEbat/unattend-gen/internal/xmlgen"
)

func chdir(t *testing.T, dir string) func() {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	return func() {
		_ = os.Chdir(old)
	}
}

// TestAppReviewMatchesGenerate covers criterion 1: filling the screens (here,
// accepting every default) and exporting from review produces the same XML
// as BuildAnswerFile on the equivalent profile, which is exactly what
// `unattend-gen generate` calls.
func TestAppReviewMatchesGenerate(t *testing.T) {
	dir := t.TempDir()
	defer chdir(t, dir)()

	tm := teatest.NewTestModel(t, NewModel(nil), teatest.WithInitialTermSize(120, 40))

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // welcome: pick "New profile"
	tm.Type("demo")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // submit name -> language screen

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlR})                     // jump straight to review; defaults are valid
	tm.Send(tea.KeyMsg{Runes: []rune("x"), Type: tea.KeyRunes}) // export autounattend.xml

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))

	got, err := os.ReadFile(filepath.Join(dir, "autounattend.xml"))
	if err != nil {
		t.Fatalf("expected autounattend.xml to be written: %v", err)
	}

	want, err := xmlgen.BuildAnswerFile(equivalentDefaultProfile("demo"))
	if err != nil {
		t.Fatalf("BuildAnswerFile: %v", err)
	}
	if string(got) != want {
		t.Fatalf("TUI export does not match BuildAnswerFile:\nTUI:\n%s\nwant:\n%s", got, want)
	}
}

// TestAppInvalidValueBlocksReview covers criterion 4: an invalid field value
// (a computer name longer than 15 characters) keeps the user on the current
// screen instead of opening review, and records an error.
func TestAppInvalidValueBlocksReview(t *testing.T) {
	dir := t.TempDir()
	defer chdir(t, dir)()

	tm := teatest.NewTestModel(t, NewModel(nil), teatest.WithInitialTermSize(120, 40))

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // welcome: pick "New profile"
	tm.Type("demo")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // submit name -> language screen

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlN}) // language -> accounts (defaults valid)
	tm.Type("this-computer-name-is-too-long")
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlR}) // attempt to jump to review

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))

	final, ok := tm.FinalModel(t).(Model)
	if !ok {
		t.Fatalf("final model is not tui.Model")
	}
	if final.current == screens.ScreenReview {
		t.Fatal("expected navigation to review to be blocked by validation")
	}
	if final.err == "" {
		t.Fatal("expected a validation error to be recorded")
	}
}

func equivalentDefaultProfile(name string) *profile.Profile {
	return &profile.Profile{
		SchemaVersion: 1,
		Name:          name,
		Language: profile.LanguageSettings{
			UILanguage:     "en-US",
			Locale:         "en-US",
			KeyboardLayout: "en-US",
		},
		Edition:         profile.EditionSettings{Mode: profile.EditionModeInteractive},
		Accounts:        []profile.UserAccount{},
		FirstLogon:      profile.FirstLogon{Mode: profile.FirstLogonNone},
		ExpressSettings: profile.ExpressSettings{Mode: profile.ExpressInteractive},
	}
}
