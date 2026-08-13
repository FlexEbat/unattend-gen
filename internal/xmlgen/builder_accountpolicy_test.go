package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFilePasswordExpirationDefaultAddsNoCommand(t *testing.T) {
	p := baseProfile()
	p.PasswordExpiration = profile.PasswordExpirationSettings{}

	doc := buildAndParseAccounts(t, p)
	if shell := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment"); shell != nil {
		t.Fatalf("expected no Microsoft-Windows-Deployment component, got %+v", shell)
	}
}

func TestBuildAnswerFilePasswordExpirationNever(t *testing.T) {
	p := baseProfile()
	p.PasswordExpiration = profile.PasswordExpirationSettings{Mode: profile.PasswordExpirationNever}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	if !strings.Contains(deployment.RunSynchronousCommand[0].Path, "net accounts /maxpwage:UNLIMITED") {
		t.Fatalf("command = %q, want net accounts /maxpwage:UNLIMITED", deployment.RunSynchronousCommand[0].Path)
	}
}

func TestBuildAnswerFilePasswordExpirationCustom(t *testing.T) {
	p := baseProfile()
	days := 90
	p.PasswordExpiration = profile.PasswordExpirationSettings{Mode: profile.PasswordExpirationCustom, Days: &days}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	if !strings.Contains(deployment.RunSynchronousCommand[0].Path, "net accounts /maxpwage:90") {
		t.Fatalf("command = %q, want net accounts /maxpwage:90", deployment.RunSynchronousCommand[0].Path)
	}
}

func TestValidateProfileRejectsCustomPasswordExpirationWithoutDays(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"password_expiration": {"mode": "custom"}
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for custom password expiration without days")
	}
}

func TestBuildAnswerFileAccountLockoutDisabled(t *testing.T) {
	p := baseProfile()
	p.AccountLockout = profile.AccountLockoutSettings{Mode: profile.AccountLockoutDisabled}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	if !strings.Contains(deployment.RunSynchronousCommand[0].Path, "net accounts /lockoutthreshold:0") {
		t.Fatalf("command = %q, want net accounts /lockoutthreshold:0", deployment.RunSynchronousCommand[0].Path)
	}
}

func TestBuildAnswerFileAccountLockoutCustom(t *testing.T) {
	p := baseProfile()
	threshold, window, duration := 5, 15, 30
	p.AccountLockout = profile.AccountLockoutSettings{
		Mode: profile.AccountLockoutCustom, Threshold: &threshold, WindowMinutes: &window, DurationMinutes: &duration,
	}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	want := "net accounts /lockoutthreshold:5 /lockoutwindow:15 /lockoutduration:30"
	if !strings.Contains(deployment.RunSynchronousCommand[0].Path, want) {
		t.Fatalf("command = %q, want it to contain %q", deployment.RunSynchronousCommand[0].Path, want)
	}
}

func TestValidateProfileRejectsCustomAccountLockoutWithoutFields(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"account_lockout": {"mode": "custom"}
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) != 3 {
		t.Fatalf("expected 3 errors (threshold, window, duration), got %v", result.Errors)
	}
}
