package components

// Every <component> element in an autounattend.xml carries the same four
// attributes. standardAttrs holds them so each component file sets them
// once instead of repeating the same four literals.
type standardAttrs struct {
	ProcessorArchitecture string `xml:"processorArchitecture,attr"`
	PublicKeyToken        string `xml:"publicKeyToken,attr"`
	Language              string `xml:"language,attr"`
	VersionScope          string `xml:"versionScope,attr"`
}

func newStandardAttrs() standardAttrs {
	return standardAttrs{
		ProcessorArchitecture: "amd64",
		PublicKeyToken:        "31bf3856ad364e35",
		Language:              "neutral",
		VersionScope:          "nonSxS",
	}
}
