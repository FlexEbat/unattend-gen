package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFileRemoveAppsEmptyAddsNoCommand(t *testing.T) {
	p := baseProfile()
	p.RemoveApps = nil

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 0 {
		t.Fatalf("expected no FirstLogonCommands entries, got %v", lines)
	}
}

func TestBuildAnswerFileRemoveAppsAddsOneCommand(t *testing.T) {
	p := baseProfile()
	p.RemoveApps = []profile.RemovableApp{profile.AppClipchamp, profile.AppCortana}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 1 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 1", len(lines))
	}
	cmd := lines[0]
	if !strings.Contains(cmd, "Remove-AppxProvisionedPackage") {
		t.Fatalf("command = %q, want it to run Remove-AppxProvisionedPackage", cmd)
	}
	if !strings.Contains(cmd, "'Clipchamp'") {
		t.Fatalf("command = %q, want it to reference Clipchamp", cmd)
	}
	if !strings.Contains(cmd, "'549981C3F5F10'") {
		t.Fatalf("command = %q, want it to reference Cortana's package id", cmd)
	}
}

func TestBuildAnswerFileWifiAndRemoveAppsAreSeparateOrderedCommands(t *testing.T) {
	p := baseProfile()
	p.Wifi = &profile.WifiSettings{SSID: "Net", Authentication: profile.WifiOpen}
	p.RemoveApps = []profile.RemovableApp{profile.AppTeams}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 2 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 2 (wifi + apps)", len(lines))
	}
	if !strings.Contains(lines[0], "netsh wlan add profile") {
		t.Fatalf("first command = %q, want the Wi-Fi command first", lines[0])
	}
	if !strings.Contains(lines[1], "Remove-AppxProvisionedPackage") {
		t.Fatalf("second command = %q, want the app-removal command second", lines[1])
	}
}

func TestValidateProfileRejectsUnknownRemoveApp(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"remove_apps": ["NotARealApp"]
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for an unknown app name")
	}
}

func TestValidateProfileAcceptsKnownRemoveApps(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"remove_apps": ["Clipchamp", "Cortana", "XboxApps"]
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) != 0 {
		t.Fatalf("expected known app names to validate, got errors: %v", result.Errors)
	}
}
