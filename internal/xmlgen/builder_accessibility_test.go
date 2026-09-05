package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFileStickyKeysDisabled(t *testing.T) {
	p := baseProfile()
	p.StickyKeys = profile.StickyKeysSettings{Mode: profile.StickyKeysModeDisabled}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 2 {
		t.Fatalf("expected exactly 2 Deployment commands (default-user hive + .DEFAULT), got %+v", deployment)
	}
	var joined string
	for _, c := range deployment.RunSynchronousCommand {
		joined += c.Path + "\n"
	}
	// SKF_AVAILABLE(0x2) | SKF_CONFIRMHOTKEY(0x8) = 10, no HotKeyActive.
	if !strings.Contains(joined, "/d 10 /f") {
		t.Fatalf("commands = %q, want Flags value 10 (available+confirm, no hotkey)", joined)
	}
	if !strings.Contains(joined, `reg.exe load HKU\DefaultUser`) {
		t.Fatalf("commands = %q, want a default-user-hive mount", joined)
	}
	if !strings.Contains(joined, `HKU\.DEFAULT\Control Panel\Accessibility\StickyKeys`) {
		t.Fatalf("commands = %q, want a direct HKU\\.DEFAULT write", joined)
	}
}

func TestBuildAnswerFileStickyKeysCustom(t *testing.T) {
	p := baseProfile()
	p.StickyKeys = profile.StickyKeysSettings{
		Mode:  profile.StickyKeysModeCustom,
		Flags: []profile.StickyKeysFlag{profile.StickyKeysHotKeyActive, profile.StickyKeysAudibleFeedback},
	}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 2 {
		t.Fatalf("expected exactly 2 Deployment commands, got %+v", deployment)
	}
	// 0x2 | 0x8 | 0x4 (HotKeyActive) | 0x40 (AudibleFeedback) = 0x4E = 78.
	var joined string
	for _, c := range deployment.RunSynchronousCommand {
		joined += c.Path + "\n"
	}
	if !strings.Contains(joined, "/d 78 /f") {
		t.Fatalf("commands = %q, want Flags value 78", joined)
	}
}

func TestBuildAnswerFileStickyKeysDefaultAddsNothing(t *testing.T) {
	p := baseProfile()
	p.StickyKeys = profile.StickyKeysSettings{}

	doc := buildAndParseAccounts(t, p)
	if deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment"); deployment != nil {
		t.Fatalf("expected no Microsoft-Windows-Deployment component, got %+v", deployment)
	}
}

func TestBuildAnswerFileLockKeysInitialState(t *testing.T) {
	p := baseProfile()
	p.LockKeys = &profile.LockKeySettings{
		CapsLock:   profile.LockKeySetting{Initial: profile.LockKeyOn, Behavior: profile.LockKeyToggle},
		NumLock:    profile.LockKeySetting{Initial: profile.LockKeyOn, Behavior: profile.LockKeyToggle},
		ScrollLock: profile.LockKeySetting{Initial: profile.LockKeyOff, Behavior: profile.LockKeyToggle},
	}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command (indicators only, no scancode map), got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	// Caps(1) + Num(2) = 3.
	if !strings.Contains(cmdLine, "InitialKeyboardIndicators /t REG_SZ /d 3 /f") {
		t.Fatalf("command = %q, want InitialKeyboardIndicators = 3", cmdLine)
	}
	if !strings.Contains(cmdLine, `HKU\.DEFAULT\Control Panel\Keyboard`) {
		t.Fatalf("command = %q, want a direct HKU\\.DEFAULT write", cmdLine)
	}
	if !strings.Contains(cmdLine, `reg.exe load HKU\DefaultUser`) {
		t.Fatalf("command = %q, want a default-user-hive mount", cmdLine)
	}
}

func TestBuildAnswerFileLockKeysIgnoreBehaviorAddsScancodeMap(t *testing.T) {
	p := baseProfile()
	p.LockKeys = &profile.LockKeySettings{
		CapsLock:   profile.LockKeySetting{Initial: profile.LockKeyOff, Behavior: profile.LockKeyIgnore},
		NumLock:    profile.LockKeySetting{Initial: profile.LockKeyOff, Behavior: profile.LockKeyToggle},
		ScrollLock: profile.LockKeySetting{Initial: profile.LockKeyOff, Behavior: profile.LockKeyIgnore},
	}

	doc := buildAndParseAccounts(t, p)
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 2 {
		t.Fatalf("expected exactly 2 Deployment commands (indicators + scancode map), got %+v", deployment)
	}
	var joined string
	for _, c := range deployment.RunSynchronousCommand {
		joined += c.Path + "\n"
	}
	if !strings.Contains(joined, `"Scancode Map" /t REG_BINARY /d`) {
		t.Fatalf("commands = %q, want a Scancode Map REG_BINARY write", joined)
	}
	// Version(4)+Flags(4)+Count(4, little-endian 3)+2 entries(8)+terminator(4) = 24 bytes = 48 hex chars.
	idx := strings.Index(joined, `/t REG_BINARY /d `)
	if idx == -1 {
		t.Fatalf("commands = %q, missing REG_BINARY value", joined)
	}
	rest := joined[idx+len(`/t REG_BINARY /d `):]
	hexValue := strings.Fields(rest)[0]
	if len(hexValue) != 48 {
		t.Fatalf("scancode map hex = %q (len %d), want 48 hex chars (24 bytes: caps+scroll ignored)", hexValue, len(hexValue))
	}
	if !strings.Contains(hexValue, "00003a00") { // CapsLock scancode 0x3A, little-endian target+source
		t.Fatalf("scancode map hex = %q, want a CapsLock (0x3A) entry", hexValue)
	}
	if !strings.Contains(hexValue, "00004600") { // ScrollLock scancode 0x46
		t.Fatalf("scancode map hex = %q, want a ScrollLock (0x46) entry", hexValue)
	}
}

func TestBuildAnswerFileLockKeysNilAddsNothing(t *testing.T) {
	p := baseProfile()
	p.LockKeys = nil

	doc := buildAndParseAccounts(t, p)
	if deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment"); deployment != nil {
		t.Fatalf("expected no Microsoft-Windows-Deployment component, got %+v", deployment)
	}
}

func TestValidateProfileRejectsUnknownStickyKeysFlag(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"sticky_keys": {"mode": "custom", "flags": ["NotReal"]}
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for an unknown Sticky Keys flag")
	}
}

func TestValidateProfileRejectsInvalidLockKeyValues(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"lock_keys": {"caps_lock": {"initial": "sideways", "behavior": "toggle"}, "num_lock": {"initial": "off", "behavior": "toggle"}, "scroll_lock": {"initial": "off", "behavior": "toggle"}}
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for an invalid lock key initial value")
	}
}
