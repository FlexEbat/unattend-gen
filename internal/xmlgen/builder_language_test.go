package xmlgen

import (
	"encoding/xml"
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// testComponent captures the fields used by both International-Core and
// Setup components so a single struct can unmarshal either.
type testComponent struct {
	Name         string `xml:"name,attr"`
	InputLocale  string `xml:"InputLocale"`
	SystemLocale string `xml:"SystemLocale"`
	UILanguage   string `xml:"UILanguage"`
	UserLocale   string `xml:"UserLocale"`
	ProductKey   *struct {
		Key string `xml:"Key"`
	} `xml:"UserData>ProductKey"`
}

type testSettingsPass struct {
	Pass       string          `xml:"pass,attr"`
	Components []testComponent `xml:"component"`
}

type testDoc struct {
	XMLName  xml.Name           `xml:"unattend"`
	Settings []testSettingsPass `xml:"settings"`
}

func buildAndParse(t *testing.T, p *profile.Profile) testDoc {
	t.Helper()
	xmlStr, err := BuildAnswerFile(p)
	if err != nil {
		t.Fatalf("BuildAnswerFile: %v", err)
	}
	var doc testDoc
	if err := xml.Unmarshal([]byte(xmlStr), &doc); err != nil {
		t.Fatalf("Unmarshal result: %v\nxml:\n%s", err, xmlStr)
	}
	return doc
}

func findComponent(doc testDoc, pass, name string) *testComponent {
	for _, s := range doc.Settings {
		if s.Pass != pass {
			continue
		}
		for i := range s.Components {
			if s.Components[i].Name == name {
				return &s.Components[i]
			}
		}
	}
	return nil
}

func baseProfile() *profile.Profile {
	return &profile.Profile{
		SchemaVersion: 1,
		Name:          "demo",
		Language: profile.LanguageSettings{
			UILanguage:     "en-US",
			Locale:         "en-US",
			KeyboardLayout: "en-US",
		},
		Edition: profile.EditionSettings{
			Mode: profile.EditionModeInteractive,
		},
		Accounts:        []profile.UserAccount{},
		FirstLogon:      profile.FirstLogon{Mode: profile.FirstLogonNone},
		ExpressSettings: profile.ExpressSettings{Mode: profile.ExpressInteractive},
	}
}

func TestBuildAnswerFileGenericKey(t *testing.T) {
	p := baseProfile()
	edition := profile.EditionPro
	p.Edition = profile.EditionSettings{Mode: profile.EditionModeGenericKey, Edition: &edition}

	doc := buildAndParse(t, p)
	setup := findComponent(doc, "windowsPE", "Microsoft-Windows-Setup")
	if setup == nil || setup.ProductKey == nil {
		t.Fatal("expected Microsoft-Windows-Setup with a ProductKey")
	}
	const wantProKey = "VK7JG-NPHTM-C97JM-9MPGT-3V66T"
	if setup.ProductKey.Key != wantProKey {
		t.Fatalf("ProductKey.Key = %q, want %q", setup.ProductKey.Key, wantProKey)
	}
}

func TestBuildAnswerFileCustomKey(t *testing.T) {
	p := baseProfile()
	key := "AAAAA-BBBBB-CCCCC-DDDDD-EEEEE"
	p.Edition = profile.EditionSettings{Mode: profile.EditionModeCustomKey, ProductKey: &key}

	doc := buildAndParse(t, p)
	setup := findComponent(doc, "windowsPE", "Microsoft-Windows-Setup")
	if setup == nil || setup.ProductKey == nil {
		t.Fatal("expected Microsoft-Windows-Setup with a ProductKey")
	}
	if setup.ProductKey.Key != key {
		t.Fatalf("ProductKey.Key = %q, want %q", setup.ProductKey.Key, key)
	}
}

func TestBuildAnswerFileInteractiveEditionOmitsProductKey(t *testing.T) {
	p := baseProfile()
	p.Edition = profile.EditionSettings{Mode: profile.EditionModeInteractive}

	doc := buildAndParse(t, p)
	if setup := findComponent(doc, "windowsPE", "Microsoft-Windows-Setup"); setup != nil {
		t.Fatalf("expected no Microsoft-Windows-Setup component, got %+v", setup)
	}
}

func TestBuildAnswerFileLanguageFieldsPassThrough(t *testing.T) {
	p := baseProfile()
	p.Language = profile.LanguageSettings{
		UILanguage:     "de-DE",
		Locale:         "fr-FR",
		KeyboardLayout: "en-GB",
	}

	doc := buildAndParse(t, p)
	for _, pass := range []string{"windowsPE", "specialize"} {
		core := findComponent(doc, pass, "Microsoft-Windows-International-Core")
		if core == nil {
			t.Fatalf("expected International-Core component in pass %s", pass)
		}
		if core.UILanguage != p.Language.UILanguage {
			t.Errorf("pass %s: UILanguage = %q, want %q", pass, core.UILanguage, p.Language.UILanguage)
		}
		if core.SystemLocale != p.Language.Locale {
			t.Errorf("pass %s: SystemLocale = %q, want %q", pass, core.SystemLocale, p.Language.Locale)
		}
		if core.UserLocale != p.Language.Locale {
			t.Errorf("pass %s: UserLocale = %q, want %q", pass, core.UserLocale, p.Language.Locale)
		}
		if core.InputLocale != p.Language.KeyboardLayout {
			t.Errorf("pass %s: InputLocale = %q, want %q", pass, core.InputLocale, p.Language.KeyboardLayout)
		}
	}
}
