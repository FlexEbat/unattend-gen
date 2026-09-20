package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFileDisableWidgets(t *testing.T) {
	p := baseProfile()
	p.SystemTweaks = profile.SystemTweaks{DisableWidgets: true}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	if !strings.Contains(deployment.RunSynchronousCommand[0].Path, "AllowNewsAndInterests") {
		t.Fatalf("command = %q, want the widgets policy key", deployment.RunSynchronousCommand[0].Path)
	}
}

func TestBuildAnswerFileLeftTaskbarHideTaskViewDisableBing(t *testing.T) {
	p := baseProfile()
	p.SystemTweaks = profile.SystemTweaks{LeftTaskbar: true, HideTaskViewButton: true, DisableBingResults: true}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 3 {
		t.Fatalf("expected exactly 3 Deployment commands, got %+v", deployment)
	}
	var joined string
	for _, c := range deployment.RunSynchronousCommand {
		joined += c.Path + "\n"
	}
	for _, want := range []string{"TaskbarAl", "ShowTaskViewButton", "DisableSearchBoxSuggestions"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("commands = %q, want it to contain %q", joined, want)
		}
	}
	if strings.Count(joined, `reg.exe load HKU\DefaultUser`) != 3 {
		t.Fatalf("commands = %q, want 3 independent default-user-hive mounts", joined)
	}
}

func TestBuildAnswerFileShowAllTrayIcons(t *testing.T) {
	p := baseProfile()
	p.SystemTweaks = profile.SystemTweaks{ShowAllTrayIcons: true}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	payloads := decodeAllBase64Payloads(t, cmdLine)
	var foundTaskXML, foundBootstrap bool
	for _, p := range payloads {
		if strings.Contains(p, "<Task") && strings.Contains(p, "NotifyIconSettings") {
			foundTaskXML = true
		}
		if strings.Contains(p, "ShowAllTrayIcons") && strings.Contains(p, "OSVersion.Version.Build") && strings.Contains(p, "EnableAutoTray") {
			foundBootstrap = true
		}
	}
	if !foundTaskXML {
		t.Fatalf("payloads = %v, want the ShowAllTrayIcons scheduled task XML embedded", payloads)
	}
	if !foundBootstrap {
		t.Fatalf("payloads = %v, want the OS-version-branching bootstrap script embedded", payloads)
	}
}

func TestBuildAnswerFileTaskbarSearchModeHide(t *testing.T) {
	p := baseProfile()
	p.TaskbarSearch = profile.TaskbarSearchModeHide

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	if !strings.Contains(cmdLine, "UnattendTaskbarSearch") {
		t.Fatalf("command = %q, want a RunOnce entry for future accounts", cmdLine)
	}
	decoded := decodeBase64Payload(t, cmdLine)
	if !strings.Contains(decoded, "SearchboxTaskbarMode") || !strings.Contains(decoded, "-Value 0") {
		t.Fatalf("decoded script = %q, want SearchboxTaskbarMode set to 0 (hide)", decoded)
	}
}

func TestBuildAnswerFileTaskbarSearchDefaultAddsNothing(t *testing.T) {
	p := baseProfile()
	p.TaskbarSearch = ""

	doc := buildAndParseAccounts(t, p)
	if deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment"); deployment != nil {
		t.Fatalf("expected no Microsoft-Windows-Deployment component, got %+v", deployment)
	}
}

func TestBuildAnswerFileStartPinsCustom(t *testing.T) {
	p := baseProfile()
	pins := `{"pinnedList":[{"desktopAppId":"MSEdge"}]}`
	p.StartPins = profile.StartPinsSettings{Mode: profile.StartPinsModeCustom, JSON: &pins}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	decoded := decodeBase64Payload(t, deployment.RunSynchronousCommand[0].Path)
	if !strings.Contains(decoded, "ConfigureStartPins") || !strings.Contains(decoded, "MSEdge") {
		t.Fatalf("decoded script = %q, want the ConfigureStartPins policy with the custom JSON embedded", decoded)
	}
}

func TestBuildAnswerFileStartPinsEmpty(t *testing.T) {
	p := baseProfile()
	p.StartPins = profile.StartPinsSettings{Mode: profile.StartPinsModeEmpty}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	decoded := decodeBase64Payload(t, deployment.RunSynchronousCommand[0].Path)
	if !strings.Contains(decoded, `pinnedList`) {
		t.Fatalf("decoded script = %q, want an empty pinnedList", decoded)
	}
}

func TestValidateProfileRejectsMalformedStartPinsJSON(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"start_pins": {"mode": "custom", "json": "not json"}
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for malformed start_pins.json")
	}
}

func TestBuildAnswerFileStartTilesCustom(t *testing.T) {
	p := baseProfile()
	xml := `<LayoutModificationTemplate Version='1' xmlns='http://schemas.microsoft.com/Start/2014/LayoutModification'></LayoutModificationTemplate>`
	p.StartTiles = profile.StartTilesSettings{Mode: profile.StartTilesModeCustom, XML: &xml}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	if !strings.Contains(cmdLine, `C:\Users\Default\AppData\Local\Microsoft\Windows\Shell\LayoutModification.xml`) {
		t.Fatalf("command = %q, want the LayoutModification.xml path", cmdLine)
	}
	decoded := decodeBase64Payload(t, cmdLine)
	if !strings.Contains(decoded, "LayoutModificationTemplate") {
		t.Fatalf("decoded content = %q, want the custom XML embedded", decoded)
	}
}

func TestValidateProfileRejectsMalformedStartTilesXML(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"start_tiles": {"mode": "custom", "xml": "<Unclosed>"}
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for malformed start_tiles.xml")
	}
}
