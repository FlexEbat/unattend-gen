package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

const taskbarIconsCustomXML = `<LayoutModificationTemplate xmlns="http://schemas.microsoft.com/Start/2014/LayoutModification" xmlns:defaultlayout="http://schemas.microsoft.com/Start/2014/FullDefaultLayout" xmlns:taskbar="http://schemas.microsoft.com/Start/2014/TaskbarLayout" Version="1"><CustomTaskbarLayoutCollection PinListPlacement="Replace"><defaultlayout:TaskbarLayout><taskbar:TaskbarPinList><taskbar:DesktopApp DesktopApplicationID="Microsoft.Windows.Explorer" /></taskbar:TaskbarPinList></defaultlayout:TaskbarLayout></CustomTaskbarLayoutCollection></LayoutModificationTemplate>`

func taskbarIconsCommandLine(t *testing.T, s profile.TaskbarIconsSettings) (string, []string) {
	t.Helper()
	p := baseProfile()
	p.TaskbarIcons = s
	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	return cmdLine, decodeAllBase64Payloads(t, cmdLine)
}

func anyContains(payloads []string, parts ...string) bool {
	for _, p := range payloads {
		ok := true
		for _, part := range parts {
			if !strings.Contains(p, part) {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

func TestBuildAnswerFileTaskbarIconsCustom(t *testing.T) {
	xml := taskbarIconsCustomXML
	cmdLine, payloads := taskbarIconsCommandLine(t, profile.TaskbarIconsSettings{Mode: profile.TaskbarIconsModeCustom, XML: &xml})
	// wrapCommand escapes double quotes for the outer powershell -Command.
	cmdLine = strings.ReplaceAll(cmdLine, `\"`, `"`)

	if !anyContains(payloads, "Microsoft.Windows.Explorer", "TaskbarPinList") {
		t.Error("custom layout XML is not embedded")
	}
	// Locked layout policy in the default user hive.
	for _, want := range []string{
		`reg.exe load HKU\DefaultUser`,
		`StartLayoutFile /t REG_SZ /d "C:\Windows\Setup\Scripts\unattend-taskbar-layout.xml"`,
		`LockedStartLayout /t REG_DWORD /d 1`,
		`reg.exe unload HKU\DefaultUser`,
	} {
		if !strings.Contains(cmdLine, want) {
			t.Errorf("command is missing %q", want)
		}
	}
	// Unlock flow: specialize script (policy, event source, task), the task
	// definition, the VBScript and the first-logon event writer.
	if !anyContains(payloads, "DisableCloudOptimizedContent", "CreateEventSource", "Register-ScheduledTask -TaskName 'UnlockStartLayout'") {
		t.Error("specialize script is missing the cloud-content policy, event source or unlock task registration")
	}
	if !anyContains(payloads, "<Task", "EventID=1", "UnattendGenerator", `C:\Windows\Setup\Scripts\unattend-unlock-start-layout.vbs`) {
		t.Error("unlock task XML is missing or does not point at our VBScript")
	}
	if !anyContains(payloads, "LockedStartLayout", "SetDWORDValue", "StdRegProv") {
		t.Error("unlock VBScript is missing")
	}
	if !anyContains(payloads, "EventLog]::WriteEntry", "UnattendGenerator") {
		t.Error("first-logon event writer is missing")
	}
	if !strings.Contains(cmdLine, "UnattendTaskbarIconsUnlock") {
		t.Error("RunOnce entry for the first-logon unlock request is missing")
	}
}

func TestBuildAnswerFileTaskbarIconsEmptyUsesLeaveEmptySentinel(t *testing.T) {
	_, payloads := taskbarIconsCommandLine(t, profile.TaskbarIconsSettings{Mode: profile.TaskbarIconsModeEmpty})
	if !anyContains(payloads, `PinListPlacement="Replace"`, `DesktopApplicationLinkPath="#leaveempty"`) {
		t.Error("empty mode must write the Replace + #leaveempty layout")
	}
}

func TestBuildAnswerFileTaskbarIconsUnlockVBScriptUsesCRLF(t *testing.T) {
	_, payloads := taskbarIconsCommandLine(t, profile.TaskbarIconsSettings{Mode: profile.TaskbarIconsModeEmpty})
	for _, p := range payloads {
		if strings.Contains(p, "StdRegProv") {
			if strings.Contains(strings.ReplaceAll(p, "\r\n", ""), "\n") {
				t.Error("VBScript contains bare LF line endings")
			}
			return
		}
	}
	t.Error("VBScript payload not found")
}

func TestBuildAnswerFileTaskbarIconsDefaultAddsNothing(t *testing.T) {
	for _, mode := range []profile.TaskbarIconsMode{"", profile.TaskbarIconsModeDefault} {
		p := baseProfile()
		p.TaskbarIcons = profile.TaskbarIconsSettings{Mode: mode}
		doc := buildAndParseAccounts(t, p)
		if d := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment"); d != nil && len(d.RunSynchronousCommand) != 0 {
			t.Errorf("mode %q: expected no Deployment commands, got %d", mode, len(d.RunSynchronousCommand))
		}
	}
}

func TestBuildAnswerFileTaskbarIconsHiddenPowerShell(t *testing.T) {
	p := baseProfile()
	p.HidePowerShellWindows = true
	p.TaskbarIcons = profile.TaskbarIconsSettings{Mode: profile.TaskbarIconsModeEmpty}
	doc := buildAndParseAccounts(t, p)
	d := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if d == nil || len(d.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", d)
	}
	if !strings.Contains(d.RunSynchronousCommand[0].Path, "-WindowStyle Hidden") {
		t.Error("hide_powershell_windows is not honoured by the taskbar icons scripts")
	}
}

func validateTaskbarIcons(t *testing.T, field string) profile.ValidationResult {
	t.Helper()
	return profile.ValidateProfile([]byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"taskbar_icons": ` + field + `
	}`))
}

func TestValidateProfileRejectsMalformedTaskbarIconsXML(t *testing.T) {
	if r := validateTaskbarIcons(t, `{"mode": "custom", "xml": "<Unclosed>"}`); len(r.Errors) == 0 {
		t.Fatal("expected an error for malformed taskbar_icons.xml")
	}
}

func TestValidateProfileRejectsCustomTaskbarIconsWithoutXML(t *testing.T) {
	for _, field := range []string{`{"mode": "custom"}`, `{"mode": "custom", "xml": "  "}`} {
		if r := validateTaskbarIcons(t, field); len(r.Errors) == 0 {
			t.Errorf("%s: expected an error for custom mode without xml", field)
		}
	}
}

func TestValidateProfileRejectsUnknownTaskbarIconsMode(t *testing.T) {
	if r := validateTaskbarIcons(t, `{"mode": "bogus"}`); len(r.Errors) == 0 {
		t.Fatal("expected an error for an unknown taskbar_icons.mode")
	}
}

func TestValidateProfileAcceptsTaskbarIcons(t *testing.T) {
	for _, field := range []string{`{"mode": "default"}`, `{"mode": "empty"}`, `{"mode": "custom", "xml": "<a/>"}`} {
		if r := validateTaskbarIcons(t, field); len(r.Errors) != 0 {
			t.Errorf("%s: unexpected errors %v", field, r.Errors)
		}
	}
}
