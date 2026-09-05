package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFileDesktopIcons(t *testing.T) {
	p := baseProfile()
	p.DesktopIcons = map[profile.DesktopIcon]bool{
		profile.DesktopIconThisPC:     true,
		profile.DesktopIconRecycleBin: false,
	}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	if !strings.Contains(cmdLine, "UnattendDesktopIcons") {
		t.Fatalf("command = %q, want a RunOnce entry for desktop icons", cmdLine)
	}
	if !strings.Contains(cmdLine, `reg.exe load HKU\DefaultUser`) {
		t.Fatalf("command = %q, want a default-user-hive mount", cmdLine)
	}
	decoded := decodeBase64Payload(t, cmdLine)
	if !strings.Contains(decoded, "ClassicStartMenu") || !strings.Contains(decoded, "NewStartPanel") {
		t.Fatalf("decoded script = %q, want both HideDesktopIcons subkeys touched", decoded)
	}
	// This PC = shown -> value 0; Recycle Bin = hidden -> value 1.
	if !strings.Contains(decoded, `"{20D04FE0-3AEA-1069-A2D8-08002B30309D}" /t REG_DWORD /d 0`) {
		t.Fatalf("decoded script = %q, want This PC set to shown (0)", decoded)
	}
	if !strings.Contains(decoded, `"{645FF040-5081-101B-9F08-00AA002F954E}" /t REG_DWORD /d 1`) {
		t.Fatalf("decoded script = %q, want Recycle Bin set to hidden (1)", decoded)
	}
	if !strings.Contains(decoded, "taskkill") || !strings.Contains(decoded, "explorer.exe") {
		t.Fatalf("decoded script = %q, want an Explorer restart", decoded)
	}
}

func TestBuildAnswerFileDesktopIconsEmptyAddsNothing(t *testing.T) {
	p := baseProfile()
	p.DesktopIcons = nil

	doc := buildAndParseAccounts(t, p)
	if deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment"); deployment != nil {
		t.Fatalf("expected no Microsoft-Windows-Deployment component, got %+v", deployment)
	}
}

func TestBuildAnswerFileStartFolders(t *testing.T) {
	p := baseProfile()
	p.StartFolders = []profile.StartFolder{profile.StartFolderSettings, profile.StartFolderDocuments}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	if !strings.Contains(cmdLine, "UnattendStartFolders") {
		t.Fatalf("command = %q, want a RunOnce entry for Start folders", cmdLine)
	}
	decoded := decodeBase64Payload(t, cmdLine)
	if !strings.Contains(decoded, "VisiblePlaces") || !strings.Contains(decoded, "REG_BINARY") {
		t.Fatalf("decoded script = %q, want a VisiblePlaces REG_BINARY write", decoded)
	}
	// Settings GUID bytes then Documents GUID bytes, concatenated in order:
	// 32 hex chars each = 64 hex chars total for 2 folders.
	idx := strings.Index(decoded, "/d ")
	if idx == -1 {
		t.Fatalf("decoded script = %q, missing /d value", decoded)
	}
	hexValue := strings.Fields(decoded[idx+len("/d "):])[0]
	if len(hexValue) != 64 {
		t.Fatalf("VisiblePlaces hex = %q (len %d), want 64 hex chars (2 GUIDs)", hexValue, len(hexValue))
	}
	if !strings.HasPrefix(hexValue, "86087352aa5143429f7b2776584659d4") {
		t.Fatalf("VisiblePlaces hex = %q, want it to start with the Settings GUID (order preserved)", hexValue)
	}
}

func TestBuildAnswerFileStartFoldersEmptyAddsNothing(t *testing.T) {
	p := baseProfile()
	p.StartFolders = nil

	doc := buildAndParseAccounts(t, p)
	if deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment"); deployment != nil {
		t.Fatalf("expected no Microsoft-Windows-Deployment component, got %+v", deployment)
	}
}

func TestValidateProfileRejectsUnknownDesktopIcon(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"desktop_icons": {"NotReal": true}
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for an unknown desktop icon")
	}
}

func TestValidateProfileRejectsUnknownStartFolder(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"start_folders": ["NotReal"]
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for an unknown Start folder")
	}
}
