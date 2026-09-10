package xmlgen

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFileDeleteEdgeDesktopIcon(t *testing.T) {
	p := baseProfile()
	p.SystemTweaks = profile.SystemTweaks{DeleteEdgeDesktopIcon: true}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 2 {
		t.Fatalf("expected exactly 2 Deployment commands (Public desktop + UserOnce), got %+v", deployment)
	}
	var joined string
	for _, c := range deployment.RunSynchronousCommand {
		joined += c.Path + "\n"
	}
	if !strings.Contains(joined, `C:\Users\Public\Desktop\Microsoft Edge.lnk`) {
		t.Fatalf("commands = %q, want the Public desktop shortcut removed", joined)
	}
	if !strings.Contains(joined, `reg.exe load HKU\DefaultUser`) {
		t.Fatalf("commands = %q, want a default-user-hive mount for the UserOnce entry", joined)
	}
	if !strings.Contains(joined, "UnattendDeleteEdgeIcon") {
		t.Fatalf("commands = %q, want a RunOnce entry for future accounts", joined)
	}
}

func TestBuildAnswerFileDeleteEdgeDesktopIconOffAddsNothing(t *testing.T) {
	p := baseProfile()
	p.SystemTweaks = profile.SystemTweaks{DeleteEdgeDesktopIcon: false}

	doc := buildAndParseAccounts(t, p)
	if deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment"); deployment != nil {
		t.Fatalf("expected no Microsoft-Windows-Deployment component, got %+v", deployment)
	}
}

func TestBuildAnswerFileHidePowerShellWindowsAffectsUserScripts(t *testing.T) {
	p := baseProfile()
	p.HidePowerShellWindows = true
	p.FirstLogonScripts = []profile.CustomScript{{Format: profile.ScriptPs1, Content: "'hi'"}}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 1 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 1", len(lines))
	}
	if !strings.Contains(lines[0], "-WindowStyle Hidden") {
		t.Fatalf("command = %q, want -WindowStyle Hidden", lines[0])
	}
	if strings.Contains(lines[0], "-WindowStyle Normal") {
		t.Fatalf("command = %q, want no -WindowStyle Normal left over", lines[0])
	}
}

func TestBuildAnswerFileHidePowerShellWindowsDefaultIsNormal(t *testing.T) {
	p := baseProfile()
	p.HidePowerShellWindows = false
	p.FirstLogonScripts = []profile.CustomScript{{Format: profile.ScriptPs1, Content: "'hi'"}}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 1 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 1", len(lines))
	}
	if !strings.Contains(lines[0], "-WindowStyle Normal") {
		t.Fatalf("command = %q, want -WindowStyle Normal (default)", lines[0])
	}
}

func TestBuildAnswerFileHidePowerShellWindowsAffectsBuiltinTweaks(t *testing.T) {
	p := baseProfile()
	p.HidePowerShellWindows = true
	p.SystemTweaks = profile.SystemTweaks{DeleteHiddenJunctions: true}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 1 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 1", len(lines))
	}
	if !strings.Contains(lines[0], "-WindowStyle Hidden") {
		t.Fatalf("command = %q, want the built-in DeleteHiddenJunctions script hidden too", lines[0])
	}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	// The wrapper .cmd's invocation line is base64-embedded as the SECOND
	// FromBase64String payload (the first is the .ps1 script content
	// itself, which doesn't mention WindowStyle).
	payloads := decodeAllBase64Payloads(t, cmdLine)
	if len(payloads) < 2 {
		t.Fatalf("expected at least 2 base64 payloads in %q, got %d", cmdLine, len(payloads))
	}
	if !strings.Contains(payloads[1], "-WindowStyle Hidden") {
		t.Fatalf("decoded wrapper = %q, want the UserOnce wrapper's invocation hidden too", payloads[1])
	}
}

func decodeAllBase64Payloads(t *testing.T, s string) []string {
	t.Helper()
	const marker = "FromBase64String('"
	var out []string
	rest := s
	for {
		idx := strings.Index(rest, marker)
		if idx < 0 {
			break
		}
		rest = rest[idx+len(marker):]
		end := strings.Index(rest, "'")
		if end < 0 {
			t.Fatalf("unterminated base64 payload in %q", s)
		}
		decoded, err := base64.StdEncoding.DecodeString(rest[:end])
		if err != nil {
			t.Fatalf("base64 decode: %v", err)
		}
		out = append(out, string(decoded))
		rest = rest[end:]
	}
	return out
}

func TestBuildAnswerFileHidePowerShellWindowsDoesNotAffectCmdOrReg(t *testing.T) {
	p := baseProfile()
	p.HidePowerShellWindows = true
	p.FirstLogonScripts = []profile.CustomScript{{Format: profile.ScriptCmd, Content: "echo hi"}}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 1 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 1", len(lines))
	}
	if strings.Contains(lines[0], "-WindowStyle") {
		t.Fatalf("command = %q, want cmd.exe invocation unaffected by HidePowerShellWindows", lines[0])
	}
}
