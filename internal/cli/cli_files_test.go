package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// inTempDir runs the test with the working directory set to a fresh temp dir.
func inTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	return dir
}

const profileWithPassword = `{
	"schema_version": 1,
	"name": "demo",
	"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
	"edition": {"mode": "interactive"},
	"accounts": [{"name": "user", "password": "hunter2", "group": "Administrators"}],
	"first_logon": {"mode": "none"},
	"express_settings": {"mode": "interactive"}
}`

func TestProfileInitWritesIntoProfilesDirAndListSeesIt(t *testing.T) {
	dir := inTempDir(t)

	stdout, stderr, err := runCLI(t, "profile", "init", "demo")
	if err != nil {
		t.Fatalf("init: %v (%s)", err, stderr)
	}
	want := filepath.Join("profiles", "demo.json")
	if got := strings.TrimSpace(stdout + stderr); got != want {
		t.Fatalf("init printed %q, want %q", got, want)
	}
	if _, err := os.Stat(filepath.Join(dir, want)); err != nil {
		t.Fatalf("profile not created: %v", err)
	}
	if runtime.GOOS != "windows" {
		info, _ := os.Stat(filepath.Join(dir, want))
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("profile permissions %o, want 600", info.Mode().Perm())
		}
	}

	stdout, stderr, err = runCLI(t, "profile", "list")
	if err != nil || !strings.Contains(stdout+stderr, "demo.json") {
		t.Fatalf("list did not show the new profile: %q %q %v", stdout, stderr, err)
	}
}

func TestProfileInitRefusesToOverwriteWithoutForce(t *testing.T) {
	inTempDir(t)
	if _, _, err := runCLI(t, "profile", "init", "demo"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("profiles", "demo.json")
	if err := os.WriteFile(path, []byte("keep me"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, err := runCLI(t, "profile", "init", "demo"); err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("want an 'already exists' error mentioning --force, got %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "keep me" {
		t.Fatal("existing profile was overwritten without --force")
	}
	if _, _, err := runCLI(t, "profile", "init", "demo", "--force"); err != nil {
		t.Fatalf("--force: %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) == "keep me" {
		t.Fatal("--force did not replace the profile")
	}
}

func TestProfileInitRejectsPathTraversal(t *testing.T) {
	dir := inTempDir(t)
	if _, _, err := runCLI(t, "profile", "init", "../evil"); err == nil {
		t.Fatal("expected an error for a name with path separators")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(dir), "evil.json")); err == nil {
		t.Fatal("file was written outside the profiles directory")
	}
}

func TestGenerateRefusesToOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	profilePath := writeProfile(t, dir, "demo.json", validProfileJSON)
	out := filepath.Join(dir, "autounattend.xml")
	if err := os.WriteFile(out, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, err := runCLI(t, "generate", profilePath); err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("want an 'already exists' error, got %v", err)
	}
	if data, _ := os.ReadFile(out); string(data) != "old" {
		t.Fatal("existing answer file was overwritten without --force")
	}
	if _, stderr, err := runCLI(t, "generate", profilePath, "--force"); err != nil {
		t.Fatalf("--force: %v (%s)", err, stderr)
	}
	if data, _ := os.ReadFile(out); !strings.Contains(string(data), "<unattend") {
		t.Fatal("--force did not write the answer file")
	}
}

func TestGenerateAnswerFileIsPrivate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permissions")
	}
	dir := t.TempDir()
	profilePath := writeProfile(t, dir, "demo.json", profileWithPassword)
	if _, stderr, err := runCLI(t, "generate", profilePath); err != nil {
		t.Fatalf("generate: %v (%s)", err, stderr)
	}
	info, err := os.Stat(filepath.Join(dir, "autounattend.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("answer file permissions %o, want 600", info.Mode().Perm())
	}
}

func TestGenerateToStdout(t *testing.T) {
	dir := inTempDir(t)
	profilePath := writeProfile(t, dir, "demo.json", validProfileJSON)

	stdout, _, err := runCLI(t, "generate", profilePath, "-o", "-")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(stdout, "<?xml") || !strings.Contains(stdout, "<unattend") {
		t.Fatalf("stdout does not look like an answer file: %.80q", stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, "autounattend.xml")); err == nil {
		t.Fatal("-o - must not write a file")
	}
}

func TestWarningsGoToStderrAndDoNotFail(t *testing.T) {
	dir := t.TempDir()
	profilePath := writeProfile(t, dir, "demo.json", profileWithPassword)

	stdout, stderr, err := runCLI(t, "validate", profilePath)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if !strings.Contains(stderr, "Предупреждение") || !strings.Contains(stderr, "открытым текстом") {
		t.Fatalf("want a plaintext-password warning on stderr, got %q", stderr)
	}
	if strings.Contains(stdout, "Предупреждение") {
		t.Fatal("warnings must not go to stdout")
	}
}

func TestProfileShowMasksPasswords(t *testing.T) {
	dir := t.TempDir()
	profilePath := writeProfile(t, dir, "demo.json", profileWithPassword)

	stdout, _, err := runCLI(t, "profile", "show", profilePath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stdout, "hunter2") {
		t.Fatal("profile show leaked a password")
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(stdout), &decoded); err != nil {
		t.Fatalf("show output is not JSON: %v", err)
	}
	if !strings.Contains(stdout, "********") {
		t.Fatal("password placeholder missing")
	}
}

func TestSchemaCommandPrintsJSON(t *testing.T) {
	stdout, _, err := runCLI(t, "schema")
	if err != nil {
		t.Fatal(err)
	}
	var s map[string]any
	if err := json.Unmarshal([]byte(stdout), &s); err != nil {
		t.Fatalf("schema output is not JSON: %v", err)
	}
	if s["title"] == nil {
		t.Fatal("schema has no title")
	}
}

func TestVersionFlag(t *testing.T) {
	stdout, stderr, err := runCLI(t, "--version")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout+stderr, "unattend-gen version") {
		t.Fatalf("unexpected --version output %q %q", stdout, stderr)
	}
}

func TestValidateRejectsUnknownField(t *testing.T) {
	dir := t.TempDir()
	path := writeProfile(t, dir, "typo.json", strings.Replace(validProfileJSON, `"accounts"`, `"acounts"`, 1))
	_, stderr, err := runCLI(t, "validate", path)
	if err == nil || !strings.Contains(stderr, "acounts") {
		t.Fatalf("want an unknown-field failure naming the field, got err=%v stderr=%q", err, stderr)
	}
}
