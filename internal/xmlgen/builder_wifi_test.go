package xmlgen

import (
	"encoding/json"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFileWifiNilOmitsComponent(t *testing.T) {
	p := baseProfile()
	p.Wifi = nil

	doc := buildAndParseAccounts(t, p)
	if wlan := findShellComponentByName(doc, "oobeSystem", "Microsoft-Windows-WLANSVC"); wlan != nil {
		t.Fatalf("expected no WLANSVC component, got %+v", wlan)
	}
}

func TestBuildAnswerFileWifiOpenOmitsPassword(t *testing.T) {
	p := baseProfile()
	p.Wifi = &profile.WifiSettings{SSID: "MyNetwork", Authentication: profile.WifiOpen}

	doc := buildAndParseAccounts(t, p)
	wlan := findShellComponentByName(doc, "oobeSystem", "Microsoft-Windows-WLANSVC")
	if wlan == nil {
		t.Fatal("expected a WLANSVC component")
	}
	if wlan.SSID != "MyNetwork" {
		t.Fatalf("SSID = %q, want %q", wlan.SSID, "MyNetwork")
	}
	if wlan.KeyMaterial != nil {
		t.Fatalf("expected no password element for Open auth, got %q", *wlan.KeyMaterial)
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

		doc := buildAndParseAccounts(t, p)
		wlan := findShellComponentByName(doc, "oobeSystem", "Microsoft-Windows-WLANSVC")
		if wlan == nil {
			t.Fatal("expected a WLANSVC component")
		}
		if wlan.NonBroadcast != hidden {
			t.Fatalf("ConnectHidden=%v: NonBroadcast = %v, want %v", hidden, wlan.NonBroadcast, hidden)
		}
		if wlan.Authentication != "WPA2PSK" {
			t.Fatalf("authentication = %q, want WPA2PSK", wlan.Authentication)
		}
		if wlan.KeyMaterial == nil || *wlan.KeyMaterial != pwd {
			t.Fatalf("keyMaterial = %v, want %q", wlan.KeyMaterial, pwd)
		}
	}
}
