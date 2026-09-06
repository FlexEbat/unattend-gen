package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFileDisableCoreIsolation(t *testing.T) {
	p := baseProfile()
	p.SystemTweaks = profile.SystemTweaks{DisableCoreIsolation: true}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	for _, want := range []string{"EnableVirtualizationBasedSecurity", "HypervisorEnforcedCodeIntegrity", "DeviceGuard"} {
		if !strings.Contains(cmdLine, want) {
			t.Fatalf("command = %q, want it to contain %q", cmdLine, want)
		}
	}
}

func TestBuildAnswerFileVMGuestTools(t *testing.T) {
	p := baseProfile()
	p.InstallVMGuestTools = []profile.VMGuestTool{
		profile.VMGuestToolVBoxGuestAdditions, profile.VMGuestToolVMwareTools,
		profile.VMGuestToolVirtIOGuestTools, profile.VMGuestToolParallelsTools,
	}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 4 {
		t.Fatalf("got %d FirstLogonCommands entries, want 4 (one per tool)", len(lines))
	}
	wantSubstrings := []string{"VBoxWindowsAdditions.exe", "VMware Tools", "virtio-win-guest-tools.exe", "PTAgent.exe"}
	for i, line := range lines {
		decoded := decodeBase64Payload(t, line)
		if !strings.Contains(decoded, wantSubstrings[i]) {
			t.Fatalf("command %d decoded = %q, want it to contain %q", i, decoded, wantSubstrings[i])
		}
	}
}

func TestBuildAnswerFileVMGuestToolsEmptyAddsNothing(t *testing.T) {
	p := baseProfile()
	p.InstallVMGuestTools = nil

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 0 {
		t.Fatalf("expected no FirstLogonCommands entries, got %v", lines)
	}
}

func TestBuildAnswerFileAppLockerPolicy(t *testing.T) {
	p := baseProfile()
	policy := `<AppLockerPolicy Version="1"><RuleCollection Type="Exe" EnforcementMode="Enabled" /></AppLockerPolicy>`
	p.AppLockerPolicyXML = &policy

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	if !strings.Contains(cmdLine, "AppIDSvc") {
		t.Fatalf("command = %q, want the AppIDSvc service to be started", cmdLine)
	}
	if !strings.Contains(cmdLine, "Set-AppLockerPolicy") {
		t.Fatalf("command = %q, want Set-AppLockerPolicy to be invoked", cmdLine)
	}
	if got := decodeBase64Payload(t, cmdLine); !strings.Contains(got, "AppLockerPolicy") || !strings.Contains(got, "RuleCollection") {
		t.Fatalf("decoded policy file = %q, want the embedded AppLocker policy XML", got)
	}
}

func TestBuildAnswerFileAppLockerPolicyNilAddsNothing(t *testing.T) {
	p := baseProfile()
	p.AppLockerPolicyXML = nil

	doc := buildAndParseAccounts(t, p)
	if deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment"); deployment != nil {
		t.Fatalf("expected no Microsoft-Windows-Deployment component, got %+v", deployment)
	}
}

func TestValidateProfileRejectsUnknownVMGuestTool(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"install_vm_guest_tools": ["NotReal"]
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for an unknown VM guest tool")
	}
}

func TestValidateProfileRejectsMalformedAppLockerXML(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"applocker_policy_xml": "<AppLockerPolicy><Unclosed></AppLockerPolicy>"
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for malformed AppLocker XML")
	}
}

func TestValidateProfileRejectsEmptyAppLockerXML(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"applocker_policy_xml": "   "
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for a blank AppLocker policy string")
	}
}

func TestValidateProfileAcceptsWellFormedAppLockerXML(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"applocker_policy_xml": "<AppLockerPolicy Version=\"1\"><RuleCollection Type=\"Exe\" EnforcementMode=\"Enabled\" /></AppLockerPolicy>"
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) != 0 {
		t.Fatalf("expected well-formed AppLocker XML to validate, got errors: %v", result.Errors)
	}
}
