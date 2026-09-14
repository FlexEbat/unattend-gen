package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFileVisualEffectsBestAppearance(t *testing.T) {
	p := baseProfile()
	p.VisualEffects = profile.VisualEffectsSettings{Mode: profile.VisualEffectsModeBestAppearance}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 2 {
		t.Fatalf("expected exactly 2 Deployment commands (HKLM template + UserOnce), got %+v", deployment)
	}
	var joined string
	for _, c := range deployment.RunSynchronousCommand {
		joined += c.Path + "\n"
	}
	for _, effect := range profile.VisualEffects {
		key := `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\VisualEffects\` + string(effect)
		if !strings.Contains(joined, key+`\" /v DefaultValue /t REG_DWORD /d 1 /f`) {
			t.Fatalf("commands = %q, want %s set to 1 (best appearance = all on)", joined, effect)
		}
	}
	if !strings.Contains(joined, `reg.exe load HKU\DefaultUser`) {
		t.Fatalf("commands = %q, want a default-user-hive mount for the UserOnce entry", joined)
	}
	if !strings.Contains(joined, "UnattendVisualEffects") {
		t.Fatalf("commands = %q, want a RunOnce entry for the live VisualFXSetting", joined)
	}
}

func TestBuildAnswerFileVisualEffectsBestPerformance(t *testing.T) {
	p := baseProfile()
	p.VisualEffects = profile.VisualEffectsSettings{Mode: profile.VisualEffectsModeBestPerformance}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 2 {
		t.Fatalf("expected exactly 2 Deployment commands, got %+v", deployment)
	}
	var joined string
	for _, c := range deployment.RunSynchronousCommand {
		joined += c.Path + "\n"
	}
	for _, effect := range profile.VisualEffects {
		key := `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\VisualEffects\` + string(effect)
		if !strings.Contains(joined, key+`\" /v DefaultValue /t REG_DWORD /d 0 /f`) {
			t.Fatalf("commands = %q, want %s set to 0 (best performance = all off)", joined, effect)
		}
	}
	decoded := decodeAllBase64Payloads(t, joined)
	found := false
	for _, p := range decoded {
		if strings.Contains(p, "VisualFXSetting") && strings.Contains(p, "-Value 2") {
			found = true
		}
	}
	if !found {
		t.Fatalf("decoded payloads = %v, want VisualFXSetting set to 2 (best performance)", decoded)
	}
}

func TestBuildAnswerFileVisualEffectsCustom(t *testing.T) {
	p := baseProfile()
	p.VisualEffects = profile.VisualEffectsSettings{
		Mode: profile.VisualEffectsModeCustom,
		Custom: map[profile.VisualEffect]bool{
			profile.EffectDragFullWindows: true,
			profile.EffectFontSmoothing:   true,
			profile.EffectDropShadow:      false,
		},
	}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 2 {
		t.Fatalf("expected exactly 2 Deployment commands, got %+v", deployment)
	}
	var joined string
	for _, c := range deployment.RunSynchronousCommand {
		joined += c.Path + "\n"
	}
	if !strings.Contains(joined, `VisualEffects\DragFullWindows\" /v DefaultValue /t REG_DWORD /d 1 /f`) {
		t.Fatalf("commands = %q, want DragFullWindows=1", joined)
	}
	if !strings.Contains(joined, `VisualEffects\DropShadow\" /v DefaultValue /t REG_DWORD /d 0 /f`) {
		t.Fatalf("commands = %q, want DropShadow=0", joined)
	}
	// Unlisted effects (e.g. CursorShadow) must not appear at all.
	if strings.Contains(joined, `VisualEffects\CursorShadow\"`) {
		t.Fatalf("commands = %q, want unlisted effects untouched", joined)
	}
}

func TestBuildAnswerFileVisualEffectsDefaultAddsNothing(t *testing.T) {
	p := baseProfile()
	p.VisualEffects = profile.VisualEffectsSettings{}

	doc := buildAndParseAccounts(t, p)
	if deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment"); deployment != nil {
		t.Fatalf("expected no Microsoft-Windows-Deployment component, got %+v", deployment)
	}
}

func TestValidateProfileRejectsUnknownVisualEffect(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"visual_effects": {"mode": "custom", "custom": {"NotReal": true}}
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for an unknown visual effect name")
	}
}
