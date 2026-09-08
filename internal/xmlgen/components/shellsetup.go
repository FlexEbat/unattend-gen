package components

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"unicode/utf16"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

const (
	shellSetupName = "Microsoft-Windows-Shell-Setup"
	adminUsername  = "Administrator"
)

// ShellSetupSpecialize is the Microsoft-Windows-Shell-Setup component in the
// specialize pass, carrying the computer name and time zone. Registry
// tweaks used to live here too, but Shell-Setup has no RunSynchronous list
// in the real schema: that belongs to Microsoft-Windows-Deployment (see
// NewDeployment below).
type ShellSetupSpecialize struct {
	XMLName xml.Name `xml:"component"`
	Name    string   `xml:"name,attr"`
	standardAttrs
	ComputerName string `xml:"ComputerName,omitempty"`
	TimeZone     string `xml:"TimeZone,omitempty"`
}

// NewShellSetupSpecialize builds the specialize-pass component from the
// computer name and time zone. Returns nil when both are nil: Windows
// generates a random name and determines the time zone itself.
// computerNameScript (slice 22), when set, overrides computerName with the
// "TEMPNAME" placeholder — ComputerNameScriptCommand overwrites it moments
// later at runtime; mutual exclusivity with computerName is enforced in
// validate.go.
func NewShellSetupSpecialize(computerName, timezone, computerNameScript *string) *ShellSetupSpecialize {
	if computerName == nil && timezone == nil && computerNameScript == nil {
		return nil
	}
	s := &ShellSetupSpecialize{
		Name:          shellSetupName,
		standardAttrs: newStandardAttrs(),
	}
	if computerNameScript != nil {
		s.ComputerName = "TEMPNAME"
	} else if computerName != nil {
		s.ComputerName = *computerName
	}
	if timezone != nil {
		s.TimeZone = *timezone
	}
	return s
}

// Registry edits for the corresponding profile settings. They run in the
// specialize pass. bypassOnlineAccountCommand sets the same registry value
// the (now removed) `oobe\bypassnro` command used to set; real-world
// reliability varies by Windows build and Microsoft has patched around it
// more than once, so this is best-effort, not a guarantee.
const (
	disableWindowsUpdateCommand = `cmd.exe /c reg add "HKLM\SOFTWARE\Policies\Microsoft\Windows\WindowsUpdate\AU" /v NoAutoUpdate /t REG_DWORD /d 1 /f`
	disableUACCommand           = `cmd.exe /c reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" /v EnableLUA /t REG_DWORD /d 0 /f`
	bypassOnlineAccountCommand  = `cmd.exe /c reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\OOBE" /v BypassNRO /t REG_DWORD /d 1 /f`

	disableSmartAppControlCommand  = `cmd.exe /c reg add "HKLM\SYSTEM\CurrentControlSet\Control\CI\Policy" /v VerifiedAndReputablePolicyState /t REG_DWORD /d 0 /f`
	disableSmartScreenCommand      = `cmd.exe /c reg add "HKLM\SOFTWARE\Policies\Microsoft\Windows\System" /v EnableSmartScreen /t REG_DWORD /d 0 /f && reg add "HKLM\SOFTWARE\Policies\Microsoft\Edge" /v SmartScreenEnabled /t REG_DWORD /d 0 /f`
	disableFastStartupCommand      = `cmd.exe /c reg add "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Power" /v HiberbootEnabled /t REG_DWORD /d 0 /f`
	disableSystemRestoreCommand    = `cmd.exe /c reg add "HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\SystemRestore" /v DisableSR /t REG_DWORD /d 1 /f`
	enableLongPathsCommand         = `cmd.exe /c reg add "HKLM\SYSTEM\CurrentControlSet\Control\FileSystem" /v LongPathsEnabled /t REG_DWORD /d 1 /f`
	enableRemoteDesktopCommand     = `cmd.exe /c reg add "HKLM\SYSTEM\CurrentControlSet\Control\Terminal Server" /v fDenyTSConnections /t REG_DWORD /d 0 /f && netsh advfirewall firewall set rule group="remote desktop" new enable=yes`
	allowPowerShellScriptsCommand  = `cmd.exe /c reg add "HKLM\SOFTWARE\Microsoft\PowerShell\1\ShellIds\Microsoft.PowerShell" /v ExecutionPolicy /t REG_SZ /d RemoteSigned /f`
	disableLastAccessCommand       = `cmd.exe /c fsutil behavior set disablelastaccess 1`
	preventDeviceEncryptionCommand = `cmd.exe /c reg add "HKLM\SYSTEM\CurrentControlSet\Control\BitLocker" /v PreventDeviceEncryption /t REG_DWORD /d 1 /f`
	disableAutoSignOnCommand       = `cmd.exe /c reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" /v DisableAutomaticRestartSignOn /t REG_DWORD /d 1 /f`
	disableWPBTCommand             = `cmd.exe /c reg add "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager" /v DisableWpbtExecution /t REG_DWORD /d 1 /f`
	auditProcessCreationCommand    = `cmd.exe /c auditpol /set /subcategory:"Process Creation" /success:enable && reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System\Audit" /v ProcessCreationIncludeCmdLine_Enabled /t REG_DWORD /d 1 /f`
	hideEdgeFirstRunCommand        = `cmd.exe /c reg add "HKLM\SOFTWARE\Policies\Microsoft\Edge" /v HideFirstRunExperience /t REG_DWORD /d 1 /f`
	disableEdgeStartupBoostCommand = `cmd.exe /c reg add "HKLM\SOFTWARE\Policies\Microsoft\Edge" /v StartupBoostEnabled /t REG_DWORD /d 0 /f && reg add "HKLM\SOFTWARE\Policies\Microsoft\Edge" /v BackgroundModeEnabled /t REG_DWORD /d 0 /f`
)

// Slice 16 additions (tech.md backlog group B) live in optimizations.go:
// disableStartupSoundCommand, disableAppSuggestionsCommand,
// preventDeviceAppsCommand are simple consts folded into enabledCommands
// below; PreventAutomaticRebootCommand, TurnOffSystemSounds*Command,
// DisableAppSuggestionsDefaultUserCommand, DisablePointerPrecisionCommand
// and DeleteJunctions*Command are funcs (they build multi-statement
// commands, some with their own default-user-hive mount/unmount cycle) and
// are appended conditionally after the loop.

// Deployment is the Microsoft-Windows-Deployment component, specialize pass:
// the real home of RunSynchronousCommand outside windowsPE.
type Deployment struct {
	XMLName xml.Name `xml:"component"`
	Name    string   `xml:"name,attr"`
	standardAttrs
	RunSynchronous *runSynchronous `xml:"RunSynchronous,omitempty"`
}

// NewDeployment builds the specialize-pass component for every profile
// setting that maps to a single registry-editing command. Returns nil when
// none are set: an empty component is not emitted.
// NewDeployment builds the specialize-pass component for every profile
// setting that maps to a single registry-editing or script-running
// command: system tweaks, the online-account bypass, System scripts,
// DefaultUser scripts and the UserOnce RunOnce registration (in that
// order). Returns nil when none are set: an empty component is not
// emitted.
func NewDeployment(tweaks profile.SystemTweaks, bypassOnlineAccountRequirement bool, passwordExpiration profile.PasswordExpirationSettings, accountLockout profile.AccountLockoutSettings, fileExplorer profile.FileExplorerSettings, personalization profile.PersonalizationSettings, removeApps []profile.RemovableApp, stickyKeys profile.StickyKeysSettings, lockKeys *profile.LockKeySettings, desktopIcons map[profile.DesktopIcon]bool, startFolders []profile.StartFolder, appLockerPolicyXML *string, useNarrator bool, computerNameScript *string, systemScripts, defaultUserScripts, userOnceScripts []profile.CustomScript) *Deployment {
	enabledCommands := []struct {
		enabled bool
		command string
	}{
		{tweaks.DisableWindowsUpdate, disableWindowsUpdateCommand},
		{tweaks.DisableUAC, disableUACCommand},
		{bypassOnlineAccountRequirement, bypassOnlineAccountCommand},
		{tweaks.DisableSmartAppControl, disableSmartAppControlCommand},
		{tweaks.DisableSmartScreen, disableSmartScreenCommand},
		{tweaks.DisableFastStartup, disableFastStartupCommand},
		{tweaks.DisableSystemRestore, disableSystemRestoreCommand},
		{tweaks.EnableLongPaths, enableLongPathsCommand},
		{tweaks.EnableRemoteDesktop, enableRemoteDesktopCommand},
		{tweaks.AllowPowerShellScripts, allowPowerShellScriptsCommand},
		{tweaks.DisableLastAccessTimestamp, disableLastAccessCommand},
		{tweaks.PreventDeviceEncryption, preventDeviceEncryptionCommand},
		{tweaks.DisableAutoSignOnLastUser, disableAutoSignOnCommand},
		{tweaks.DisableWPBT, disableWPBTCommand},
		{tweaks.AuditProcessCreation, auditProcessCreationCommand},
		{tweaks.HideEdgeFirstRun, hideEdgeFirstRunCommand},
		{tweaks.DisableEdgeStartupBoost, disableEdgeStartupBoostCommand},
		{tweaks.TurnOffSystemSounds, disableStartupSoundCommand},
		{tweaks.DisableAppSuggestions, disableAppSuggestionsCommand},
		{tweaks.PreventDeviceApps, preventDeviceAppsCommand},
		{tweaks.HardenSystemDriveACL, hardenSystemDriveACLCommand},
		{tweaks.DisableCoreIsolation, disableCoreIsolationCommand},
	}

	var commands []runSynchronousCommand
	for _, c := range enabledCommands {
		if c.enabled {
			commands = append(commands, newRunSynchronousCommand(len(commands)+1, c.command))
		}
	}
	if tweaks.MakeEdgeUninstallable {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, MakeEdgeUninstallableCommand()))
	}
	if tweaks.PreventAutomaticReboot {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, PreventAutomaticRebootCommand()))
	}
	if tweaks.TurnOffSystemSounds {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, TurnOffSystemSoundsDefaultUserCommand()))
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, TurnOffSystemSoundsUserOnceCommand()))
	}
	if tweaks.DisableAppSuggestions {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, DisableAppSuggestionsDefaultUserCommand()))
	}
	if tweaks.DisablePointerPrecision {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, DisablePointerPrecisionCommand()))
	}
	if tweaks.DeleteHiddenJunctions {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, DeleteJunctionsUserOnceCommand()))
	}
	if ContainsApp(removeApps, profile.AppOneDrive) {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, RemoveOneDriveFilesCommand()))
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, RemoveOneDriveDefaultUserCommand()))
	}
	if cmd := PasswordExpirationCommand(passwordExpiration); cmd != "" {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, cmd))
	}
	if cmd := AccountLockoutCommand(accountLockout); cmd != "" {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, cmd))
	}
	if cmd := FileExplorerCommand(fileExplorer); cmd != "" {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, cmd))
	}
	if cmd := PersonalizationCommand(personalization); cmd != "" {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, cmd))
	}
	if cmd := StickyKeysDefaultUserCommand(stickyKeys); cmd != "" {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, cmd))
	}
	if cmd := StickyKeysSystemDefaultCommand(stickyKeys); cmd != "" {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, cmd))
	}
	if cmd := LockKeyIndicatorsCommand(lockKeys); cmd != "" {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, cmd))
	}
	if cmd := LockKeyScancodeMapCommand(lockKeys); cmd != "" {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, cmd))
	}
	if cmd := DesktopIconsUserOnceCommand(desktopIcons); cmd != "" {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, cmd))
	}
	if cmd := StartFoldersUserOnceCommand(startFolders); cmd != "" {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, cmd))
	}
	if cmd := AppLockerCommand(appLockerPolicyXML); cmd != "" {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, cmd))
	}
	if useNarrator {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, NarratorSpecializeCommand()))
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, NarratorUserOnceCommand()))
	}
	if computerNameScript != nil {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, ComputerNameScriptCommand(*computerNameScript)))
	}
	for _, cmd := range SystemScriptsCommands(systemScripts) {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, cmd))
	}
	if cmd := DefaultUserScriptCommand(defaultUserScripts); cmd != "" {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, cmd))
	}
	if cmd := UserOnceScriptCommand(userOnceScripts); cmd != "" {
		commands = append(commands, newRunSynchronousCommand(len(commands)+1, cmd))
	}
	if len(commands) == 0 {
		return nil
	}
	return &Deployment{
		Name:           "Microsoft-Windows-Deployment",
		standardAttrs:  newStandardAttrs(),
		RunSynchronous: &runSynchronous{RunSynchronousCommand: commands},
	}
}

// password is the shared <Password> element used by local accounts, the
// built-in Administrator account and AutoLogon.
type password struct {
	Value     string `xml:"Value"`
	PlainText bool   `xml:"PlainText"`
}

// newPasswordElement builds a password element. When obscure is true, the
// value is Base64(UTF-16LE(raw+element)) and PlainText is false — the
// standard Windows unattend obscuration convention (Microsoft Learn:
// Microsoft-Windows-Shell-Setup | UserAccounts). This is obfuscation, not
// encryption: element is a fixed, publicly known salt ("Password" or
// "AdministratorPassword"), so an obscured value is exactly as recoverable
// as a plaintext one to anyone who knows the convention — it only keeps
// the password from being trivially grep-able in the raw XML.
func newPasswordElement(raw string, element string, obscure bool) *password {
	if !obscure {
		return &password{Value: raw, PlainText: true}
	}
	return &password{Value: base64.StdEncoding.EncodeToString(utf16LEBytes(raw + element)), PlainText: false}
}

func utf16LEBytes(s string) []byte {
	units := utf16.Encode([]rune(s))
	buf := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(buf[i*2:], u)
	}
	return buf
}

type localAccount struct {
	Action      string    `xml:"wcm:action,attr"`
	Password    *password `xml:"Password,omitempty"`
	DisplayName string    `xml:"DisplayName,omitempty"`
	Group       string    `xml:"Group"`
	Name        string    `xml:"Name"`
}

type localAccounts struct {
	LocalAccount []localAccount `xml:"LocalAccount"`
}

type userAccounts struct {
	LocalAccounts         *localAccounts `xml:"LocalAccounts,omitempty"`
	AdministratorPassword *password      `xml:"AdministratorPassword,omitempty"`
}

// autoLogon is Microsoft-Windows-Shell-Setup/AutoLogon. LogonCount is always
// 1: log on automatically once, then behave normally.
type autoLogon struct {
	Password   *password `xml:"Password,omitempty"`
	Enabled    bool      `xml:"Enabled"`
	LogonCount int       `xml:"LogonCount"`
	Username   string    `xml:"Username"`
}

// oobeBlock controls OOBE screens. ProtectYourPC (1 = apply recommended
// settings automatically i.e. telemetry on, 3 = turn automatic protection
// off) covers the express-settings prompt. The Hide* fields and
// NetworkLocation additionally suppress the EULA/OEM/online-account/
// wireless-setup screens so a non-interactive profile does not stop at one
// of them; see Microsoft-Windows-Shell-Setup | OOBE on learn.microsoft.com.
type oobeBlock struct {
	ProtectYourPC             int    `xml:"ProtectYourPC,omitempty"`
	HideEULAPage              bool   `xml:"HideEULAPage,omitempty"`
	HideOEMRegistrationScreen bool   `xml:"HideOEMRegistrationScreen,omitempty"`
	HideOnlineAccountScreens  bool   `xml:"HideOnlineAccountScreens,omitempty"`
	HideWirelessSetupInOOBE   bool   `xml:"HideWirelessSetupInOOBE,omitempty"`
	HideLocalAccountScreen    bool   `xml:"HideLocalAccountScreen,omitempty"`
	NetworkLocation           string `xml:"NetworkLocation,omitempty"`
}

// firstLogonCommands is Microsoft-Windows-Shell-Setup/FirstLogonCommands:
// commands run once, after the first user logs on. Used here to provision
// Wi-Fi, since the WLAN service is not guaranteed to be up any earlier.
type firstLogonCommands struct {
	SynchronousCommand []synchronousCommand `xml:"SynchronousCommand"`
}

type synchronousCommand struct {
	Action      string `xml:"wcm:action,attr"`
	CommandLine string `xml:"CommandLine"`
	Order       int    `xml:"Order"`
}

// ShellSetupOOBE is the Microsoft-Windows-Shell-Setup component in the
// oobeSystem pass, carrying local accounts, auto-logon, express settings and
// (via FirstLogonCommands) Wi-Fi provisioning.
type ShellSetupOOBE struct {
	XMLName xml.Name `xml:"component"`
	Name    string   `xml:"name,attr"`
	standardAttrs
	OOBE               *oobeBlock          `xml:"OOBE,omitempty"`
	UserAccounts       *userAccounts       `xml:"UserAccounts,omitempty"`
	AutoLogon          *autoLogon          `xml:"AutoLogon,omitempty"`
	FirstLogonCommands *firstLogonCommands `xml:"FirstLogonCommands,omitempty"`
}

// NewShellSetupOOBE builds the oobeSystem-pass component from accounts,
// firstLogon, express and wifi. It returns nil when there is nothing to
// configure.
func NewShellSetupOOBE(accounts []profile.UserAccount, firstLogon profile.FirstLogon, express profile.ExpressSettings, wifi *profile.WifiSettings, bypassOnlineAccountRequirement bool, removeApps []profile.RemovableApp, removeFeatures []profile.RemovableFeature, removeOptionalFeatures []profile.RemovableOptionalFeature, deleteHiddenJunctions bool, deleteWindowsOld bool, keepSensitiveFiles bool, installVMGuestTools []profile.VMGuestTool, obscurePasswords bool, firstLogonScripts []profile.CustomScript, restartExplorerAfterScripts bool) *ShellSetupOOBE {
	var ua *userAccounts
	if len(accounts) > 0 {
		ua = &userAccounts{LocalAccounts: &localAccounts{}}
		for _, a := range accounts {
			la := localAccount{
				Action: wcmActionAdd,
				Group:  string(a.Group),
				Name:   a.Name,
			}
			if a.DisplayName != nil {
				la.DisplayName = *a.DisplayName
			}
			if a.Password != nil {
				la.Password = newPasswordElement(*a.Password, "Password", obscurePasswords)
			}
			ua.LocalAccounts.LocalAccount = append(ua.LocalAccounts.LocalAccount, la)
		}
	}

	var al *autoLogon
	switch firstLogon.Mode {
	case profile.FirstLogonBuiltinAdmin:
		pwd := ""
		if firstLogon.BuiltinAdministratorPassword != nil {
			pwd = *firstLogon.BuiltinAdministratorPassword
		}
		if ua == nil {
			ua = &userAccounts{}
		}
		ua.AdministratorPassword = newPasswordElement(pwd, "AdministratorPassword", obscurePasswords)
		al = &autoLogon{
			Enabled:    true,
			LogonCount: 1,
			Username:   adminUsername,
			Password:   newPasswordElement(pwd, "Password", obscurePasswords),
		}
	case profile.FirstLogonFirstCreatedAccount:
		if len(accounts) > 0 {
			first := accounts[0]
			al = &autoLogon{Enabled: true, LogonCount: 1, Username: first.Name}
			if first.Password != nil {
				al.Password = newPasswordElement(*first.Password, "Password", obscurePasswords)
			}
		}
	case profile.FirstLogonNone:
		// no AutoLogon element
	}

	oobe := &oobeBlock{}
	switch express.Mode {
	case profile.ExpressAllEnabled:
		oobe.ProtectYourPC = 1
	case profile.ExpressAllDisabled:
		oobe.ProtectYourPC = 3
	case profile.ExpressInteractive:
		// ProtectYourPC left unset: Windows Setup asks interactively
	}
	if express.Mode != profile.ExpressInteractive {
		oobe.HideEULAPage = true
		oobe.HideOEMRegistrationScreen = true
		oobe.HideLocalAccountScreen = true
		oobe.HideWirelessSetupInOOBE = true
		oobe.NetworkLocation = "Work"
	}
	// A local account defined in the profile is, in practice, the reliable
	// way to skip the Microsoft-account requirement; BypassOnlineAccountRequirement
	// covers the case where no account is predefined either.
	if len(accounts) > 0 || bypassOnlineAccountRequirement {
		oobe.HideOnlineAccountScreens = true
	}
	if *oobe == (oobeBlock{}) {
		oobe = nil
	}

	var flCommands []synchronousCommand
	if cmd := WifiFirstLogonCommand(wifi); cmd != "" {
		flCommands = append(flCommands, synchronousCommand{Action: wcmActionAdd, Order: len(flCommands) + 1, CommandLine: cmd})
	}
	if cmd := RemoveAppsFirstLogonCommand(removeApps); cmd != "" {
		flCommands = append(flCommands, synchronousCommand{Action: wcmActionAdd, Order: len(flCommands) + 1, CommandLine: cmd})
	}
	if cmd := RemoveFeaturesFirstLogonCommand(removeFeatures); cmd != "" {
		flCommands = append(flCommands, synchronousCommand{Action: wcmActionAdd, Order: len(flCommands) + 1, CommandLine: cmd})
	}
	if cmd := RemoveOptionalFeaturesFirstLogonCommand(removeOptionalFeatures); cmd != "" {
		flCommands = append(flCommands, synchronousCommand{Action: wcmActionAdd, Order: len(flCommands) + 1, CommandLine: cmd})
	}
	if deleteHiddenJunctions {
		flCommands = append(flCommands, synchronousCommand{Action: wcmActionAdd, Order: len(flCommands) + 1, CommandLine: DeleteJunctionsFirstLogonCommand()})
	}
	if deleteWindowsOld {
		flCommands = append(flCommands, synchronousCommand{Action: wcmActionAdd, Order: len(flCommands) + 1, CommandLine: deleteWindowsOldCommand})
	}
	for _, cmd := range VMGuestToolsFirstLogonCommands(installVMGuestTools) {
		flCommands = append(flCommands, synchronousCommand{Action: wcmActionAdd, Order: len(flCommands) + 1, CommandLine: cmd})
	}
	for _, cmd := range FirstLogonScriptsCommands(firstLogonScripts) {
		flCommands = append(flCommands, synchronousCommand{Action: wcmActionAdd, Order: len(flCommands) + 1, CommandLine: cmd})
	}
	if cmd := KeepSensitiveFilesFirstLogonCommand(keepSensitiveFiles); cmd != "" {
		flCommands = append(flCommands, synchronousCommand{Action: wcmActionAdd, Order: len(flCommands) + 1, CommandLine: cmd})
	}
	if restartExplorerAfterScripts && len(flCommands) > 0 {
		flCommands = append(flCommands, synchronousCommand{Action: wcmActionAdd, Order: len(flCommands) + 1, CommandLine: RestartExplorerCommand()})
	}
	var flc *firstLogonCommands
	if len(flCommands) > 0 {
		flc = &firstLogonCommands{SynchronousCommand: flCommands}
	}

	if ua == nil && al == nil && oobe == nil && flc == nil {
		return nil
	}
	return &ShellSetupOOBE{
		Name:               shellSetupName,
		standardAttrs:      newStandardAttrs(),
		OOBE:               oobe,
		UserAccounts:       ua,
		AutoLogon:          al,
		FirstLogonCommands: flc,
	}
}
