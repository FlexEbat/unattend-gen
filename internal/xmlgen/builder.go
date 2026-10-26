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
	XMLName    xml.Name         `xml:"unattend"`
	Xmlns      string           `xml:"xmlns,attr"`
	Settings   []settingsPass   `xml:"settings"`
	Extensions *extensionsBlock `xml:"Extensions,omitempty"`
}

// extensionsBlock is the <Extensions> element: files that Windows Setup does
// not know about but that our extraction script writes to disk during the
// specialize pass (see components.OffloadLongCommands).
type extensionsBlock struct {
	XMLName       xml.Name        `xml:"https://schneegans.de/windows/unattend-generator/ Extensions"`
	ExtractScript cdataText       `xml:"ExtractScript"`
	Files         []extensionFile `xml:"File"`
}

type cdataText struct {
	Text string `xml:",cdata"`
}

type extensionFile struct {
	Path string `xml:"path,attr"`
	Text string `xml:",cdata"`
}

// settingsPass is one <settings pass="..."> block holding an ordered list of
// heterogeneous <component> elements. It implements xml.Marshaler because its
// component list mixes concrete types that only share the "component" tag.
type settingsPass struct {
	Pass       string
	Components []any
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
//
// Every command in the result fits Windows Setup's 259-character limit:
// longer ones are split or moved into files extracted from <Extensions>.
func BuildAnswerFile(p *profile.Profile) (string, error) {
	return buildAnswerFile(p, true)
}

// buildAnswerFile is BuildAnswerFile with the long-command offloading
// switchable: tests inspect the inline form, where each feature's command is
// readable as one string, and separately check that offloading it is lossless.
func buildAnswerFile(p *profile.Profile, offload bool) (string, error) {
	doc := unattendDoc{Xmlns: unattendNamespace}

	setup := components.NewSetup(p.Edition, p.SystemTweaks.BypassWin11Requirements, p.UseNarrator, p.ProcessorArchitecture)
	deployment := components.NewDeployment(p)
	shellOOBE := components.NewShellSetupOOBE(p)

	if offload {
		var files []components.ExtensionFile
		deployment, files = components.OffloadLongCommands(setup, deployment, shellOOBE, p.ProcessorArchitecture)
		if len(files) > 0 {
			doc.Extensions = &extensionsBlock{ExtractScript: cdataText{Text: components.ExtractScript}}
			for _, f := range files {
				doc.Extensions.Files = append(doc.Extensions.Files, extensionFile{Path: f.Path, Text: f.Content})
			}
		}
	}

	winPE := []any{components.NewInternationalCoreWinPE(p.Language, p.ProcessorArchitecture)}
	if setup != nil {
		winPE = append(winPE, setup)
	}
	doc.Settings = append(doc.Settings, settingsPass{Pass: "windowsPE", Components: winPE})

	specialize := []any{components.NewInternationalCoreSpecialize(p.Language, p.ProcessorArchitecture)}
	activationKey := components.ResolveActivationKey(p.Edition, p.ActivationKey)
	if shellSpecialize := components.NewShellSetupSpecialize(p.ComputerName, p.Timezone, p.ComputerNameScript, activationKey, p.ProcessorArchitecture); shellSpecialize != nil {
		specialize = append(specialize, shellSpecialize)
	}
	if deployment != nil {
		specialize = append(specialize, deployment)
	}
	doc.Settings = append(doc.Settings, settingsPass{Pass: "specialize", Components: specialize})

	if shellOOBE != nil {
		doc.Settings = append(doc.Settings, settingsPass{Pass: "oobeSystem", Components: []any{shellOOBE}})
	}

	body, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}

	return xml.Header + string(body), nil
}
