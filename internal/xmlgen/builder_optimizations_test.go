package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFilePreventDeviceApps(t *testing.T) {
	p := baseProfile()
	p.SystemTweaks = profile.SystemTweaks{PreventDeviceApps: true}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	if !strings.Contains(cmdLine, "PreventDeviceMetadataFromNetwork") || !strings.Contains(cmdLine, "DisableCoInstallers") {
		t.Fatalf("command = %q, want both device-metadata and co-installer keys", cmdLine)
	}
}

func TestBuildAnswerFileTurnOffSystemSounds(t *testing.T) {
	p := baseProfile()
	p.SystemTweaks = profile.SystemTweaks{TurnOffSystemSounds: true}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil {
		t.Fatal("expected a Microsoft-Windows-Deployment component")
	}
	// One simple specialize-only command (BootAnimation/EditionOverrides) plus
	// two default-user-hive commands (template scheme + live HKCU RunOnce).
	if len(deployment.RunSynchronousCommand) != 3 {
		t.Fatalf("got %d RunSynchronousCommand elements, want exactly 3", len(deployment.RunSynchronousCommand))
	}
	var joined string
	for _, c := range deployment.RunSynchronousCommand {
		joined += c.Path + "\n"
	}
	if !strings.Contains(joined, "DisableStartupSound") {
		t.Fatalf("commands = %q, want the boot-sound registry key", joined)
	}
	if !strings.Contains(joined, `reg.exe load HKU\DefaultUser`) {
		t.Fatalf("commands = %q, want a default-user-hive mount for the app-event schemes", joined)
	}
	if !strings.Contains(joined, "UnattendSystemSounds") {
		t.Fatalf("commands = %q, want a RunOnce entry for the live account scheme", joined)
	}
	if got := decodeBase64Payload(t, joined); !strings.Contains(got, "AppEvents") {
		t.Fatalf("decoded default-user script = %q, want it to touch AppEvents", got)
	}
}

func TestBuildAnswerFileDisableAppSuggestions(t *testing.T) {
	p := baseProfile()
	p.SystemTweaks = profile.SystemTweaks{DisableAppSuggestions: true}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 2 {
		t.Fatalf("expected exactly 2 Deployment commands (specialize policy + default-user hive), got %+v", deployment)
	}
	var joined string
	for _, c := range deployment.RunSynchronousCommand {
		joined += c.Path + "\n"
	}
	if !strings.Contains(joined, "DisableWindowsConsumerFeatures") {
		t.Fatalf("commands = %q, want the CloudContent policy key", joined)
	}
	if !strings.Contains(joined, `ContentDeliveryManager`) {
		t.Fatalf("commands = %q, want the ContentDeliveryManager hive key", joined)
	}
	if !strings.Contains(joined, "SubscribedContentEnabled") {
		t.Fatalf("commands = %q, want at least one of the 17 ContentDeliveryManager values", joined)
	}
}

func TestBuildAnswerFileDisablePointerPrecision(t *testing.T) {
	p := baseProfile()
	p.SystemTweaks = profile.SystemTweaks{DisablePointerPrecision: true}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	for _, want := range []string{"MouseSpeed", "MouseThreshold1", "MouseThreshold2", `Control Panel\Mouse`} {
		if !strings.Contains(cmdLine, want) {
			t.Fatalf("command = %q, want it to contain %q", cmdLine, want)
		}
	}
}

func TestBuildAnswerFilePreventAutomaticReboot(t *testing.T) {
	p := baseProfile()
	p.SystemTweaks = profile.SystemTweaks{PreventAutomaticReboot: true}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	if !strings.Contains(cmdLine, "NoAutoRebootWithLoggedOnUsers") {
		t.Fatalf("command = %q, want the no-auto-reboot policy key", cmdLine)
	}
	if !strings.Contains(cmdLine, "Register-ScheduledTask") || !strings.Contains(cmdLine, "MoveActiveHours") {
		t.Fatalf("command = %q, want it to register the MoveActiveHours scheduled task", cmdLine)
	}
	if got := decodeBase64Payload(t, cmdLine); !strings.Contains(got, "<Task") || !strings.Contains(got, "ActiveHoursStart") {
		t.Fatalf("decoded task XML = %q, want a <Task> definition touching ActiveHoursStart", got)
	}
}

func TestBuildAnswerFileDeleteHiddenJunctions(t *testing.T) {
	p := baseProfile()
	p.SystemTweaks = profile.SystemTweaks{DeleteHiddenJunctions: true}

	// FirstLogonCommands side (oobeSystem), same family as Wi-Fi/apps/features.
	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 1 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 1", len(lines))
	}
	if got := decodeBase64Payload(t, lines[0]); !strings.Contains(got, "ReparsePoint") {
		t.Fatalf("decoded FirstLogon script = %q, want it to filter on ReparsePoint", got)
	}

	// UserOnce side (specialize, default-user hive + RunOnce), for future accounts.
	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	if !strings.Contains(cmdLine, "UnattendDeleteJunctions") {
		t.Fatalf("command = %q, want a RunOnce entry for future accounts", cmdLine)
	}
	if !strings.Contains(cmdLine, `reg.exe load HKU\DefaultUser`) {
		t.Fatalf("command = %q, want a default-user-hive mount", cmdLine)
	}
}

func TestBuildAnswerFileGroupBTweaksOffAddNothing(t *testing.T) {
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
