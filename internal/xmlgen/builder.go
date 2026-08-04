// Package xmlgen builds a Windows autounattend.xml answer file from a
// *profile.Profile. It never reads from disk and never imports cobra or
// bubbletea: input is a validated *profile.Profile, output is an XML string.
package xmlgen

import (
	"encoding/xml"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

const unattendNamespace = "urn:schemas-microsoft-com:unattend"

// unattendDoc is the root element of the answer file. Slice 0 emits it
// without any component children; later slices add them under Settings.
type unattendDoc struct {
	XMLName xml.Name `xml:"unattend"`
	Xmlns   string   `xml:"xmlns,attr"`
}

// BuildAnswerFile assembles a complete autounattend.xml document for profile.
// profile must already be validated by profile.ValidateProfile; BuildAnswerFile
// does not re-validate it. The only error it can return is an unexpected
// failure during XML serialization.
func BuildAnswerFile(_ *profile.Profile) (string, error) {
	doc := unattendDoc{Xmlns: unattendNamespace}

	body, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}

	return xml.Header + string(body), nil
}
