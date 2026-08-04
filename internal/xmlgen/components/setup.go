package components

import (
	"encoding/xml"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// genericKeys holds the Microsoft-published generic (KMS client setup) keys
// used for EditionModeGenericKey, one per WindowsEdition.
var genericKeys = map[profile.WindowsEdition]string{
	profile.EditionHome:       "YTMG3-N6DKC-DKB77-7M9GH-8HVX7",
	profile.EditionPro:        "VK7JG-NPHTM-C97JM-9MPGT-3V66T",
	profile.EditionEducation:  "NW6C2-QMPVW-D7KKK-3GKT6-VCFB2",
	profile.EditionEnterprise: "NPPR9-FWDCX-D2C8J-H872K-2YT43",
}

// Setup is the Microsoft-Windows-Setup component, windowsPE pass, carrying
// the product key.
type Setup struct {
	XMLName               xml.Name `xml:"component"`
	Name                  string   `xml:"name,attr"`
	ProcessorArchitecture string   `xml:"processorArchitecture,attr"`
	PublicKeyToken        string   `xml:"publicKeyToken,attr"`
	Language              string   `xml:"language,attr"`
	VersionScope          string   `xml:"versionScope,attr"`
	UserData              UserData `xml:"UserData"`
}

// UserData wraps the product key entry.
type UserData struct {
	ProductKey ProductKey `xml:"ProductKey"`
}

// ProductKey is the license key applied during setup.
type ProductKey struct {
	Key string `xml:"Key"`
}

// NewSetup builds the Microsoft-Windows-Setup component for edition. It
// returns nil when Mode is EditionModeInteractive, since there is no product
// key to carry and an empty component is not emitted.
func NewSetup(edition profile.EditionSettings) *Setup {
	key := resolveProductKey(edition)
	if key == "" {
		return nil
	}
	return &Setup{
		Name:                  "Microsoft-Windows-Setup",
		ProcessorArchitecture: "amd64",
		PublicKeyToken:        "31bf3856ad364e35",
		Language:              "neutral",
		VersionScope:          "nonSxS",
		UserData:              UserData{ProductKey: ProductKey{Key: key}},
	}
}

func resolveProductKey(edition profile.EditionSettings) string {
	switch edition.Mode {
	case profile.EditionModeGenericKey:
		if edition.Edition == nil {
			return ""
		}
		return genericKeys[*edition.Edition]
	case profile.EditionModeCustomKey:
		if edition.ProductKey == nil {
			return ""
		}
		return *edition.ProductKey
	default:
		return ""
	}
}
