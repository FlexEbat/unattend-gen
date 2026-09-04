package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFileRemoveAppsSlice17Additions(t *testing.T) {
	p := baseProfile()
	p.RemoveApps = []profile.RemovableApp{
		profile.AppBingSearch, profile.AppDevHome, profile.AppGameAssist,
		profile.AppStore, profile.AppNotepad, profile.AppOutlook,
		profile.AppPaint, profile.AppWallet, profile.AppMediaPlayer, profile.AppTerminal,
	}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 1 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 1", len(lines))
	}
	cmd := lines[0]
	for _, want := range []string{
		"'BingSearch'", "'DevHome'", "'Edge.GameAssist'", "'WindowsStore'",
		"'StorePurchaseApp'", "'WindowsNotepad'", "'OutlookForWindows'",
		"'Microsoft.Paint'", "'Wallet'", "'ZuneMusic'", "'WindowsTerminal'",
	} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("command = %q, want it to contain %q", cmd, want)
		}
	}
}

func TestBuildAnswerFilePaintPatternDoesNotCollideWithPaint3D(t *testing.T) {
	p := baseProfile()
	p.RemoveApps = []profile.RemovableApp{profile.AppPaint}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 1 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 1", len(lines))
	}
	if strings.Contains(lines[0], "'Paint3D'") {
		t.Fatalf("command = %q, selecting Paint must not also select Paint3D", lines[0])
	}
}

func TestBuildAnswerFileRemoveOneDrive(t *testing.T) {
	p := baseProfile()
	p.RemoveApps = []profile.RemovableApp{profile.AppOneDrive}

	// OneDrive isn't an Appx package, so it must NOT add a FirstLogonCommands
	// entry via the normal Remove-AppxProvisionedPackage path.
	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 0 {
		t.Fatalf("expected no FirstLogonCommands entries for OneDrive, got %v", lines)
	}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 2 {
		t.Fatalf("expected exactly 2 Deployment commands (file cleanup + default-user hive), got %+v", deployment)
	}
	var joined string
	for _, c := range deployment.RunSynchronousCommand {
		joined += c.Path + "\n"
	}
	if decoded := decodeBase64Payload(t, joined); !strings.Contains(decoded, "OneDriveSetup.exe") {
		t.Fatalf("decoded script = %q, want the OneDriveSetup.exe cleanup", decoded)
	}
	if !strings.Contains(joined, `reg.exe load HKU\DefaultUser`) {
		t.Fatalf("commands = %q, want a default-user-hive mount", joined)
	}
	if !strings.Contains(joined, "OneDriveSetup") || !strings.Contains(joined, "reg.exe delete") {
		t.Fatalf("commands = %q, want the OneDriveSetup run-key removed", joined)
	}
}

func TestBuildAnswerFileRemoveFeaturesSlice17Additions(t *testing.T) {
	p := baseProfile()
	p.RemoveFeatures = []profile.RemovableFeature{
		profile.FeatureWindowsHello, profile.FeatureMathInputPanel,
		profile.FeatureOneSync, profile.FeatureStepsRecorder,
	}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 1 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 1", len(lines))
	}
	cmd := lines[0]
	for _, want := range []string{
		"'Hello.Face.18967'", "'Hello.Face.Migration.18967'", "'Hello.Face.20134'",
		"'MathRecognizer'", "'OneCoreUAP.OneSync'", "'App.StepsRecorder'",
	} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("command = %q, want it to contain %q", cmd, want)
		}
	}
	if !strings.Contains(cmd, "Remove-WindowsCapability") {
		t.Fatalf("command = %q, want it to run Remove-WindowsCapability", cmd)
	}
}

func TestBuildAnswerFileRemoveOptionalFeaturesEmptyAddsNoCommand(t *testing.T) {
	p := baseProfile()
	p.RemoveOptionalFeatures = nil

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 0 {
		t.Fatalf("expected no FirstLogonCommands entries, got %v", lines)
	}
}

func TestBuildAnswerFileRemoveOptionalFeatures(t *testing.T) {
	p := baseProfile()
	p.RemoveOptionalFeatures = []profile.RemovableOptionalFeature{
		profile.OptionalFeatureRecall,
		profile.OptionalFeatureMediaFeatures,
		profile.OptionalFeatureRemoteDesktopClient,
	}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 1 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 1", len(lines))
	}
	cmd := lines[0]
	if !strings.Contains(cmd, "Disable-WindowsOptionalFeature") {
		t.Fatalf("command = %q, want it to run Disable-WindowsOptionalFeature", cmd)
	}
	for _, want := range []string{"'Recall'", "'MediaPlayback'", "'Microsoft-RemoteDesktopConnection'"} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("command = %q, want it to contain %q", cmd, want)
		}
	}
}

func TestBuildAnswerFileAppsFeaturesOptionalFeaturesAreSeparateOrderedCommands(t *testing.T) {
	p := baseProfile()
	p.RemoveApps = []profile.RemovableApp{profile.AppTeams}
	p.RemoveFeatures = []profile.RemovableFeature{profile.FeatureOpenSSHClient}
	p.RemoveOptionalFeatures = []profile.RemovableOptionalFeature{profile.OptionalFeatureRecall}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 3 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 3 (apps + features + optional features)", len(lines))
	}
	if !strings.Contains(lines[0], "Remove-AppxProvisionedPackage") {
		t.Fatalf("first command = %q, want app removal first", lines[0])
	}
	if !strings.Contains(lines[1], "Remove-WindowsCapability") {
		t.Fatalf("second command = %q, want capability removal second", lines[1])
	}
	if !strings.Contains(lines[2], "Disable-WindowsOptionalFeature") {
		t.Fatalf("third command = %q, want optional-feature removal third", lines[2])
	}
}

func TestValidateProfileRejectsUnknownRemoveOptionalFeature(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"remove_optional_features": ["NotReal"]
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for an unknown optional-feature name")
	}
}

func TestValidateProfileAcceptsKnownRemoveOptionalFeatures(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"remove_optional_features": ["Recall", "MediaFeatures", "RemoteDesktopClient"]
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) != 0 {
		t.Fatalf("expected known optional-feature names to validate, got errors: %v", result.Errors)
	}
}
