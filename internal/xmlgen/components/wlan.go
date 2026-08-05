package components

import (
	"encoding/xml"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// WLANSVC is the Microsoft-Windows-WLANSVC component, oobeSystem pass,
// carrying a single Wi-Fi profile.
type WLANSVC struct {
	XMLName               xml.Name     `xml:"component"`
	Name                  string       `xml:"name,attr"`
	ProcessorArchitecture string       `xml:"processorArchitecture,attr"`
	PublicKeyToken        string       `xml:"publicKeyToken,attr"`
	Language              string       `xml:"language,attr"`
	VersionScope          string       `xml:"versionScope,attr"`
	WLANProfiles          wlanProfiles `xml:"WLANProfiles"`
}

type wlanProfiles struct {
	WLANProfile wlanProfile `xml:"WLANProfile"`
}

type wlanProfile struct {
	SSIDConfig ssidConfig `xml:"SSIDConfig"`
	MSM        msm        `xml:"MSM"`
}

type ssidConfig struct {
	SSID         ssidName `xml:"SSID"`
	NonBroadcast bool     `xml:"nonBroadcast"`
}

type ssidName struct {
	Name string `xml:"name"`
}

type msm struct {
	Security wlanSecurity `xml:"security"`
}

type wlanSecurity struct {
	AuthEncryption authEncryption `xml:"authEncryption"`
	SharedKey      *sharedKey     `xml:"sharedKey,omitempty"`
}

type authEncryption struct {
	Authentication string `xml:"authentication"`
	Encryption     string `xml:"encryption"`
}

type sharedKey struct {
	KeyType     string `xml:"keyType"`
	Protected   bool   `xml:"protected"`
	KeyMaterial string `xml:"keyMaterial"`
}

// NewWLANSVC builds the component from wifi. It returns nil when wifi is
// nil: Wi-Fi is configured manually or via a $WinPEDriver$ instead.
func NewWLANSVC(wifi *profile.WifiSettings) *WLANSVC {
	if wifi == nil {
		return nil
	}

	auth, enc := wifiAuthEncryption(wifi.Authentication)
	sec := wlanSecurity{AuthEncryption: authEncryption{Authentication: auth, Encryption: enc}}
	if wifi.Authentication != profile.WifiOpen && wifi.Password != nil {
		sec.SharedKey = &sharedKey{KeyType: "passPhrase", Protected: false, KeyMaterial: *wifi.Password}
	}

	return &WLANSVC{
		Name:                  "Microsoft-Windows-WLANSVC",
		ProcessorArchitecture: "amd64",
		PublicKeyToken:        "31bf3856ad364e35",
		Language:              "neutral",
		VersionScope:          "nonSxS",
		WLANProfiles: wlanProfiles{WLANProfile: wlanProfile{
			SSIDConfig: ssidConfig{SSID: ssidName{Name: wifi.SSID}, NonBroadcast: wifi.ConnectHidden},
			MSM:        msm{Security: sec},
		}},
	}
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
