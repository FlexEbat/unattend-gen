package xmlgen

import (
	"testing"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

func TestBuildAnswerFileFirmwareEdition(t *testing.T) {
	p := baseProfile()
	p.Edition = profile.EditionSettings{Mode: profile.EditionModeFirmware}

	doc := buildAndParse(t, p)
	setup := findComponent(doc, "windowsPE", "Microsoft-Windows-Setup")
	if setup == nil {
		t.Fatal("expected a windowsPE Microsoft-Windows-Setup component")
	}
	if setup.ProductKey != nil {
		t.Fatalf("expected no ProductKey/Key element for firmware mode, got %q", setup.ProductKey.Key)
	}
	if setup.WillShowUI != "Never" {
		t.Fatalf("WillShowUI = %q, want %q", setup.WillShowUI, "Never")
	}
}

func TestBuildAnswerFileActivationKeyExplicit(t *testing.T) {
	p := baseProfile()
	key := "ABCDE-12345-ABCDE-12345-ABCDE"
	p.ActivationKey = &key

	doc := buildAndParseAccounts(t, p)
	shell := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Shell-Setup")
	if shell == nil {
		t.Fatal("expected a specialize Shell-Setup component")
	}
	if shell.ProductKey != key {
		t.Fatalf("ProductKey = %q, want %q", shell.ProductKey, key)
	}
}

func TestBuildAnswerFileActivationKeyFallsBackToCustomEditionKey(t *testing.T) {
	p := baseProfile()
	key := "FGHIJ-67890-FGHIJ-67890-FGHIJ"
	p.Edition = profile.EditionSettings{Mode: profile.EditionModeCustomKey, ProductKey: &key}
	p.ActivationKey = nil

	doc := buildAndParseAccounts(t, p)
	shell := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Shell-Setup")
	if shell == nil {
		t.Fatal("expected a specialize Shell-Setup component (activation key reused from edition)")
	}
	if shell.ProductKey != key {
		t.Fatalf("ProductKey = %q, want the edition's custom key %q reused for activation", shell.ProductKey, key)
	}
}

func TestBuildAnswerFileActivationKeyExplicitOverridesEditionKey(t *testing.T) {
	p := baseProfile()
	editionKey := "FGHIJ-67890-FGHIJ-67890-FGHIJ"
	p.Edition = profile.EditionSettings{Mode: profile.EditionModeCustomKey, ProductKey: &editionKey}
	activationKey := "ABCDE-12345-ABCDE-12345-ABCDE"
	p.ActivationKey = &activationKey

	doc := buildAndParseAccounts(t, p)
	shell := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Shell-Setup")
	if shell == nil {
		t.Fatal("expected a specialize Shell-Setup component")
	}
	if shell.ProductKey != activationKey {
		t.Fatalf("ProductKey = %q, want the explicit activation key %q, not the edition key", shell.ProductKey, activationKey)
	}
}

func TestBuildAnswerFileNoActivationKeyForGenericEdition(t *testing.T) {
	p := baseProfile()
	edition := profile.EditionPro
	p.Edition = profile.EditionSettings{Mode: profile.EditionModeGenericKey, Edition: &edition}
	p.ActivationKey = nil
	// Force a specialize Shell-Setup component via ComputerName so we can
	// inspect it even though ProductKey should stay empty.
	name := "PC1"
	p.ComputerName = &name

	doc := buildAndParseAccounts(t, p)
	shell := findShellComponentByName(doc, "specialize", "Microsoft-Windows-Shell-Setup")
	if shell == nil {
		t.Fatal("expected a specialize Shell-Setup component")
	}
	if shell.ProductKey != "" {
		t.Fatalf("ProductKey = %q, want empty for generic-key edition with no explicit activation key", shell.ProductKey)
	}
}

func TestValidateProfileRejectsMalformedActivationKey(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"activation_key": "not-a-key"
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for a malformed activation_key")
	}
}

func TestBuildAnswerFileProcessorArchitectureDefaultAMD64(t *testing.T) {
	p := baseProfile()
	p.ProcessorArchitecture = ""

	doc := buildAndParse(t, p)
	intl := findComponent(doc, "windowsPE", "Microsoft-Windows-International-Core-WinPE")
	if intl == nil {
		t.Fatal("expected a windowsPE International-Core-WinPE component")
	}
	if intl.ProcessorArchitecture != "amd64" {
		t.Fatalf("processorArchitecture = %q, want default %q", intl.ProcessorArchitecture, "amd64")
	}
}

func TestBuildAnswerFileProcessorArchitectureARM64(t *testing.T) {
	p := baseProfile()
	p.ProcessorArchitecture = profile.ArchARM64

	doc := buildAndParse(t, p)
	intl := findComponent(doc, "windowsPE", "Microsoft-Windows-International-Core-WinPE")
	if intl == nil {
		t.Fatal("expected a windowsPE International-Core-WinPE component")
	}
	if intl.ProcessorArchitecture != "arm64" {
		t.Fatalf("processorArchitecture = %q, want %q", intl.ProcessorArchitecture, "arm64")
	}

	specializeIntl := findComponent(doc, "specialize", "Microsoft-Windows-International-Core")
	if specializeIntl == nil {
		t.Fatal("expected a specialize International-Core component")
	}
	if specializeIntl.ProcessorArchitecture != "arm64" {
		t.Fatalf("specialize processorArchitecture = %q, want %q", specializeIntl.ProcessorArchitecture, "arm64")
	}
}

func TestValidateProfileRejectsUnknownProcessorArchitecture(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"processor_architecture": "ia64"
	}`)

	result := profile.ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for an unknown processor_architecture")
	}
}
