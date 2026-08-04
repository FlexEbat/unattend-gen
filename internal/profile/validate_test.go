package profile

import "testing"

func defaultValidProfileJSON() []byte {
	return []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"},
		"system_tweaks": {}
	}`)
}

func TestValidateProfileDefaultIsValid(t *testing.T) {
	result := ValidateProfile(defaultValidProfileJSON())
	if len(result.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}
	if result.Profile == nil {
		t.Fatal("expected non-nil Profile on success")
	}
}

func TestValidateProfileRejectsWrongSchemaVersion(t *testing.T) {
	data := []byte(`{
		"schema_version": 2,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"}
	}`)

	result := ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for schema_version 2, got none")
	}
	if result.Profile != nil {
		t.Fatal("expected nil Profile on validation failure")
	}
}

func TestValidateProfileRejectsInvalidBCP47(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "not_a_bcp47_tag!", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"}
	}`)

	result := ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for invalid ui_language, got none")
	}
	if result.Profile != nil {
		t.Fatal("expected nil Profile on validation failure")
	}
}

func TestValidateProfileGenericKeyRequiresEdition(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "generic_key"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"}
	}`)

	result := ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for generic_key without edition, got none")
	}
}

func TestValidateProfileCustomKeyRequiresFormat(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "custom_key", "product_key": "not-a-valid-key"},
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"}
	}`)

	result := ValidateProfile(data)
	if len(result.Errors) == 0 {
		t.Fatal("expected an error for malformed product_key, got none")
	}
}

func TestValidateProfileRejectsLongComputerName(t *testing.T) {
	data := []byte(`{
		"schema_version": 1,
		"name": "demo",
		"language": {"ui_language": "en-US", "locale": "en-US", "keyboard_layout": "en-US"},
		"edition": {"mode": "interactive"},
		"computer_name": "this-name-is-too-long",
		"accounts": [],
		"first_logon": {"mode": "none"},
		"express_settings": {"mode": "interactive"}
	}`)

	result := ValidateProfile(data)
	found := false
	for _, e := range result.Errors {
		if e == "Имя компьютера длиннее 15 символов" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected the exact computer-name-too-long message, got %v", result.Errors)
	}
}
