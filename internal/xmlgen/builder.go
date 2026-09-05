// Package xmlgen builds a Windows autounattend.xml answer file from a
// *profile.Profile. It never reads from disk and never imports cobra or
// bubbletea: input is a validated *profile.Profile, output is an XML string.
package xmlgen

import (
	"encoding/xml"

	"github.com/FlexEbat/unattend-gen/internal/profile"
	"github.com/FlexEbat/unattend-gen/internal/xmlgen/components"
)

const unattendNamespace = "urn:schemas-microsoft-com:unattend"

// unattendDoc is the root element of the answer file.
type unattendDoc struct {
	XMLName  xml.Name       `xml:"unattend"`
	Xmlns    string         `xml:"xmlns,attr"`
	Settings []settingsPass `xml:"settings"`
}

// settingsPass is one <settings pass="..."> block holding an ordered list of
// heterogeneous <component> elements. It implements xml.Marshaler because its
// component list mixes concrete types that only share the "component" tag.
type settingsPass struct {
	Pass       string
	Components []interface{}
}

func (s settingsPass) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	start := xml.StartElement{
		Name: xml.Name{Local: "settings"},
		Attr: []xml.Attr{{Name: xml.Name{Local: "pass"}, Value: s.Pass}},
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for _, c := range s.Components {
		if err := e.Encode(c); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

// BuildAnswerFile assembles a complete autounattend.xml document for profile.
// profile must already be validated by profile.ValidateProfile; BuildAnswerFile
// does not re-validate it. The only error it can return is an unexpected
// failure during XML serialization.
func BuildAnswerFile(p *profile.Profile) (string, error) {
	doc := unattendDoc{Xmlns: unattendNamespace}

	winPE := []interface{}{components.NewInternationalCoreWinPE(p.Language)}
	if setup := components.NewSetup(p.Edition, p.SystemTweaks.BypassWin11Requirements); setup != nil {
		winPE = append(winPE, setup)
	}
	doc.Settings = append(doc.Settings, settingsPass{Pass: "windowsPE", Components: winPE})

	specialize := []interface{}{components.NewInternationalCoreSpecialize(p.Language)}
	if shellSpecialize := components.NewShellSetupSpecialize(p.ComputerName, p.Timezone); shellSpecialize != nil {
		specialize = append(specialize, shellSpecialize)
	}
	if deployment := components.NewDeployment(p.SystemTweaks, p.BypassOnlineAccountRequirement, p.PasswordExpiration, p.AccountLockout, p.FileExplorer, p.Personalization, p.RemoveApps, p.StickyKeys, p.LockKeys, p.DesktopIcons, p.StartFolders, p.SystemScripts, p.DefaultUserScripts, p.UserOnceScripts); deployment != nil {
		specialize = append(specialize, deployment)
	}
	doc.Settings = append(doc.Settings, settingsPass{Pass: "specialize", Components: specialize})

	if shellOOBE := components.NewShellSetupOOBE(p.Accounts, p.FirstLogon, p.ExpressSettings, p.Wifi, p.BypassOnlineAccountRequirement, p.RemoveApps, p.RemoveFeatures, p.RemoveOptionalFeatures, p.SystemTweaks.DeleteHiddenJunctions, p.SystemTweaks.DeleteWindowsOld, p.FirstLogonScripts, p.RestartExplorerAfterScripts); shellOOBE != nil {
		doc.Settings = append(doc.Settings, settingsPass{Pass: "oobeSystem", Components: []interface{}{shellOOBE}})
	}

	body, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}

	return xml.Header + string(body), nil
}
