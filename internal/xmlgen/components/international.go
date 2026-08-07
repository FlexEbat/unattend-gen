// Package components builds individual autounattend.xml <component> elements
// from profile settings. It never reads from disk and never imports cobra or
// bubbletea.
package components

import (
	"encoding/xml"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// wcmNamespace is the "Windows Configuration Manager" namespace unattend.xml
// uses for the wcm:action attribute on every list-item element (LocalAccount,
// RunSynchronousCommand, SynchronousCommand, ...).
const wcmNamespace = "http://schemas.microsoft.com/WMIConfig/2002/State"

// wcmActionAdd is the value wcm:action carries on every list item we emit;
// unattend.xml has no other action verb we need.
const wcmActionAdd = "add"

// Every <component> element in an autounattend.xml carries the same
// attributes. standardAttrs holds them so each component file sets them
// once instead of repeating the same literals.
type standardAttrs struct {
	ProcessorArchitecture string `xml:"processorArchitecture,attr"`
	PublicKeyToken        string `xml:"publicKeyToken,attr"`
	Language              string `xml:"language,attr"`
	VersionScope          string `xml:"versionScope,attr"`
	XmlnsWcm              string `xml:"xmlns:wcm,attr"`
}

func newStandardAttrs() standardAttrs {
	return standardAttrs{
		ProcessorArchitecture: "amd64",
		PublicKeyToken:        "31bf3856ad364e35",
		Language:              "neutral",
		VersionScope:          "nonSxS",
		XmlnsWcm:              wcmNamespace,
	}
}

// InternationalCore is the Microsoft-Windows-International-Core component.
// Real Windows Setup uses a different component name per pass: the
// "-WinPE" suffix in windowsPE, no suffix in specialize. Using the same
// name in both (a plausible but wrong reading of a same-named source
// requirement) makes windowsPE silently ignore the setting.
type InternationalCore struct {
	XMLName xml.Name `xml:"component"`
	Name    string   `xml:"name,attr"`
	standardAttrs
	InputLocale  string `xml:"InputLocale"`
	SystemLocale string `xml:"SystemLocale"`
	UILanguage   string `xml:"UILanguage"`
	UserLocale   string `xml:"UserLocale"`
}

// NewInternationalCoreWinPE builds the windowsPE-pass component.
func NewInternationalCoreWinPE(lang profile.LanguageSettings) InternationalCore {
	return newInternationalCore("Microsoft-Windows-International-Core-WinPE", lang)
}

// NewInternationalCoreSpecialize builds the specialize-pass component.
func NewInternationalCoreSpecialize(lang profile.LanguageSettings) InternationalCore {
	return newInternationalCore("Microsoft-Windows-International-Core", lang)
}

// newInternationalCore builds the component from LanguageSettings.
// KeyboardLayout maps to InputLocale, Locale maps to both SystemLocale and
// UserLocale.
func newInternationalCore(name string, lang profile.LanguageSettings) InternationalCore {
	return InternationalCore{
		Name:          name,
		standardAttrs: newStandardAttrs(),
		InputLocale:   lang.KeyboardLayout,
		SystemLocale:  lang.Locale,
		UILanguage:    lang.UILanguage,
		UserLocale:    lang.Locale,
	}
}
