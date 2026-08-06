// Package profile defines the Profile data model shared by the CLI and TUI.
package profile

// LanguageSettings configures Windows UI language, regional locale and keyboard layout.
type LanguageSettings struct {
	UILanguage     string `json:"ui_language" validate:"required"`
	Locale         string `json:"locale" validate:"required"`
	KeyboardLayout string `json:"keyboard_layout" validate:"required"`
}

// EditionMode selects how the Windows edition/product key is resolved.
type EditionMode string

const (
	EditionModeGenericKey  EditionMode = "generic_key"
	EditionModeCustomKey   EditionMode = "custom_key"
	EditionModeInteractive EditionMode = "interactive"
)

// WindowsEdition is a Windows edition selectable via a built-in generic key.
type WindowsEdition string

const (
	EditionHome       WindowsEdition = "Home"
	EditionPro        WindowsEdition = "Pro"
	EditionEducation  WindowsEdition = "Education"
	EditionEnterprise WindowsEdition = "Enterprise"
)

// EditionSettings controls how the Windows edition and product key end up in the answer file.
type EditionSettings struct {
	Mode       EditionMode     `json:"mode" validate:"required,oneof=generic_key custom_key interactive"`
	Edition    *WindowsEdition `json:"edition"`     // required when Mode == EditionModeGenericKey
	ProductKey *string         `json:"product_key"` // required when Mode == EditionModeCustomKey
}

// AccountGroup is the local group a UserAccount is added to.
type AccountGroup string

const (
	GroupAdministrators AccountGroup = "Administrators"
	GroupUsers          AccountGroup = "Users"
)

// UserAccount is a single local account created during OOBE.
type UserAccount struct {
	Name        string       `json:"name" validate:"required,max=20"`
	DisplayName *string      `json:"display_name"`
	Password    *string      `json:"password" validate:"omitempty,min=1"` // empty string invalid, nil = no password
	Group       AccountGroup `json:"group" validate:"required,oneof=Administrators Users"`
}

// FirstLogonMode selects which account, if any, is auto-logged into after setup.
type FirstLogonMode string

const (
	FirstLogonFirstCreatedAccount FirstLogonMode = "first_created_account"
	FirstLogonBuiltinAdmin        FirstLogonMode = "builtin_administrator"
	FirstLogonNone                FirstLogonMode = "none"
)

// FirstLogon configures the auto-logon behavior after setup completes.
type FirstLogon struct {
	Mode                         FirstLogonMode `json:"mode" validate:"required,oneof=first_created_account builtin_administrator none"`
	BuiltinAdministratorPassword *string        `json:"builtin_administrator_password"` // required when Mode == FirstLogonBuiltinAdmin
}

// ExpressSettingsMode controls the OOBE express settings (telemetry) prompt.
type ExpressSettingsMode string

const (
	ExpressAllDisabled ExpressSettingsMode = "all_disabled"
	ExpressAllEnabled  ExpressSettingsMode = "all_enabled"
	ExpressInteractive ExpressSettingsMode = "interactive"
)

// ExpressSettings configures the OOBE express settings step.
type ExpressSettings struct {
	Mode ExpressSettingsMode `json:"mode" validate:"required,oneof=all_disabled all_enabled interactive"`
}

// SystemTweaks are boolean registry tweaks applied during specialize.
type SystemTweaks struct {
	DisableWindowsUpdate    bool `json:"disable_windows_update"`
	DisableUAC              bool `json:"disable_uac"`
	BypassWin11Requirements bool `json:"bypass_win11_requirements"`
}

// WifiAuthentication is the authentication type of a pre-configured Wi-Fi profile.
type WifiAuthentication string

const (
	WifiOpen         WifiAuthentication = "Open"
	WifiWPA2Personal WifiAuthentication = "WPA2Personal"
	WifiWPA3Personal WifiAuthentication = "WPA3Personal"
)

// WifiSettings describes a single Wi-Fi network profile to pre-configure.
type WifiSettings struct {
	SSID           string             `json:"ssid" validate:"required,max=32"`
	Authentication WifiAuthentication `json:"authentication" validate:"required,oneof=Open WPA2Personal WPA3Personal"`
	Password       *string            `json:"password"` // required when Authentication != WifiOpen
	ConnectHidden  bool               `json:"connect_hidden"`
}

// Profile is the full set of answer-file settings the CLI and TUI operate on.
type Profile struct {
	SchemaVersion   int              `json:"schema_version" validate:"required,eq=1"`
	Name            string           `json:"name" validate:"required"`
	Language        LanguageSettings `json:"language"`
	Edition         EditionSettings  `json:"edition"`
	ComputerName    *string          `json:"computer_name"` // nil = Windows generates a random name
	Accounts        []UserAccount    `json:"accounts" validate:"max=5,dive"`
	FirstLogon      FirstLogon       `json:"first_logon"`
	ExpressSettings ExpressSettings  `json:"express_settings"`
	SystemTweaks    SystemTweaks     `json:"system_tweaks"`
	Wifi            *WifiSettings    `json:"wifi"` // nil = Wi-Fi setup is skipped
}

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
