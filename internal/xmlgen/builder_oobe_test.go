package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFileTimezoneSet(t *testing.T) {
	p := baseProfile()
	tz := "Russian Standard Time"
	p.Timezone = &tz

	doc := buildAndParseAccounts(t, p)
	shell := findShellComponent(doc, "specialize")
	if shell == nil {
		t.Fatal("expected Shell-Setup component in specialize pass")
	}
	if shell.TimeZone != tz {
		t.Fatalf("TimeZone = %q, want %q", shell.TimeZone, tz)
	}
}

func TestBuildAnswerFileTimezoneNilOmitsElement(t *testing.T) {
	p := baseProfile()
	p.Timezone = nil

	doc := buildAndParseAccounts(t, p)
	if shell := findShellComponent(doc, "specialize"); shell != nil {
		t.Fatalf("expected no Shell-Setup component in specialize pass, got %+v", shell)
	}
}

func TestBuildAnswerFileNonInteractiveExpressHidesOOBEScreens(t *testing.T) {
	p := baseProfile()
	p.ExpressSettings = profile.ExpressSettings{Mode: profile.ExpressAllDisabled}

	doc := buildAndParseAccounts(t, p)
	shell := findShellComponent(doc, "oobeSystem")
	if shell == nil || shell.OOBE == nil {
		t.Fatal("expected an OOBE block")
	}
	oobe := shell.OOBE
	if !oobe.HideEULAPage || !oobe.HideOEMRegistrationScreen || !oobe.HideLocalAccountScreen || !oobe.HideWirelessSetupInOOBE {
		t.Fatalf("expected the full hide bundle for non-interactive express settings, got %+v", oobe)
	}
	if oobe.NetworkLocation != "Work" {
		t.Fatalf("NetworkLocation = %q, want Work", oobe.NetworkLocation)
	}
}

func TestBuildAnswerFileInteractiveExpressNoAccountsOmitsOOBEBlock(t *testing.T) {
	p := baseProfile() // ExpressInteractive, no accounts, no bypass by default

	doc := buildAndParseAccounts(t, p)
	if shell := findShellComponent(doc, "oobeSystem"); shell != nil {
		t.Fatalf("expected no oobeSystem Shell-Setup component, got %+v", shell)
	}
}

func TestBuildAnswerFileBypassOnlineAccountRequirement(t *testing.T) {
	p := baseProfile()
	p.BypassOnlineAccountRequirement = true

	doc := buildAndParseAccounts(t, p)

	shell := findShellComponent(doc, "oobeSystem")
	if shell == nil || shell.OOBE == nil || !shell.OOBE.HideOnlineAccountScreens {
		t.Fatal("expected HideOnlineAccountScreens for BypassOnlineAccountRequirement")
	}

	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil {
		t.Fatal("expected a Microsoft-Windows-Deployment component")
	}
	if len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("got %d RunSynchronousCommand elements, want exactly 1", len(deployment.RunSynchronousCommand))
	}
	if !strings.Contains(deployment.RunSynchronousCommand[0].Path, "BypassNRO") {
		t.Fatalf("command Path = %q, want it to contain BypassNRO", deployment.RunSynchronousCommand[0].Path)
	}
}

func TestBuildAnswerFileBypassOnlineAccountRequirementOffAddsNothing(t *testing.T) {
	p := baseProfile()
	p.BypassOnlineAccountRequirement = false

	doc := buildAndParseAccounts(t, p)
	if shell := findShellComponent(doc, "oobeSystem"); shell != nil {
		t.Fatalf("expected no oobeSystem Shell-Setup component, got %+v", shell)
	}
	if shell := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment"); shell != nil {
		t.Fatalf("expected no Microsoft-Windows-Deployment component, got %+v", shell)
	}
}

func TestValidateProfileRejectsEmptyTimezone(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"timezone": "",
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"}
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for an empty timezone string")
	}
}
