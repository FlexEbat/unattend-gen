package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
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

// TestPresetMinimalMatchesDefaultsWithReplacedName verifies that
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

func TestDecodeProfileRejectsUnknownFieldsAndTrailingData(t *testing.T) {
	if _, err := DecodeProfile([]byte(`{"schema_version":1,"nam":"x"}`)); err == nil {
		t.Fatal("expected an error for an unknown field")
	}
	if _, err := DecodeProfile([]byte(`{"schema_version":1} {"schema_version":1}`)); err == nil {
		t.Fatal("expected an error for trailing data")
	}
	if _, err := DecodeProfile([]byte(`{"schema_version":1}` + "\n")); err != nil {
		t.Fatalf("trailing whitespace must be accepted: %v", err)
	}
}

func TestLoadProfileRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "typo.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":1,"acounts":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadProfile(path); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("want an unknown-field error, got %v", err)
	}
}

func TestProfilePathRejectsTraversal(t *testing.T) {
	for _, name := range []string{"", ".", "..", "../x", `a\b`, "a/b"} {
		if _, err := ProfilePath(name); err == nil {
			t.Errorf("ProfilePath(%q) should fail", name)
		}
	}
	got, err := ProfilePath("laptop")
	if err != nil || got != filepath.Join("profiles", "laptop.json") {
		t.Fatalf("ProfilePath(laptop) = %q, %v", got, err)
	}
}

func TestSaveProfileIsPrivateAtomicAndCreatesDirs(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "profiles", "nested")
	path := filepath.Join(dir, "demo.json")
	p := Default("demo")
	for i := 0; i < 2; i++ { // second call replaces the first
		if err := SaveProfile(p, path); err != nil {
			t.Fatalf("SaveProfile #%d: %v", i+1, err)
		}
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Fatalf("profile permissions = %o, want 600", perm)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("temporary files left behind: %v", entries)
	}
}
