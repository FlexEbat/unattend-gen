package components

import (
	"fmt"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// Slice 22 (tech.md backlog group E): small setup-stage mechanisms found
// missing during the schneegans.de audit. Mechanisms sourced from the
// reference implementation (github.com/cschneegans/unattend-generator,
// modifier/Delete.cs, modifier/Accessibility.cs, modifier/ComputerName.cs,
// resource/SetComputerName.ps1), not invented from memory.

// keepSensitiveFilesScript deletes the answer file Windows Setup copies to
// C:\Windows\Panther for its own logging/replay purposes — the file that
// contains every setting in this profile, including plaintext passwords
// unless ObscurePasswords is set. Matches the reference implementation's
// default: sensitive files ARE deleted unless KeepSensitiveFiles is set to
// keep them (the one field in this schema whose zero value causes an
// action, rather than "changes nothing" — a deliberate exception, since
// silently leaving plaintext credentials on disk is the worse default).
// Unlike the reference, there is no C:\Windows\Setup\Scripts\Wifi.xml to
// delete here: our Wi-Fi mechanism (wlan.go) writes its profile to %TEMP%,
// not that path, so that file is targeted instead.
const keepSensitiveFilesScript = `@(
	'C:\Windows\Panther\unattend.xml';
	'C:\Windows\Panther\unattend-original.xml';
	"$env:TEMP\wifi-profile.xml";
) | Where-Object -FilterScript { [System.IO.File]::Exists( $_ ); } | Remove-Item -Force -Verbose -ErrorAction 'Continue';
`

// KeepSensitiveFilesFirstLogonCommand returns one oobeSystem/
// FirstLogonCommands entry that deletes the files above, or "" when
// keepSensitiveFiles is true.
func KeepSensitiveFilesFirstLogonCommand(keepSensitiveFiles bool, hidePowerShellWindows bool) string {
	if keepSensitiveFiles {
		return ""
	}
	path := scriptsDir + `\unattend-delete-sensitive-files.ps1`
	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		writeFileStatement(path, []byte(keepSensitiveFilesScript)),
		invokeCommand(profile.ScriptPs1, path, hidePowerShellWindows),
	})
}

// narratorWindowsPECommand launches Narrator during the windowsPE pass
// itself, so it's available even while Setup's own UI is showing.
const narratorWindowsPECommand = `cmd.exe /c "start X:\Windows\System32\Narrator.exe"`

// NarratorSpecializeCommand returns one specialize-pass command that
// launches Narrator and sets its "last used" configuration to the
// narrator profile, for the account created during setup.
func NarratorSpecializeCommand() string {
	return wrapCommand([]string{
		`& 'C:\Windows\System32\Narrator.exe'`,
		`reg.exe ADD "HKLM\Software\Microsoft\Windows NT\CurrentVersion\Accessibility" /v Configuration /t REG_SZ /d narrator /f`,
	})
}

// NarratorUserOnceCommand returns one specialize-pass command that mounts
// the default user hive just long enough to register a RunOnce entry;
// that entry launches Narrator and sets the same configuration value in
// the live account's own HKCU at first logon — the same RunOnce-to-live-
// HKCU mechanism DesktopIconsUserOnceCommand uses, so this applies to
// every future account too, not just the one created during setup.
func NarratorUserOnceCommand() string {
	scriptContent := "@echo off\r\n" +
		`start "" "C:\Windows\System32\Narrator.exe"` + "\r\n" +
		`reg.exe ADD "HKCU\Software\Microsoft\Windows NT\CurrentVersion\Accessibility" /v Configuration /t REG_SZ /d narrator /f` + "\r\n"
	path := scriptsDir + `\unattend-narrator-uo.cmd`
	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		writeFileStatement(path, []byte(scriptContent)),
		fmt.Sprintf(`reg.exe add "%s" /v UnattendNarrator /d "%s" /f`, defaultUserRunOnceKey, path),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}

// setComputerNameScript polls the registry every 50ms forever, re-applying
// the new computer name to ComputerName/Hostname/NV Hostname. It has to
// keep re-applying rather than set the values once, because Windows itself
// overwrites them during specialize after this script starts (a known
// quirk the reference implementation works around this same way). Copied
// verbatim from resource/SetComputerName.ps1 — the file paths inside
// already match this project's scriptsDir (C:\Windows\Setup\Scripts), no
// changes needed.
const setComputerNameScript = `$ErrorActionPreference = 'Stop';
Set-StrictMode -Version 'Latest';
& {
	$newName = ( Get-Content -LiteralPath 'C:\Windows\Setup\Scripts\ComputerName.txt' -Raw ).Trim();
	if( [string]::IsNullOrWhitespace( $newName ) ) {
		throw "No computer name was provided.";
	}

	$keys = @(
		@{
			LiteralPath = 'Registry::HKLM\SYSTEM\CurrentControlSet\Control\ComputerName\ComputerName';
			Name = 'ComputerName';
		};
		@{
			LiteralPath = 'Registry::HKLM\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters';
			Name = 'Hostname';
		};
		@{
			LiteralPath = 'Registry::HKLM\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters';
			Name = 'NV Hostname';
		};
	);

	while( $true ) {
		foreach( $key in $keys ) {
			Set-ItemProperty @key -Type 'String' -Value $newName;
		}
		Start-Sleep -Milliseconds 50;
	}
} *>&1 | Out-String -Width 1KB -Stream >> 'C:\Windows\Setup\Scripts\SetComputerName.log';`

// ComputerNameScriptCommand returns one specialize-pass command that runs
// the user-supplied script to compute a name, writes it to a file, then
// spawns a hidden background PowerShell process (setComputerNameScript)
// that keeps applying it to the registry. Mutually exclusive with a static
// Profile.ComputerName (enforced in validate.go); the ComputerName element
// itself is set to a "TEMPNAME" placeholder by the caller, since Windows
// Setup requires some value there and this script overwrites it moments
// later anyway.
func ComputerNameScriptCommand(script string) string {
	getterPath := scriptsDir + `\unattend-get-computername.ps1`
	setterPath := scriptsDir + `\unattend-set-computername.ps1`
	namePath := scriptsDir + `\ComputerName.txt`
	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		writeFileStatement(getterPath, []byte(script)),
		writeFileStatement(setterPath, []byte(setComputerNameScript)),
		fmt.Sprintf(`[string] $newName = & "%s"`, getterPath),
		fmt.Sprintf(`$newName > "%s"`, namePath),
		`"Will set the computer name to '${newName}'."`,
		`Start-Process -FilePath ( Get-Process -Id $PID ).Path -ArgumentList '-ExecutionPolicy "Unrestricted" -NoProfile -File "` + setterPath + `"' -WindowStyle 'Hidden'`,
		`Start-Sleep -Seconds 10`,
	})
}
