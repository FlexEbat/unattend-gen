package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFilePersonalizationZeroValueAddsNoCommand(t *testing.T) {
	p := baseProfile()
	p.Personalization = profile.PersonalizationSettings{}

	doc := buildAndParseAccounts(t, p)
	if shell := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment"); shell != nil {
		t.Fatalf("expected no Microsoft-Windows-Deployment component, got %+v", shell)
	}
}

func TestBuildAnswerFilePersonalizationDarkTheme(t *testing.T) {
	p := baseProfile()
	p.Personalization = profile.PersonalizationSettings{
		SystemTheme: profile.ColorThemeDark,
		AppsTheme:   profile.ColorThemeDark,
	}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmd := deployment.RunSynchronousCommand[0].Path
	if !strings.Contains(cmd, "SystemUsesLightTheme /t REG_DWORD /d 0") {
		t.Fatalf("command = %q, want SystemUsesLightTheme=0 for dark theme", cmd)
	}
	if !strings.Contains(cmd, "AppsUseLightTheme /t REG_DWORD /d 0") {
		t.Fatalf("command = %q, want AppsUseLightTheme=0 for dark theme", cmd)
	}
}

func TestBuildAnswerFilePersonalizationAccentColor(t *testing.T) {
	p := baseProfile()
	color := "1A2B3C"
	p.Personalization = profile.PersonalizationSettings{AccentColor: &color}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmd := deployment.RunSynchronousCommand[0].Path
	// RRGGBB 1A2B3C packs to AABBGGRR = FF3C2B1A.
	if !strings.Contains(cmd, "AccentColor /t REG_DWORD /d 0xFF3C2B1A") {
		t.Fatalf("command = %q, want AccentColor packed as 0xFF3C2B1A", cmd)
	}
	if !strings.Contains(cmd, "ColorizationColor /t REG_DWORD /d 0xFF3C2B1A") {
		t.Fatalf("command = %q, want ColorizationColor packed as 0xFF3C2B1A", cmd)
	}
}

func TestValidateProfileRejectsMalformedAccentColor(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"personalization": {"accent_color": "not-a-color"}
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for a malformed accent_color")
	}
}

func TestValidateProfileAcceptsValidAccentColor(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"personalization": {"accent_color": "ff8800"}
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) != 0 {
		t.Fatalf("expected a valid accent_color to validate, got errors: %v", result.Errors)
	}
}

func TestBuildAnswerFileSolidColorWallpaper(t *testing.T) {
	p := baseProfile()
	color := "0078D4"
	p.Personalization = profile.PersonalizationSettings{SolidColorWallpaper: &color}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmd := deployment.RunSynchronousCommand[0].Path
	// 0078D4 -> R=00 G=78(120) B=D4(212)
	if !strings.Contains(cmd, `Background /t REG_SZ /d \"0 120 212\"`) {
		t.Fatalf("command = %q, want Background set to decimal 0 120 212", cmd)
	}
	if !strings.Contains(cmd, `Wallpaper /t REG_SZ /d \"\"`) {
		t.Fatalf("command = %q, want the image Wallpaper value cleared", cmd)
	}
}

func TestValidateProfileRejectsMalformedSolidColorWallpaper(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"personalization": {"solid_color_wallpaper": "nope"}
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for a malformed solid_color_wallpaper")
	}
}
