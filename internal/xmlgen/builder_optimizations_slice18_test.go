package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFileHardenSystemDriveACL(t *testing.T) {
	p := baseProfile()
	p.SystemTweaks = profile.SystemTweaks{HardenSystemDriveACL: true}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	if !strings.Contains(cmdLine, "icacls.exe") || !strings.Contains(cmdLine, "S-1-5-11") {
		t.Fatalf("command = %q, want an icacls.exe removal of S-1-5-11", cmdLine)
	}
}

func TestBuildAnswerFileMakeEdgeUninstallable(t *testing.T) {
	p := baseProfile()
	p.SystemTweaks = profile.SystemTweaks{MakeEdgeUninstallable: true}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	if got := decodeBase64Payload(t, cmdLine); !strings.Contains(got, "IntegratedServicesRegionPolicySet.json") || !strings.Contains(got, "1bca278a-5d11-4acf-ad2f-f9ab6d7f93a6") {
		t.Fatalf("decoded script = %q, want it to touch IntegratedServicesRegionPolicySet.json's Edge policy guid", got)
	}
}

func TestBuildAnswerFileDeleteWindowsOld(t *testing.T) {
	p := baseProfile()
	p.SystemTweaks = profile.SystemTweaks{DeleteWindowsOld: true}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 1 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 1", len(lines))
	}
	if !strings.Contains(lines[0], `rmdir C:\Windows.old`) {
		t.Fatalf("command = %q, want it to rmdir C:\\Windows.old", lines[0])
	}
}

func TestBuildAnswerFileGroupBRemainderOffAddsNothing(t *testing.T) {
	p := baseProfile()
	p.SystemTweaks = profile.SystemTweaks{}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 0 {
		t.Fatalf("expected no FirstLogonCommands entries, got %v", lines)
	}
	doc := buildAndParseAccounts(t, p)
	if deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment"); deployment != nil {
		t.Fatalf("expected no Microsoft-Windows-Deployment component, got %+v", deployment)
	}
}
