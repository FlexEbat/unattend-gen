package cli

import (
	"bytes"
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"
)

const validProfileJSON = `{
	"schema_version": 1,
	"name": "demo",
	"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
	"edition": {"mode": "interactive"},
	"accounts": [],
	"first_logon": {"mode": "none"},
	"express_settings": {"mode": "interactive"}
}`

const invalidProfileJSON = `{
	"schema_version": 1,
	"name": "demo",
	"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
	"edition": {"mode": "custom_key"},
	"accounts": [],
	"first_logon": {"mode": "none"},
	"express_settings": {"mode": "interactive"}
}`

func writeProfile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write profile fixture: %v", err)
	}
	return path
}

func runCLI(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := NewRootCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return out.String(), errOut.String(), err
}

func TestGenerateValidProfileCreatesParsableFile(t *testing.T) {
	dir := t.TempDir()
	profilePath := writeProfile(t, dir, "demo.json", validProfileJSON)

	_, stderr, err := runCLI(t, "generate", profilePath)
	if err != nil {
		t.Fatalf("generate: unexpected error %v, stderr: %s", err, stderr)
	}

	answerPath := filepath.Join(dir, "autounattend.xml")
	data, readErr := os.ReadFile(answerPath)
	if readErr != nil {
		t.Fatalf("expected %s to exist: %v", answerPath, readErr)
	}
	var doc struct {
		XMLName xml.Name `xml:"unattend"`
	}
	if err := xml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("answer file does not parse as XML: %v", err)
	}
}

func TestGenerateInvalidProfileWritesNoFile(t *testing.T) {
	dir := t.TempDir()
	profilePath := writeProfile(t, dir, "demo.json", invalidProfileJSON)

	_, stderr, err := runCLI(t, "generate", profilePath)
	if err == nil {
		t.Fatal("expected an error for invalid profile, got nil")
	}
	if stderr == "" {
		t.Fatal("expected validation errors on stderr")
	}

	answerPath := filepath.Join(dir, "autounattend.xml")
	if _, statErr := os.Stat(answerPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected no answer file to be written, stat err: %v", statErr)
	}
}

func TestGenerateWithoutOutputFlagWritesNextToProfile(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "profiles")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	profilePath := writeProfile(t, sub, "demo.json", validProfileJSON)

	if _, _, err := runCLI(t, "generate", profilePath); err != nil {
		t.Fatalf("generate: %v", err)
	}

	wantPath := filepath.Join(sub, "autounattend.xml")
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("expected answer file next to profile at %s: %v", wantPath, err)
	}
}

func TestGenerateMissingProfileFileFailsCleanly(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "does-not-exist.json")

	_, _, err := runCLI(t, "generate", missing)
	if err == nil {
		t.Fatal("expected an error for a missing profile file, got nil")
	}
	if err.Error() == "" {
		t.Fatal("expected a non-empty, non-panic error message")
	}
}
