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

// bypassWin11Command sets the well-known LabConfig registry values that make
// Windows Setup skip the TPM/Secure Boot/RAM hardware checks. It must run in
// the windowsPE pass, before Setup evaluates hardware requirements.
const bypassWin11Command = `cmd.exe /c reg add "HKLM\SYSTEM\Setup\LabConfig" /v BypassTPMCheck /t REG_DWORD /d 1 /f && reg add "HKLM\SYSTEM\Setup\LabConfig" /v BypassSecureBootCheck /t REG_DWORD /d 1 /f && reg add "HKLM\SYSTEM\Setup\LabConfig" /v BypassRAMCheck /t REG_DWORD /d 1 /f`

// runSynchronousCommand is one command run in order during a pass. Shared by
// every component that exposes a RunSynchronous list.
type runSynchronousCommand struct {
	Order int    `xml:"Order"`
	Path  string `xml:"Path"`
}

type runSynchronous struct {
	RunSynchronousCommand []runSynchronousCommand `xml:"RunSynchronousCommand"`
}

// Setup is the Microsoft-Windows-Setup component, windowsPE pass, carrying
// the product key and the Windows 11 hardware-check bypass.
type Setup struct {
	XMLName xml.Name `xml:"component"`
	Name    string   `xml:"name,attr"`
	standardAttrs
	UserData       *UserData       `xml:"UserData,omitempty"`
	RunSynchronous *runSynchronous `xml:"RunSynchronous,omitempty"`
}

// UserData wraps the product key entry.
type UserData struct {
	ProductKey ProductKey `xml:"ProductKey"`
}

// ProductKey is the license key applied during setup.
type ProductKey struct {
	Key string `xml:"Key"`
}

// NewSetup builds the Microsoft-Windows-Setup component for edition and the
// Windows 11 bypass flag. It returns nil when there is neither a product key
// nor a bypass to apply: an empty component is not emitted.
func NewSetup(edition profile.EditionSettings, bypassWin11Requirements bool) *Setup {
	key := resolveProductKey(edition)

	var runSync *runSynchronous
	if bypassWin11Requirements {
		runSync = &runSynchronous{RunSynchronousCommand: []runSynchronousCommand{
			{Order: 1, Path: bypassWin11Command},
		}}
	}

	if key == "" && runSync == nil {
		return nil
	}

	s := &Setup{
		Name:           "Microsoft-Windows-Setup",
		standardAttrs:  newStandardAttrs(),
		RunSynchronous: runSync,
	}
	if key != "" {
		s.UserData = &UserData{ProductKey: ProductKey{Key: key}}
	}
	return s
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
