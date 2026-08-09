package xmlgen

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// testFirstLogonCommands captures Shell-Setup's FirstLogonCommands, which is
// how Wi-Fi provisioning travels: there is no unattend component that takes
// an SSID/password directly (Microsoft-Windows-Wlansvc only configures WLAN
// driver/antenna capabilities), so the real mechanism is a command that
// writes a netsh WLAN profile to disk and runs `netsh wlan add profile`.
type testFirstLogonCommands struct {
	SynchronousCommand []struct {
		CommandLine string `xml:"CommandLine"`
		Order       int    `xml:"Order"`
	} `xml:"SynchronousCommand"`
}

type testWifiComponent struct {
	Name               string                  `xml:"name,attr"`
	FirstLogonCommands *testFirstLogonCommands `xml:"FirstLogonCommands"`
}

type testWifiPass struct {
	Pass       string              `xml:"pass,attr"`
	Components []testWifiComponent `xml:"component"`
}

type testWifiDoc struct {
	XMLName  xml.Name       `xml:"unattend"`
	Settings []testWifiPass `xml:"settings"`
}

// wifiCommandLine returns the CommandLine of the oobeSystem Shell-Setup's
// first FirstLogonCommands entry, or "" if there is none.
func wifiCommandLine(t *testing.T, p *profile.Profile) string {
	t.Helper()
	xmlStr, err := BuildAnswerFile(p)
	if err != nil {
		t.Fatalf("BuildAnswerFile: %v", err)
	}
	var doc testWifiDoc
	if err := xml.Unmarshal([]byte(xmlStr), &doc); err != nil {
		t.Fatalf("Unmarshal result: %v\nxml:\n%s", err, xmlStr)
	}
	for _, s := range doc.Settings {
		if s.Pass != "oobeSystem" {
			continue
		}
		for _, c := range s.Components {
			if c.Name != "Microsoft-Windows-Shell-Setup" || c.FirstLogonCommands == nil {
				continue
			}
			if len(c.FirstLogonCommands.SynchronousCommand) == 0 {
				return ""
			}
			return c.FirstLogonCommands.SynchronousCommand[0].CommandLine
		}
	}
	return ""
}

// allFirstLogonCommandLines returns every FirstLogonCommands CommandLine in
// the oobeSystem Shell-Setup component, in Order.
func allFirstLogonCommandLines(t *testing.T, p *profile.Profile) []string {
	t.Helper()
	xmlStr, err := BuildAnswerFile(p)
	if err != nil {
		t.Fatalf("BuildAnswerFile: %v", err)
	}
	var doc testWifiDoc
	if err := xml.Unmarshal([]byte(xmlStr), &doc); err != nil {
		t.Fatalf("Unmarshal result: %v\nxml:\n%s", err, xmlStr)
	}
	for _, s := range doc.Settings {
		if s.Pass != "oobeSystem" {
			continue
		}
		for _, c := range s.Components {
			if c.Name != "Microsoft-Windows-Shell-Setup" || c.FirstLogonCommands == nil {
				continue
			}
			lines := make([]string, len(c.FirstLogonCommands.SynchronousCommand))
			for i, sc := range c.FirstLogonCommands.SynchronousCommand {
				lines[i] = sc.CommandLine
			}
			return lines
		}
	}
	return nil
}

// decodedWlanProfile extracts and base64-decodes the WLAN profile XML
// embedded in a `netsh wlan add profile` command line.
func decodedWlanProfile(t *testing.T, commandLine string) string {
	t.Helper()
	start := strings.Index(commandLine, "FromBase64String('") + len("FromBase64String('")
	end := strings.Index(commandLine[start:], "'")
	if start < len("FromBase64String('") || end < 0 {
		t.Fatalf("could not find base64 payload in command line: %q", commandLine)
	}
	encoded := commandLine[start : start+end]
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	return string(decoded)
}

func TestBuildAnswerFileWifiNilOmitsFirstLogonCommand(t *testing.T) {
	p := baseProfile()
	p.Wifi = nil

	if cmd := wifiCommandLine(t, p); cmd != "" {
		t.Fatalf("expected no Wi-Fi FirstLogonCommands entry, got %q", cmd)
	}
}

func TestBuildAnswerFileWifiOpenOmitsPassword(t *testing.T) {
	p := baseProfile()
	p.Wifi = &profile.WifiSettings{SSID: "MyNetwork", Authentication: profile.WifiOpen}

	cmd := wifiCommandLine(t, p)
	if cmd == "" {
		t.Fatal("expected a Wi-Fi FirstLogonCommands entry")
	}
	if !strings.Contains(cmd, "netsh wlan add profile") {
		t.Fatalf("command line = %q, want it to run netsh wlan add profile", cmd)
	}

	wlanXML := decodedWlanProfile(t, cmd)
	if !strings.Contains(wlanXML, "<name>MyNetwork</name>") {
		t.Fatalf("decoded WLAN profile = %q, want SSID MyNetwork", wlanXML)
	}
	if strings.Contains(wlanXML, "<sharedKey>") {
		t.Fatalf("decoded WLAN profile = %q, want no sharedKey for Open auth", wlanXML)
	}
}

func TestValidateProfileRejectsShortWifiPassword(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"wifi": {"ssid": "MyNetwork", "authentication": "WPA2Personal", "password": "short"}
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for a Wi-Fi password shorter than 8 characters")
	}
}

func TestBuildAnswerFileWifiConnectHidden(t *testing.T) {
	for _, hidden := range []bool{true, false} {
		pwd := "supersecret"
		p := baseProfile()
		p.Wifi = &profile.WifiSettings{
			SSID:           "MyNetwork",
			Authentication: profile.WifiWPA2Personal,
			Password:       &pwd,
			ConnectHidden:  hidden,
		}

		// Sanity: this profile must also pass the real validator, matching
		// how BuildAnswerFile is actually used.
		data, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("marshal profile: %v", err)
		}
		if result := profile.ValidateProfile(data); len(result.Errors) != 0 {
			t.Fatalf("profile should validate, got errors: %v", result.Errors)
		}

		cmd := wifiCommandLine(t, p)
		if cmd == "" {
			t.Fatal("expected a Wi-Fi FirstLogonCommands entry")
		}
		wlanXML := decodedWlanProfile(t, cmd)

		wantBroadcast := "<nonBroadcast>false</nonBroadcast>"
		if hidden {
			wantBroadcast = "<nonBroadcast>true</nonBroadcast>"
		}
		if !strings.Contains(wlanXML, wantBroadcast) {
			t.Fatalf("ConnectHidden=%v: decoded WLAN profile = %q, want %q", hidden, wlanXML, wantBroadcast)
		}
		if !strings.Contains(wlanXML, "<authentication>WPA2PSK</authentication>") {
			t.Fatalf("decoded WLAN profile = %q, want authentication WPA2PSK", wlanXML)
		}
		if !strings.Contains(wlanXML, "<keyMaterial>"+pwd+"</keyMaterial>") {
			t.Fatalf("decoded WLAN profile = %q, want keyMaterial %q", wlanXML, pwd)
		}
	}
}
