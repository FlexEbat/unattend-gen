package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFileFileExplorerZeroValueAddsNoCommand(t *testing.T) {
	p := baseProfile()
	p.FileExplorer = profile.FileExplorerSettings{}

	doc := buildAndParseAccounts(t, p)
	if shell := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment"); shell != nil {
		t.Fatalf("expected no Microsoft-Windows-Deployment component, got %+v", shell)
	}
}

func TestBuildAnswerFileFileExplorerHiddenFiles(t *testing.T) {
	cases := []struct {
		mode            profile.HiddenFilesMode
		wantSuperHidden string
	}{
		{profile.HiddenFilesShowHidden, "ShowSuperHidden /t REG_DWORD /d 0"},
		{profile.HiddenFilesShowAll, "ShowSuperHidden /t REG_DWORD /d 1"},
	}
	for _, tc := range cases {
		p := baseProfile()
		p.FileExplorer = profile.FileExplorerSettings{HiddenFiles: tc.mode}

		doc := buildAndParseAccounts(t, p)
		deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
		if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
			t.Fatalf("mode %s: expected exactly 1 Deployment command, got %+v", tc.mode, deployment)
		}
		cmd := deployment.RunSynchronousCommand[0].Path
		if !strings.Contains(cmd, `reg.exe add \"HKU\DefaultUser\Software\Microsoft\Windows\CurrentVersion\Explorer\Advanced\" /v Hidden /t REG_DWORD /d 1`) {
			t.Fatalf("mode %s: command = %q, want it to set Hidden=1", tc.mode, cmd)
		}
		if !strings.Contains(cmd, tc.wantSuperHidden) {
			t.Fatalf("mode %s: command = %q, want it to contain %q", tc.mode, cmd, tc.wantSuperHidden)
		}
		if !strings.Contains(cmd, `reg.exe load HKU\DefaultUser`) || !strings.Contains(cmd, `reg.exe unload HKU\DefaultUser`) {
			t.Fatalf("mode %s: command = %q, want the default user hive mounted and unmounted", tc.mode, cmd)
		}
	}
}

func TestBuildAnswerFileFileExplorerClassicContextMenu(t *testing.T) {
	p := baseProfile()
	p.FileExplorer = profile.FileExplorerSettings{ClassicContextMenu: true}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmd := deployment.RunSynchronousCommand[0].Path
	if !strings.Contains(cmd, `{86ca1aa0-34aa-4e8b-a509-50c905bae2a2}`) {
		t.Fatalf("command = %q, want the classic context menu CLSID", cmd)
	}
}

func TestBuildAnswerFileFileExplorerCombinesMultipleSettingsIntoOneCommand(t *testing.T) {
	p := baseProfile()
	p.FileExplorer = profile.FileExplorerSettings{
		ShowFileExtensions:   true,
		OpenToThisPC:         true,
		ShowEndTaskInTaskbar: true,
	}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmd := deployment.RunSynchronousCommand[0].Path
	for _, want := range []string{"HideFileExt", "LaunchTo", "TaskbarEndTask"} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("command = %q, want it to contain %q", cmd, want)
		}
	}
}
