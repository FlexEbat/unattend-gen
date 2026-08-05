package profile

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSaveLoadProfileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "demo.json")

	edition := EditionModeInteractive
	original := &Profile{
		SchemaVersion: 1,
		Name:          "demo",
		Language: LanguageSettings{
			UILanguage:     "en-US",
			Locale:         "en-US",
			KeyboardLayout: "en-US",
		},
		Edition: EditionSettings{
			Mode: edition,
		},
		Accounts: []UserAccount{},
		FirstLogon: FirstLogon{
			Mode: FirstLogonNone,
		},
		ExpressSettings: ExpressSettings{
			Mode: ExpressInteractive,
		},
	}

	if err := SaveProfile(original, path); err != nil {
		t.Fatalf("SaveProfile: %v", err)
	}

	loaded, err := LoadProfile(path)
	if err != nil {
		t.Fatalf("LoadProfile: %v", err)
	}

	if !reflect.DeepEqual(original, loaded) {
		t.Fatalf("round trip mismatch:\noriginal: %+v\nloaded:   %+v", original, loaded)
	}
}

func TestLoadProfileMissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadProfile(filepath.Join(dir, "missing.json"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestListProfilesSortedJSONOnly(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"b.json", "a.json", "notes.txt"} {
		p := &Profile{SchemaVersion: 1, Name: name}
		if err := SaveProfile(p, filepath.Join(dir, name)); err != nil {
			t.Fatalf("SaveProfile(%s): %v", name, err)
		}
	}

	paths, err := ListProfiles(dir)
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}

	want := []string{filepath.Join(dir, "a.json"), filepath.Join(dir, "b.json")}
	if !reflect.DeepEqual(want, paths) {
		t.Fatalf("ListProfiles = %v, want %v", paths, want)
	}
}

// TestPresetMinimalMatchesDefaultsWithReplacedName covers slice 5 criterion 3:
// `profile init demo --preset minimal` must produce a profile identical to
// presets/minimal.json except for the replaced name.
func TestPresetMinimalMatchesDefaultsWithReplacedName(t *testing.T) {
	preset, err := LoadProfile(filepath.Join("..", "..", "presets", "minimal.json"))
	if err != nil {
		t.Fatalf("LoadProfile(presets/minimal.json): %v", err)
	}
	preset.Name = "demo"

	want := &Profile{
		SchemaVersion: 1,
		Name:          "demo",
		Language: LanguageSettings{
			UILanguage:     "en-US",
			Locale:         "en-US",
			KeyboardLayout: "en-US",
		},
		Edition:         EditionSettings{Mode: EditionModeInteractive},
		Accounts:        []UserAccount{},
		FirstLogon:      FirstLogon{Mode: FirstLogonNone},
		ExpressSettings: ExpressSettings{Mode: ExpressInteractive},
	}
	if !reflect.DeepEqual(want, preset) {
		t.Fatalf("preset minimal with replaced name = %+v, want %+v", preset, want)
	}

	data, err := json.Marshal(preset)
	if err != nil {
		t.Fatalf("marshal preset: %v", err)
	}
	if result := ValidateProfile(data); len(result.Errors) != 0 {
		t.Fatalf("preset minimal should validate, got errors: %v", result.Errors)
	}
}

// TestPresetSingleUserValidates covers the other built-in preset: one local
// administrator, telemetry disabled, auto-logon into that account.
func TestPresetSingleUserValidates(t *testing.T) {
	preset, err := LoadProfile(filepath.Join("..", "..", "presets", "single-user.json"))
	if err != nil {
		t.Fatalf("LoadProfile(presets/single-user.json): %v", err)
	}

	if len(preset.Accounts) != 1 || preset.Accounts[0].Group != GroupAdministrators {
		t.Fatalf("expected exactly one Administrators account, got %+v", preset.Accounts)
	}
	if preset.ExpressSettings.Mode != ExpressAllDisabled {
		t.Fatalf("expected telemetry disabled, got mode %q", preset.ExpressSettings.Mode)
	}
	if preset.FirstLogon.Mode != FirstLogonFirstCreatedAccount {
		t.Fatalf("expected auto-logon into the first created account, got mode %q", preset.FirstLogon.Mode)
	}

	data, err := json.Marshal(preset)
	if err != nil {
		t.Fatalf("marshal preset: %v", err)
	}
	if result := ValidateProfile(data); len(result.Errors) != 0 {
		t.Fatalf("preset single-user should validate, got errors: %v", result.Errors)
	}
}
