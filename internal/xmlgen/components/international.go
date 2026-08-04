// Package components builds individual autounattend.xml <component> elements
// from profile settings. It never reads from disk and never imports cobra or
// bubbletea.
package components

import (
	"encoding/xml"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// InternationalCore is the Microsoft-Windows-International-Core component,
// used identically in both the windowsPE and specialize passes.
type InternationalCore struct {
	XMLName               xml.Name `xml:"component"`
	Name                  string   `xml:"name,attr"`
	ProcessorArchitecture string   `xml:"processorArchitecture,attr"`
	PublicKeyToken        string   `xml:"publicKeyToken,attr"`
	Language              string   `xml:"language,attr"`
	VersionScope          string   `xml:"versionScope,attr"`
	InputLocale           string   `xml:"InputLocale"`
	SystemLocale          string   `xml:"SystemLocale"`
	UILanguage            string   `xml:"UILanguage"`
	UserLocale            string   `xml:"UserLocale"`
}

// NewInternationalCore builds the component from LanguageSettings. KeyboardLayout
// maps to InputLocale, Locale maps to both SystemLocale and UserLocale.
func NewInternationalCore(lang profile.LanguageSettings) InternationalCore {
	return InternationalCore{
		Name:                  "Microsoft-Windows-International-Core",
		ProcessorArchitecture: "amd64",
		PublicKeyToken:        "31bf3856ad364e35",
		Language:              "neutral",
		VersionScope:          "nonSxS",
		InputLocale:           lang.KeyboardLayout,
		SystemLocale:          lang.Locale,
		UILanguage:            lang.UILanguage,
		UserLocale:            lang.Locale,
	}
}
