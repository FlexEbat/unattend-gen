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
	DisableWindowsUpdate       bool `json:"disable_windows_update"`
	DisableUAC                 bool `json:"disable_uac"`
	BypassWin11Requirements    bool `json:"bypass_win11_requirements"`
	DisableSmartAppControl     bool `json:"disable_smart_app_control"`
	DisableSmartScreen         bool `json:"disable_smart_screen"`
	DisableFastStartup         bool `json:"disable_fast_startup"`
	DisableSystemRestore       bool `json:"disable_system_restore"`
	EnableLongPaths            bool `json:"enable_long_paths"`
	EnableRemoteDesktop        bool `json:"enable_remote_desktop"`
	AllowPowerShellScripts     bool `json:"allow_powershell_scripts"`
	DisableLastAccessTimestamp bool `json:"disable_last_access_timestamp"`
	PreventDeviceEncryption    bool `json:"prevent_device_encryption"`
	DisableAutoSignOnLastUser  bool `json:"disable_auto_sign_on_last_user"`
	DisableWPBT                bool `json:"disable_wpbt"`
	AuditProcessCreation       bool `json:"audit_process_creation"`
	HideEdgeFirstRun           bool `json:"hide_edge_first_run"`
	DisableEdgeStartupBoost    bool `json:"disable_edge_startup_boost"`
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

// RemovableApp is a preinstalled app xmlgen knows how to remove. The set is
// intentionally not every app schneegans.de/windows/unattend-generator
// offers: apps removed via Remove-WindowsCapability (Internet Explorer,
// WordPad, legacy PowerShell, Speech, ...) need a different mechanism and
// aren't covered yet.
type RemovableApp string

const (
	App3DViewer            RemovableApp = "3DViewer"
	AppCalculator          RemovableApp = "Calculator"
	AppCamera              RemovableApp = "Camera"
	AppClipchamp           RemovableApp = "Clipchamp"
	AppClock               RemovableApp = "Clock"
	AppCopilot             RemovableApp = "Copilot"
	AppCortana             RemovableApp = "Cortana"
	AppFamily              RemovableApp = "Family"
	AppFeedbackHub         RemovableApp = "FeedbackHub"
	AppGetHelp             RemovableApp = "GetHelp"
	AppMailAndCalendar     RemovableApp = "MailAndCalendar"
	AppMaps                RemovableApp = "Maps"
	AppMixedReality        RemovableApp = "MixedReality"
	AppMoviesAndTV         RemovableApp = "MoviesAndTV"
	AppNews                RemovableApp = "News"
	AppOffice              RemovableApp = "Office"
	AppOneNote             RemovableApp = "OneNote"
	AppPaint3D             RemovableApp = "Paint3D"
	AppPeople              RemovableApp = "People"
	AppPhoneLink           RemovableApp = "PhoneLink"
	AppPowerAutomate       RemovableApp = "PowerAutomate"
	AppQuickAssist         RemovableApp = "QuickAssist"
	AppSkype               RemovableApp = "Skype"
	AppSnippingTool        RemovableApp = "SnippingTool"
	AppSolitaireCollection RemovableApp = "SolitaireCollection"
	AppStickyNotes         RemovableApp = "StickyNotes"
	AppTeams               RemovableApp = "Teams"
	AppTips                RemovableApp = "Tips"
	AppToDo                RemovableApp = "ToDo"
	AppVoiceRecorder       RemovableApp = "VoiceRecorder"
	AppWeather             RemovableApp = "Weather"
	AppXboxApps            RemovableApp = "XboxApps"
)

// RemovableApps lists every valid RemoveApps entry, in the order shown in
// the TUI.
var RemovableApps = []RemovableApp{
	App3DViewer, AppCalculator, AppCamera, AppClipchamp, AppClock, AppCopilot,
	AppCortana, AppFamily, AppFeedbackHub, AppGetHelp, AppMailAndCalendar,
	AppMaps, AppMixedReality, AppMoviesAndTV, AppNews, AppOffice, AppOneNote,
	AppPaint3D, AppPeople, AppPhoneLink, AppPowerAutomate, AppQuickAssist,
	AppSkype, AppSnippingTool, AppSolitaireCollection, AppStickyNotes,
	AppTeams, AppTips, AppToDo, AppVoiceRecorder, AppWeather, AppXboxApps,
}

// RemovableFeature is a Windows capability (DISM Get-WindowsCapability /
// Remove-WindowsCapability) xmlgen knows how to remove. Different mechanism
// from RemovableApp (Appx packages): PowerShell 2.0 and Windows Fax and
// Scan are on schneegans.de's list too but use a third mechanism (legacy
// Windows optional features, Disable-WindowsOptionalFeature) not covered
// here.
type RemovableFeature string

const (
	FeatureInternetExplorer RemovableFeature = "InternetExplorer"
	FeatureWordPad          RemovableFeature = "WordPad"
	FeaturePowerShellISE    RemovableFeature = "PowerShellISE"
	FeatureOpenSSHClient    RemovableFeature = "OpenSSHClient"
	FeatureMediaPlayer      RemovableFeature = "MediaPlayer"
	FeatureSpeech           RemovableFeature = "Speech"
	FeatureHandwriting      RemovableFeature = "Handwriting"
)

// RemovableFeatures lists every valid RemoveFeatures entry, in the order
// shown in the TUI.
var RemovableFeatures = []RemovableFeature{
	FeatureInternetExplorer, FeatureWordPad, FeaturePowerShellISE,
	FeatureOpenSSHClient, FeatureMediaPlayer, FeatureSpeech, FeatureHandwriting,
}

// ScriptFormat is the file format a CustomScript is written and run as.
type ScriptFormat string

const (
	ScriptCmd ScriptFormat = "cmd"
	ScriptPs1 ScriptFormat = "ps1"
	ScriptReg ScriptFormat = "reg"
	ScriptVbs ScriptFormat = "vbs"
)

// CustomScript is one script to run during setup, in one of System,
// DefaultUser, FirstLogon or UserOnce (see Profile). DefaultUser scripts
// don't support ScriptVbs.
type CustomScript struct {
	Format  ScriptFormat `json:"format" validate:"required,oneof=cmd ps1 reg vbs"`
	Content string       `json:"content" validate:"required"`
}

// PasswordExpirationMode controls whether local account passwords expire.
type PasswordExpirationMode string

const (
	PasswordExpirationDefault PasswordExpirationMode = "default" // Windows default: 42 days
	PasswordExpirationNever   PasswordExpirationMode = "never"   // NIST no longer recommends expiration
	PasswordExpirationCustom  PasswordExpirationMode = "custom"
)

// PasswordExpirationSettings maps to `net accounts /maxpwage`. The empty
// Mode ("") behaves like PasswordExpirationDefault: no command is emitted,
// Windows' own default (42 days) applies.
type PasswordExpirationSettings struct {
	Mode PasswordExpirationMode `json:"mode" validate:"omitempty,oneof=default never custom"`
	Days *int                   `json:"days"` // required when Mode == custom
}

// AccountLockoutMode controls the local account lockout policy.
type AccountLockoutMode string

const (
	AccountLockoutDefault  AccountLockoutMode = "default" // Windows default: 10 attempts / 10 min window / 10 min duration
	AccountLockoutDisabled AccountLockoutMode = "disabled"
	AccountLockoutCustom   AccountLockoutMode = "custom"
)

// AccountLockoutSettings maps to `net accounts /lockoutthreshold` etc. The
// empty Mode ("") behaves like AccountLockoutDefault: no command is
// emitted, Windows' own default applies.
type AccountLockoutSettings struct {
	Mode            AccountLockoutMode `json:"mode" validate:"omitempty,oneof=default disabled custom"`
	Threshold       *int               `json:"threshold"`        // failed attempts; required when Mode == custom
	WindowMinutes   *int               `json:"window_minutes"`   // required when Mode == custom
	DurationMinutes *int               `json:"duration_minutes"` // required when Mode == custom
}

// HiddenFilesMode controls which hidden/system files File Explorer shows.
type HiddenFilesMode string

const (
	HiddenFilesDefault    HiddenFilesMode = "default"     // File Explorer default: don't show hidden files
	HiddenFilesShowHidden HiddenFilesMode = "show_hidden" // show hidden files, keep protected OS files hidden
	HiddenFilesShowAll    HiddenFilesMode = "show_all"    // show hidden files and protected OS files
)

// FileExplorerSettings are applied to C:\Users\Default\NTUSER.DAT (the
// account template), the same mechanism as DefaultUserScripts, so they
// affect every account including ones created after setup. The zero value
// changes nothing.
type FileExplorerSettings struct {
	HiddenFiles          HiddenFilesMode `json:"hidden_files" validate:"omitempty,oneof=default show_hidden show_all"`
	ShowFileExtensions   bool            `json:"show_file_extensions"`
	ClassicContextMenu   bool            `json:"classic_context_menu"` // Windows 11 only; no-op on Windows 10
	HideFolderTooltips   bool            `json:"hide_folder_tooltips"`
	OpenToThisPC         bool            `json:"open_to_this_pc"` // vs. Quick access, the default
	ShowEndTaskInTaskbar bool            `json:"show_end_task_in_taskbar"`
}

// ColorTheme is a Windows light/dark theme choice.
type ColorTheme string

const (
	ColorThemeLight ColorTheme = "light"
	ColorThemeDark  ColorTheme = "dark"
)

// PersonalizationSettings covers the "Colors" part of the site's
// Personalization section (not wallpaper/lock screen, which need actual
// image bytes and a different mechanism — a later slice). Applied to
// C:\Users\Default\NTUSER.DAT, same as FileExplorerSettings, so every
// account gets them. The zero value changes nothing: DisableTransparency
// is false by default because Windows' own default is transparency *on*.
type PersonalizationSettings struct {
	SystemTheme              ColorTheme `json:"system_theme" validate:"omitempty,oneof=light dark"`
	AppsTheme                ColorTheme `json:"apps_theme" validate:"omitempty,oneof=light dark"`
	AccentColor              *string    `json:"accent_color"` // 6 hex digits, RRGGBB; nil = Windows default
	ShowAccentOnStartTaskbar bool       `json:"show_accent_on_start_taskbar"`
	ShowAccentOnTitleBars    bool       `json:"show_accent_on_title_bars"`
	DisableTransparency      bool       `json:"disable_transparency"`
}

// Profile is the full set of answer-file settings the CLI and TUI operate on.
type Profile struct {
	SchemaVersion                  int                        `json:"schema_version" validate:"required,eq=1"`
	Name                           string                     `json:"name" validate:"required"`
	Language                       LanguageSettings           `json:"language"`
	Edition                        EditionSettings            `json:"edition"`
	ComputerName                   *string                    `json:"computer_name"` // nil = Windows generates a random name
	Timezone                       *string                    `json:"timezone"`      // nil = Windows determines it automatically; a Windows time zone ID such as "Russian Standard Time"
	Accounts                       []UserAccount              `json:"accounts" validate:"max=5,dive"`
	FirstLogon                     FirstLogon                 `json:"first_logon"`
	ExpressSettings                ExpressSettings            `json:"express_settings"`
	SystemTweaks                   SystemTweaks               `json:"system_tweaks"`
	Wifi                           *WifiSettings              `json:"wifi"` // nil = Wi-Fi setup is skipped
	BypassOnlineAccountRequirement bool                       `json:"bypass_online_account_requirement"`
	RemoveApps                     []RemovableApp             `json:"remove_apps"`
	RemoveFeatures                 []RemovableFeature         `json:"remove_features"`
	PasswordExpiration             PasswordExpirationSettings `json:"password_expiration"`
	AccountLockout                 AccountLockoutSettings     `json:"account_lockout"`
	FileExplorer                   FileExplorerSettings       `json:"file_explorer"`
	Personalization                PersonalizationSettings    `json:"personalization"`
	// SystemScripts run in the system context, before user accounts are
	// created. Max 4.
	SystemScripts []CustomScript `json:"system_scripts" validate:"max=4,dive"`
	// DefaultUserScripts modify C:\Users\Default\NTUSER.DAT (the template
	// every new account is created from), so they affect every account,
	// including ones created after setup. Run after SystemScripts, before
	// accounts are created. Max 3, no ScriptVbs.
	DefaultUserScripts []CustomScript `json:"default_user_scripts" validate:"max=3,dive"`
	// FirstLogonScripts run once, when the first user logs on (typically
	// the account used for auto-logon). Max 4.
	FirstLogonScripts []CustomScript `json:"first_logon_scripts" validate:"max=4,dive"`
	// UserOnceScripts run once for every user the first time they log on,
	// including accounts created after setup — implemented via a RunOnce
	// entry written to the default user hive, same as DefaultUserScripts.
	// Max 4.
	UserOnceScripts []CustomScript `json:"user_once_scripts" validate:"max=4,dive"`
	// RestartExplorerAfterScripts restarts explorer.exe once FirstLogon and
	// UserOnce scripts have run, so Start menu/taskbar changes take effect
	// immediately.
	RestartExplorerAfterScripts bool `json:"restart_explorer_after_scripts"`
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
