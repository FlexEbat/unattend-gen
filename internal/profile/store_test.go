package profile

import (
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
