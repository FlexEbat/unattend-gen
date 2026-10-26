package profile

import (
	"encoding/json"
	"strings"
	"testing"
)

// validateWith validates Default("demo") after applying mutate.
func validateWith(t *testing.T, mutate func(p *Profile)) ValidationResult {
	t.Helper()
	p := Default("demo")
	mutate(p)
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	return ValidateProfile(data)
}

func wantError(t *testing.T, res ValidationResult, substr string) {
	t.Helper()
	for _, e := range res.Errors {
		if strings.Contains(e, substr) {
			return
		}
	}
	t.Fatalf("no error containing %q in %q", substr, res.Errors)
}

func wantValid(t *testing.T, res ValidationResult) {
	t.Helper()
	if len(res.Errors) != 0 || res.Profile == nil {
		t.Fatalf("expected a valid profile, got errors %q", res.Errors)
	}
}

func account(name string) UserAccount {
	return UserAccount{Name: name, Group: GroupUsers}
}

func TestValidateProfileRejectsUnknownField(t *testing.T) {
	res := ValidateProfile([]byte(`{"schema_version":1,"name":"x","acounts":[]}`))
	wantError(t, res, "Неизвестное поле")
	wantError(t, res, "acounts")
	if res.Profile != nil {
		t.Fatal("Profile must be nil on error")
	}
}

func TestValidateProfileRejectsUnknownNestedField(t *testing.T) {
	data := strings.Replace(string(defaultValidProfileJSON()), `"mode": "none"`, `"mode": "none", "typo": 1`, 1)
	wantError(t, ValidateProfile([]byte(data)), "Неизвестное поле")
}

func TestValidateProfileRejectsTrailingData(t *testing.T) {
	data := append(defaultValidProfileJSON(), []byte(` {}`)...)
	res := ValidateProfile(data)
	if len(res.Errors) == 0 {
		t.Fatal("expected an error for data after the JSON object")
	}
}

func TestValidateAccountsRejectsDuplicateNamesIgnoringCase(t *testing.T) {
	res := validateWith(t, func(p *Profile) { p.Accounts = []UserAccount{account("Alice"), account("alice")} })
	wantError(t, res, "больше одного раза")
}

func TestValidateAccountsRejectsReservedNames(t *testing.T) {
	for _, name := range []string{"Administrator", "guest", "SYSTEM"} {
		res := validateWith(t, func(p *Profile) { p.Accounts = []UserAccount{account(name)} })
		wantError(t, res, "зарезервировано")
	}
}

func TestValidateAccountsRejectsTrailingPeriodAndSpaces(t *testing.T) {
	wantError(t, validateWith(t, func(p *Profile) { p.Accounts = []UserAccount{account("bob.")} }), "точкой")
	wantError(t, validateWith(t, func(p *Profile) { p.Accounts = []UserAccount{account(" bob")} }), "пробелом")
}

func TestValidateAccountsCountsCharactersNotBytes(t *testing.T) {
	twenty := strings.Repeat("я", 20) // 40 bytes, 20 characters
	wantValid(t, validateWith(t, func(p *Profile) { p.Accounts = []UserAccount{account(twenty)} }))

	res := validateWith(t, func(p *Profile) { p.Accounts = []UserAccount{account(twenty + "я")} })
	wantError(t, res, "accounts[0].name: не больше 20 символов")
}

func TestValidateAccountsTooManyIsReportedOnce(t *testing.T) {
	res := validateWith(t, func(p *Profile) {
		for _, n := range []string{"a", "b", "c", "d", "e", "f"} {
			p.Accounts = append(p.Accounts, account(n))
		}
	})
	n := 0
	for _, e := range res.Errors {
		if strings.Contains(e, "accounts") {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("want exactly one accounts error, got %q", res.Errors)
	}
	wantError(t, res, "accounts: не больше 5 элементов")
}

func TestStructErrorsNameTheJSONPath(t *testing.T) {
	res := validateWith(t, func(p *Profile) {
		p.Accounts = []UserAccount{account("ok"), {Name: "bad", Group: "Root"}}
	})
	wantError(t, res, "accounts[1].group: допустимые значения: Administrators, Users")
}

func TestValidateRejectsDuplicateListEntries(t *testing.T) {
	res := validateWith(t, func(p *Profile) {
		p.RemoveApps = []RemovableApp{RemovableApps[0], RemovableApps[0]}
	})
	wantError(t, res, "remove_apps: значение")
}

func TestValidateRejectsWhitespaceOnlyScript(t *testing.T) {
	res := validateWith(t, func(p *Profile) {
		p.SystemScripts = []CustomScript{{Format: ScriptPs1, Content: "  \n\t "}}
	})
	wantError(t, res, "system_scripts[0]")
}

func TestValidateWifiPasswordLength(t *testing.T) {
	long := strings.Repeat("a", 64)
	res := validateWith(t, func(p *Profile) {
		p.Wifi = &WifiSettings{SSID: "n", Authentication: WifiWPA2Personal, Password: &long}
	})
	wantError(t, res, "не может быть длиннее 63")

	wantValid(t, validateWith(t, func(p *Profile) {
		p.Wifi = &WifiSettings{SSID: "n", Authentication: WifiWPA3Personal, Password: &long}
	}))
}

func TestValidateProductKeyAlphabet(t *testing.T) {
	good := "VK7JG-NPHTM-C97JM-9MPGT-3V66T"
	wantValid(t, validateWith(t, func(p *Profile) {
		p.Edition = EditionSettings{Mode: EditionModeCustomKey, ProductKey: &good}
	}))
	bad := "AAAAA-BBBBB-CCCCC-DDDDD-EEEEE" // A and E are not in the key alphabet
	wantError(t, validateWith(t, func(p *Profile) {
		p.Edition = EditionSettings{Mode: EditionModeCustomKey, ProductKey: &bad}
	}), "product_key")
}

func TestWarningsForPlaintextPasswords(t *testing.T) {
	pw := "secret"
	res := validateWith(t, func(p *Profile) {
		p.Accounts = []UserAccount{{Name: "u", Password: &pw, Group: GroupAdministrators}}
	})
	wantValid(t, res)
	if len(res.Warnings) == 0 || !strings.Contains(res.Warnings[0], "открытым текстом") {
		t.Fatalf("want a plaintext-password warning, got %q", res.Warnings)
	}

	res = validateWith(t, func(p *Profile) {
		p.Accounts = []UserAccount{{Name: "u", Password: &pw, Group: GroupAdministrators}}
		p.ObscurePasswords = true
	})
	if len(res.Warnings) == 0 || !strings.Contains(res.Warnings[0], "не шифрование") {
		t.Fatalf("want an obscure_passwords warning, got %q", res.Warnings)
	}
}

func TestWarningsForRiskyTweaksAndNoneForDefault(t *testing.T) {
	if res := validateWith(t, func(p *Profile) {}); len(res.Warnings) != 0 {
		t.Fatalf("default profile should not warn, got %q", res.Warnings)
	}
	res := validateWith(t, func(p *Profile) {
		p.SystemTweaks.DisableUAC = true
		p.SystemTweaks.DisableWindowsUpdate = true
	})
	if len(res.Warnings) != 2 {
		t.Fatalf("want 2 warnings, got %q", res.Warnings)
	}
}

func TestWarningsAreNotReportedWhenInvalid(t *testing.T) {
	res := validateWith(t, func(p *Profile) {
		p.SystemTweaks.DisableUAC = true
		p.Accounts = []UserAccount{account("Administrator")}
	})
	if len(res.Errors) == 0 || res.Warnings != nil {
		t.Fatalf("errors=%q warnings=%q", res.Errors, res.Warnings)
	}
}
