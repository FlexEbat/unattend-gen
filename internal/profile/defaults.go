package profile

// ProfilesDir is the conventional directory `profile list` (CLI) and the
// welcome screen (TUI) both read profiles from.
const ProfilesDir = "profiles"

// Default returns a profile that passes ValidateProfile: no accounts, no
// computer name (Windows generates one), interactive edition/express
// settings, and en-US language/locale/keyboard. Both `profile init` (with
// no --preset) and a fresh TUI session start from this.
func Default(name string) *Profile {
	return &Profile{
		SchemaVersion: 1,
		Name:          name,
		Language: LanguageSettings{
			UILanguage:     "en-US",
			Locale:         "en-US",
			KeyboardLayout: "en-US",
		},
		Edition:         EditionSettings{Mode: EditionModeInteractive},
		Accounts:        []UserAccount{},
		FirstLogon:      FirstLogon{Mode: FirstLogonNone},
		ExpressSettings: ExpressSettings{Mode: ExpressInteractive},
	}
}
