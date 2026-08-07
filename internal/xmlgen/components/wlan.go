package components

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// There is no unattend component that accepts a Wi-Fi SSID/password
// directly: Microsoft-Windows-Wlansvc only configures WLAN driver/antenna
// capabilities (NumAntennaConnected and similar), not a network profile.
// The supported mechanism is the one netsh itself uses: write a real netsh
// WLAN profile XML to disk, then `netsh wlan add profile` it. The WLAN
// service is only guaranteed to be running post-boot, so this has to run
// from FirstLogonCommands, not RunSynchronous in specialize/windowsPE.

// WifiFirstLogonCommand returns the single command line that provisions
// wifi as a real Windows WLAN profile. It returns "" when wifi is nil, so
// callers can skip adding a FirstLogonCommands entry entirely.
func WifiFirstLogonCommand(wifi *profile.WifiSettings) string {
	if wifi == nil {
		return ""
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(wlanProfileXML(wifi)))
	return fmt.Sprintf(
		`powershell -NoProfile -Command "[IO.File]::WriteAllBytes('$env:TEMP\wifi-profile.xml',[Convert]::FromBase64String('%s'));netsh wlan add profile filename=\"$env:TEMP\wifi-profile.xml\" user=all"`,
		encoded,
	)
}

// wlanProfileXML builds a standalone netsh WLAN profile document (the same
// format `netsh wlan export profile` produces). It is never part of the
// unattend document tree itself; it travels base64-encoded inside the
// FirstLogonCommands command line above.
func wlanProfileXML(wifi *profile.WifiSettings) string {
	auth, enc := wifiAuthEncryption(wifi.Authentication)
	nonBroadcast := "false"
	if wifi.ConnectHidden {
		nonBroadcast = "true"
	}
	sharedKey := ""
	if wifi.Authentication != profile.WifiOpen && wifi.Password != nil {
		sharedKey = fmt.Sprintf(
			"<sharedKey><keyType>passPhrase</keyType><protected>false</protected><keyMaterial>%s</keyMaterial></sharedKey>",
			xmlEscape(*wifi.Password),
		)
	}
	ssid := xmlEscape(wifi.SSID)
	return fmt.Sprintf(
		`<?xml version="1.0"?>`+
			`<WLANProfile xmlns="http://www.microsoft.com/networking/WLAN/profile/v1">`+
			`<name>%s</name>`+
			`<SSIDConfig><SSID><name>%s</name></SSID><nonBroadcast>%s</nonBroadcast></SSIDConfig>`+
			`<connectionType>ESS</connectionType><connectionMode>auto</connectionMode>`+
			`<MSM><security><authEncryption><authentication>%s</authentication><encryption>%s</encryption><useOneX>false</useOneX></authEncryption>%s</security></MSM>`+
			`</WLANProfile>`,
		ssid, ssid, nonBroadcast, auth, enc, sharedKey,
	)
}

// wifiAuthEncryption maps a WifiAuthentication to the netsh WLAN profile
// authentication/encryption pair.
func wifiAuthEncryption(a profile.WifiAuthentication) (authentication, encryption string) {
	switch a {
	case profile.WifiWPA2Personal:
		return "WPA2PSK", "AES"
	case profile.WifiWPA3Personal:
		return "WPA3SAE", "AES"
	default:
		return "open", "none"
	}
}

// xmlEscape escapes s for use as XML text content (SSID and password can
// contain characters like & or < that would otherwise break the embedded
// WLAN profile document).
func xmlEscape(s string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(s))
	return buf.String()
}
