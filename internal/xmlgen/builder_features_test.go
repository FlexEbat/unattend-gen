package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFileRemoveFeaturesEmptyAddsNoCommand(t *testing.T) {
	p := baseProfile()
	p.RemoveFeatures = nil

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 0 {
		t.Fatalf("expected no FirstLogonCommands entries, got %v", lines)
	}
}

func TestBuildAnswerFileRemoveFeaturesAddsOneCommand(t *testing.T) {
	p := baseProfile()
	p.RemoveFeatures = []profile.RemovableFeature{profile.FeatureWordPad, profile.FeatureInternetExplorer}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 1 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 1", len(lines))
	}
	cmd := lines[0]
	if !strings.Contains(cmd, "Remove-WindowsCapability") {
		t.Fatalf("command = %q, want it to run Remove-WindowsCapability", cmd)
	}
	if !strings.Contains(cmd, "'Microsoft.Windows.WordPad'") {
		t.Fatalf("command = %q, want it to reference the WordPad capability prefix", cmd)
	}
	if !strings.Contains(cmd, "'Browser.InternetExplorer'") {
		t.Fatalf("command = %q, want it to reference the Internet Explorer capability prefix", cmd)
	}
}

func TestBuildAnswerFileAppsAndFeaturesAreSeparateOrderedCommands(t *testing.T) {
	p := baseProfile()
	p.RemoveApps = []profile.RemovableApp{profile.AppTeams}
	p.RemoveFeatures = []profile.RemovableFeature{profile.FeatureOpenSSHClient}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 2 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 2 (apps + features)", len(lines))
	}
	if !strings.Contains(lines[0], "Remove-AppxProvisionedPackage") {
		t.Fatalf("first command = %q, want the app-removal command first", lines[0])
	}
	if !strings.Contains(lines[1], "Remove-WindowsCapability") {
		t.Fatalf("second command = %q, want the feature-removal command second", lines[1])
	}
}

func TestValidateProfileRejectsUnknownRemoveFeature(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"remove_features": ["NotARealFeature"]
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for an unknown feature name")
	}
}

func TestValidateProfileAcceptsKnownRemoveFeatures(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"remove_features": ["WordPad", "InternetExplorer", "Speech"]
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) != 0 {
		t.Fatalf("expected known feature names to validate, got errors: %v", result.Errors)
	}
}
