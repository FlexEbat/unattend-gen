package xmlgen

import (
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFileKeepSensitiveFilesFalseDeletes(t *testing.T) {
	p := baseProfile()
	p.KeepSensitiveFiles = false // production default; baseProfile() overrides to true for other tests' sake

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 1 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 1", len(lines))
	}
	decoded := decodeBase64Payload(t, lines[0])
	for _, want := range []string{"unattend.xml", "unattend-original.xml", "wifi-profile.xml"} {
		if !strings.Contains(decoded, want) {
			t.Fatalf("decoded script = %q, want it to reference %q", decoded, want)
		}
	}
}

func TestBuildAnswerFileKeepSensitiveFilesTrueSkipsDeletion(t *testing.T) {
	p := baseProfile()
	p.KeepSensitiveFiles = true

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 0 {
		t.Fatalf("expected no FirstLogonCommands entries when KeepSensitiveFiles is true, got %v", lines)
	}
}

func TestBuildAnswerFileUseNarrator(t *testing.T) {
	p := baseProfile()
	p.UseNarrator = true

	doc := buildAndParseAccounts(t, p)

	// windowsPE: Microsoft-Windows-Setup/RunSynchronous should include the
	// narrator start command alongside (or instead of) BypassWin11.
	setup := findShellComponentByName(doc, "windowsPE", "Microsoft-Windows-Setup")
	if setup == nil || len(setup.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 windowsPE Setup command, got %+v", setup)
	}
	if !strings.Contains(setup.RunSynchronousCommand[0].Path, "Narrator.exe") {
		t.Fatalf("windowsPE command = %q, want it to start Narrator.exe", setup.RunSynchronousCommand[0].Path)
	}

	// specialize: Deployment should carry 2 commands (specialize launch +
	// UserOnce registration for future accounts).
	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 2 {
		t.Fatalf("expected exactly 2 Deployment commands, got %+v", deployment)
	}
	var joined string
	for _, c := range deployment.RunSynchronousCommand {
		joined += c.Path + "\n"
	}
	if !strings.Contains(joined, "Narrator.exe") {
		t.Fatalf("commands = %q, want a Narrator.exe launch", joined)
	}
	if !strings.Contains(joined, `reg.exe load HKU\DefaultUser`) {
		t.Fatalf("commands = %q, want a default-user-hive mount for the UserOnce entry", joined)
	}
	if !strings.Contains(joined, "UnattendNarrator") {
		t.Fatalf("commands = %q, want a RunOnce entry for future accounts", joined)
	}
}

func TestBuildAnswerFileUseNarratorOffAddsNothing(t *testing.T) {
	p := baseProfile()
	p.UseNarrator = false

	doc := buildAndParseAccounts(t, p)
	if setup := findShellComponentByName(doc, "windowsPE", "Microsoft-Windows-Setup"); setup != nil {
		t.Fatalf("expected no windowsPE Setup component, got %+v", setup)
	}
}

func TestBuildAnswerFileComputerNameScript(t *testing.T) {
	p := baseProfile()
	p.ComputerName = nil
	script := "Get-Random -Minimum 1 -Maximum 100 | ForEach-Object { \"PC-$_\" }"
	p.ComputerNameScript = &script

	doc := buildAndParseAccounts(t, p)
	specialize := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Shell-Setup")
	if specialize == nil {
		t.Fatal("expected a specialize Shell-Setup component")
	}
	if specialize.ComputerName != "TEMPNAME" {
		t.Fatalf("ComputerName = %q, want the TEMPNAME placeholder", specialize.ComputerName)
	}

	deployment := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Deployment")
	if deployment == nil || len(deployment.RunSynchronousCommand) != 1 {
		t.Fatalf("expected exactly 1 Deployment command, got %+v", deployment)
	}
	cmdLine := deployment.RunSynchronousCommand[0].Path
	if !strings.Contains(cmdLine, "ComputerName.txt") {
		t.Fatalf("command = %q, want it to reference ComputerName.txt", cmdLine)
	}
	if got := decodeBase64Payload(t, cmdLine); !strings.Contains(got, "Get-Random") {
		t.Fatalf("decoded command = %q, want the user-supplied getter script embedded", got)
	}
}

func TestValidateProfileRejectsComputerNameAndScriptTogether(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"computer_name": "MYPC",
		"computer_name_script": "Write-Output 'x'"
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error when computer_name and computer_name_script are both set")
	}
}

func TestBuildAnswerFileObscurePasswords(t *testing.T) {
	p := baseProfile()
	pwd := "Sup3rSecret!"
	p.Accounts = []profile.UserAccount{{Name: "alice", Password: &pwd, Group: profile.GroupAdministrators}}
	p.ObscurePasswords = true

	doc := buildAndParseAccounts(t, p)
	shell := findShellComponentByName(doc, "oobeSystem", "Microsoft-Windows-Shell-Setup")
	if shell == nil || len(shell.LocalAccounts) != 1 {
		t.Fatalf("expected 1 local account, got %+v", shell)
	}
	acc := shell.LocalAccounts[0]
	if acc.Password == nil {
		t.Fatal("expected a Password element")
	}
	if acc.Password.PlainText {
		t.Fatal("expected PlainText=false when ObscurePasswords is set")
	}
	if acc.Password.Value == pwd {
		t.Fatalf("expected the password value to be obscured, got the raw password back")
	}
}

func TestBuildAnswerFilePasswordsPlainByDefault(t *testing.T) {
	p := baseProfile()
	pwd := "Sup3rSecret!"
	p.Accounts = []profile.UserAccount{{Name: "alice", Password: &pwd, Group: profile.GroupAdministrators}}
	p.ObscurePasswords = false

	doc := buildAndParseAccounts(t, p)
	shell := findShellComponentByName(doc, "oobeSystem", "Microsoft-Windows-Shell-Setup")
	if shell == nil || len(shell.LocalAccounts) != 1 {
		t.Fatalf("expected 1 local account, got %+v", shell)
	}
	acc := shell.LocalAccounts[0]
	if acc.Password == nil {
		t.Fatal("expected a Password element")
	}
	if !acc.Password.PlainText {
		t.Fatal("expected PlainText=true by default (this project's existing behavior)")
	}
	if acc.Password.Value != pwd {
		t.Fatalf("Password.Value = %q, want the raw password %q", acc.Password.Value, pwd)
	}
}

func TestBuildAnswerFileWifiRawProfileXML(t *testing.T) {
	p := baseProfile()
	raw := `<?xml version="1.0"?><WLANProfile xmlns="http://www.microsoft.com/networking/WLAN/profile/v1"><name>CustomNet</name></WLANProfile>`
	p.Wifi = &profile.WifiSettings{RawProfileXML: &raw}

	lines := allFirstLogonCommandLines(t, p)
	if len(lines) != 1 {
		t.Fatalf("got %d FirstLogonCommands entries, want exactly 1", len(lines))
	}
	decoded := decodeBase64Payload(t, lines[0])
	if !strings.Contains(decoded, "CustomNet") {
		t.Fatalf("decoded WLAN profile = %q, want the raw XML embedded verbatim", decoded)
	}
}

func TestValidateProfileRejectsWifiWithoutSSIDOrRawXML(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"wifi": {"connect_hidden": false}
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error when wifi is set but has neither ssid nor raw_profile_xml")
	}
}

func TestValidateProfileAcceptsWifiRawProfileXMLWithoutSSID(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"wifi": {"raw_profile_xml": "<WLANProfile><name>X</name></WLANProfile>"}
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) != 0 {
		t.Fatalf("expected raw_profile_xml alone to validate, got errors: %v", result.Errors)
	}
}
