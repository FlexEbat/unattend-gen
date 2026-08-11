package xmlgen

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// decodeBase64Payload extracts and decodes the first
// [Convert]::FromBase64String('...') payload in s.
func decodeBase64Payload(t *testing.T, s string) string {
	t.Helper()
	const marker = "FromBase64String('"
	start := strings.Index(s, marker)
	if start < 0 {
		t.Fatalf("no base64 payload found in %q", s)
	}
	start += len(marker)
	end := strings.Index(s[start:], "'")
	if end < 0 {
		t.Fatalf("unterminated base64 payload in %q", s)
	}
	decoded, err := base64.StdEncoding.DecodeString(s[start : start+end])
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	return string(decoded)
}

func TestBuildAnswerFileSystemScript(t *testing.T) {
	p := baseProfile()
	p.SystemScripts = []profile.CustomScript{{Format: profile.ScriptCmd, Content: "echo hello"}}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil {
		t.Fatal("expected a Microsoft-Windows-Deployment component")
	}
	if len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("got %d RunSynchronousCommand elements, want exactly 1", len(deployment.RunSynchronousCommand))
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	if !strings.Contains(cmdLine, `cmd.exe /c \"C:\Windows\Setup\Scripts\unattend-sys-01.cmd\"`) {
		t.Fatalf("command = %q, want it to invoke the staged .cmd script", cmdLine)
	}
	if got := decodeBase64Payload(t, cmdLine); got != "echo hello" {
		t.Fatalf("decoded script content = %q, want %q", got, "echo hello")
	}
}

func TestBuildAnswerFileDefaultUserScript(t *testing.T) {
	p := baseProfile()
	p.DefaultUserScripts = []profile.CustomScript{{Format: profile.ScriptReg, Content: "Windows Registry Editor Version 5.00"}}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	if !strings.Contains(cmdLine, `reg.exe load HKU\DefaultUser`) {
		t.Fatalf("command = %q, want it to load the default user hive", cmdLine)
	}
	if !strings.Contains(cmdLine, `reg.exe unload HKU\DefaultUser`) {
		t.Fatalf("command = %q, want it to unload the default user hive", cmdLine)
	}
	if !strings.Contains(cmdLine, `reg.exe import \"C:\Windows\Setup\Scripts\unattend-du-01.reg\"`) {
		t.Fatalf("command = %q, want it to import the staged .reg script", cmdLine)
	}
}

func TestValidateProfileRejectsVbsForDefaultUserScripts(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"default_user_scripts": [{"format": "vbs", "content": "WScript.Echo 1"}]
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for a .vbs DefaultUser script")
	}
}

func TestBuildAnswerFileFirstLogonScript(t *testing.T) {
	p := baseProfile()
	p.FirstLogonScripts = []profile.CustomScript{{Format: profile.ScriptPs1, Content: "Write-Host 'hi'"}}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 1 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 1", len(lines))
	}
	if !strings.Contains(lines[0], `powershell.exe -WindowStyle Normal -ExecutionPolicy Unrestricted -NoProfile -File \"C:\Windows\Setup\Scripts\unattend-fl-01.ps1\"`) {
		t.Fatalf("command = %q, want it to invoke the staged .ps1 script", lines[0])
	}
}

func TestBuildAnswerFileUserOnceScript(t *testing.T) {
	p := baseProfile()
	p.UserOnceScripts = []profile.CustomScript{{Format: profile.ScriptCmd, Content: "echo once"}}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	if !strings.Contains(cmdLine, `RunOnce`) {
		t.Fatalf("command = %q, want it to register a RunOnce entry", cmdLine)
	}
	if !strings.Contains(cmdLine, `unattend-uo-01-run.cmd`) {
		t.Fatalf("command = %q, want it to reference the wrapper .cmd file", cmdLine)
	}
}

func TestBuildAnswerFileRestartExplorerAfterScripts(t *testing.T) {
	p := baseProfile()
	p.FirstLogonScripts = []profile.CustomScript{{Format: profile.ScriptCmd, Content: "echo x"}}
	p.RestartExplorerAfterScripts = true

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 2 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 2 (script + restart)", len(lines))
	}
	if !strings.Contains(lines[1], "Stop-Process -Name explorer") {
		t.Fatalf("second command = %q, want the explorer restart", lines[1])
	}
}

func TestBuildAnswerFileRestartExplorerWithoutScriptsAddsNothing(t *testing.T) {
	p := baseProfile()
	p.RestartExplorerAfterScripts = true

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 0 {
		t.Fatalf("expected no FirstLogonCommands entries when there is nothing to restart for, got %v", lines)
	}
}
