package components

import (
	"fmt"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// Slice 16 (Group B of the tech.md backlog): six SystemTweaks that were
// missing against schneegans.de/windows/unattend-generator. Mechanisms
// sourced from the reference implementation
// (github.com/cschneegans/unattend-generator, modifier/Optimizations.cs),
// not invented from memory, per the schema-accuracy discipline in tech.md
// section 12.

// Simple specialize-pass commands: single reg.exe invocations, folded into
// the same enabledCommands list as the existing 17 tweaks in NewDeployment.
const (
	disableStartupSoundCommand   = `cmd.exe /c reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Authentication\LogonUI\BootAnimation" /v DisableStartupSound /t REG_DWORD /d 1 /f && reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\EditionOverrides" /v UserSetting_DisableStartupSound /t REG_DWORD /d 1 /f`
	disableAppSuggestionsCommand = `cmd.exe /c reg add "HKLM\Software\Policies\Microsoft\Windows\CloudContent" /v DisableWindowsConsumerFeatures /t REG_DWORD /d 1 /f`
	preventDeviceAppsCommand     = `cmd.exe /c reg add "HKLM\Software\Microsoft\Windows\CurrentVersion\Device Metadata" /v PreventDeviceMetadataFromNetwork /t REG_DWORD /d 1 /f && reg add "HKLM\Software\Microsoft\Windows\CurrentVersion\Device Installer" /v DisableCoInstallers /t REG_DWORD /d 1 /f`
)

// moveActiveHoursTaskXML is the scheduled-task definition that periodically
// moves Windows Update's "active hours" window to the current time, so
// Windows keeps believing the device is in use and never reboots
// unattended. Copied verbatim (BOM stripped) from the reference
// implementation's resource/MoveActiveHours.xml.
const moveActiveHoursTaskXML = `<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
	<Triggers>
		<BootTrigger>
			<Repetition>
				<Interval>PT4H</Interval>
				<StopAtDurationEnd>false</StopAtDurationEnd>
			</Repetition>
			<Enabled>true</Enabled>
		</BootTrigger>
		<RegistrationTrigger>
			<Repetition>
				<Interval>PT4H</Interval>
				<StopAtDurationEnd>false</StopAtDurationEnd>
			</Repetition>
			<Enabled>true</Enabled>
		</RegistrationTrigger>
	</Triggers>
	<Principals>
		<Principal id="Author">
			<UserId>S-1-5-19</UserId>
			<RunLevel>LeastPrivilege</RunLevel>
		</Principal>
	</Principals>
	<Settings>
		<MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy>
		<DisallowStartIfOnBatteries>false</DisallowStartIfOnBatteries>
		<StopIfGoingOnBatteries>false</StopIfGoingOnBatteries>
		<AllowHardTerminate>true</AllowHardTerminate>
		<StartWhenAvailable>false</StartWhenAvailable>
		<RunOnlyIfNetworkAvailable>false</RunOnlyIfNetworkAvailable>
		<IdleSettings>
			<StopOnIdleEnd>true</StopOnIdleEnd>
			<RestartOnIdle>false</RestartOnIdle>
		</IdleSettings>
		<AllowStartOnDemand>true</AllowStartOnDemand>
		<Enabled>true</Enabled>
		<Hidden>false</Hidden>
		<RunOnlyIfIdle>false</RunOnlyIfIdle>
		<WakeToRun>false</WakeToRun>
		<ExecutionTimeLimit>PT72H</ExecutionTimeLimit>
		<Priority>7</Priority>
	</Settings>
	<Actions Context="Author">
		<Exec>
			<Command>%windir%\System32\conhost.exe</Command>
			<Arguments>--headless %windir%\System32\WindowsPowerShell\v1.0\powershell.exe -WindowStyle Hidden -NoProfile -NonInteractive -Command "$p = @{ LiteralPath = 'Registry::HKLM\Software\Microsoft\WindowsUpdate\UX\Settings'; Type = 'DWord'; }; $h = [datetime]::Now.Hour; Set-ItemProperty @p -Name 'ActiveHoursStart' -Value (($h + 23) % 24); Set-ItemProperty @p -Name 'ActiveHoursEnd' -Value (($h + 11) % 24); Set-ItemProperty @p -Name 'SmartActiveHoursState' -Value 0;"</Arguments>
		</Exec>
	</Actions>
</Task>`

// PreventAutomaticRebootCommand returns one specialize-pass command that
// stops Windows Update from rebooting an in-use machine: it disables the
// auto-reboot-with-logged-on-users policy and registers a scheduled task
// that keeps nudging "active hours" to the current time.
func PreventAutomaticRebootCommand() string {
	xmlPath := scriptsDir + `\unattend-move-active-hours.xml`
	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		writeFileStatement(xmlPath, []byte(moveActiveHoursTaskXML)),
		`reg.exe add "HKLM\SOFTWARE\Policies\Microsoft\Windows\WindowsUpdate\AU" /v AUOptions /t REG_DWORD /d 4 /f`,
		`reg.exe add "HKLM\SOFTWARE\Policies\Microsoft\Windows\WindowsUpdate\AU" /v NoAutoRebootWithLoggedOnUsers /t REG_DWORD /d 1 /f`,
		fmt.Sprintf(`Register-ScheduledTask -TaskName 'MoveActiveHours' -Xml $( Get-Content -LiteralPath "%s" -Raw ) -Force | Out-Null`, xmlPath),
	})
}

// deleteJunctionsFirstLogonScript removes hidden NTFS junction points
// (ReparsePoint items) left by Windows for legacy-compatibility, such as
// C:\Documents and Settings. Runs once at first logon (the account created
// during setup already exists by then).
const deleteJunctionsFirstLogonScript = `@(
	Get-ChildItem -LiteralPath 'C:\' -Force;
	Get-ChildItem -LiteralPath 'C:\Users' -Force;
	Get-ChildItem -LiteralPath 'C:\Users\Default' -Force -Recurse -Depth 2;
	Get-ChildItem -LiteralPath 'C:\Users\Public' -Force -Recurse -Depth 2;
	Get-ChildItem -LiteralPath 'C:\ProgramData' -Force;
) | Where-Object -FilterScript {
	$_.Attributes.HasFlag( [System.IO.FileAttributes]::ReparsePoint );
} | Remove-Item -Force -Recurse -Verbose;
`

// deleteJunctionsUserOnceScript is the per-account counterpart: it cleans
// up junctions under the profile directory of every account that logs on
// for the first time, including ones created after setup.
const deleteJunctionsUserOnceScript = `@(
	Get-ChildItem -LiteralPath $env:USERPROFILE -Force -Recurse -Depth 2;
) | Where-Object -FilterScript {
	$_.Attributes.HasFlag( [System.IO.FileAttributes]::ReparsePoint );
} | Remove-Item -Force -Recurse -Verbose;
`

// DeleteJunctionsFirstLogonCommand returns one command for the
// oobeSystem/FirstLogonCommands list (same family as Wi-Fi/apps/features),
// removing junction points visible to the account created during setup.
func DeleteJunctionsFirstLogonCommand() string {
	path := scriptsDir + `\unattend-delete-junctions.ps1`
	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		writeFileStatement(path, []byte(deleteJunctionsFirstLogonScript)),
		invokeCommand(profile.ScriptPs1, path),
	})
}

// DeleteJunctionsUserOnceCommand returns one specialize-pass command that
// mounts the default user hive and registers a RunOnce entry — the same
// mechanism UserOnceScriptCommand uses for user-supplied scripts — so every
// future account also gets its junctions cleaned up on first logon.
func DeleteJunctionsUserOnceCommand() string {
	scriptPath := scriptsDir + `\unattend-delete-junctions-uo.ps1`
	wrapperPath := scriptsDir + `\unattend-delete-junctions-uo-run.cmd`
	wrapperContent := "@echo off\r\n" + invokeCommand(profile.ScriptPs1, scriptPath) + "\r\n"
	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		writeFileStatement(scriptPath, []byte(deleteJunctionsUserOnceScript)),
		writeFileStatement(wrapperPath, []byte(wrapperContent)),
		fmt.Sprintf(`reg.exe add "%s" /v UnattendDeleteJunctions /d "%s" /f`, defaultUserRunOnceKey, wrapperPath),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}

// turnOffSystemSoundsDefaultUserScript clears every app event's sound
// scheme in the default user hive, except events whose EventLabel is
// marked ExcludeFromCPL (this replicates the exact filter Windows' own
// "No Sounds" scheme selection uses, so it doesn't silence system-critical
// alerts that are excluded from that picker for a reason).
const turnOffSystemSoundsDefaultUserScript = `$excludes = Get-ChildItem -LiteralPath 'Registry::HKU\DefaultUser\AppEvents\EventLabels' |
    Where-Object -FilterScript { ($_ | Get-ItemProperty).ExcludeFromCPL -eq 1; } |
    Select-Object -ExpandProperty 'PSChildName';
Get-ChildItem -Path 'Registry::HKU\DefaultUser\AppEvents\Schemes\Apps\*\*' |
    Where-Object -Property 'PSChildName' -NotIn $excludes |
    Get-ChildItem -Include '.Current' | Set-ItemProperty -Name '(Default)' -Value '';
`

// TurnOffSystemSoundsDefaultUserCommand returns one specialize-pass command
// that mounts the default user hive and clears app event sounds there, so
// every future account starts with sounds off.
func TurnOffSystemSoundsDefaultUserCommand() string {
	path := scriptsDir + `\unattend-system-sounds.ps1`
	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		writeFileStatement(path, []byte(turnOffSystemSoundsDefaultUserScript)),
		invokeCommand(profile.ScriptPs1, path),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}

// TurnOffSystemSoundsUserOnceCommand returns one specialize-pass command
// that registers a RunOnce entry setting the live HKCU sound scheme to
// ".None" — the DefaultUser command above only affects the template new
// accounts are created from, not an account's own scheme selector, which
// still needs to be flipped once at that account's first logon.
func TurnOffSystemSoundsUserOnceCommand() string {
	const script = `Set-ItemProperty -LiteralPath 'Registry::HKCU\AppEvents\Schemes' -Name '(Default)' -Type 'String' -Value '.None';`
	scriptPath := scriptsDir + `\unattend-system-sounds-uo.ps1`
	wrapperPath := scriptsDir + `\unattend-system-sounds-uo-run.cmd`
	wrapperContent := "@echo off\r\n" + invokeCommand(profile.ScriptPs1, scriptPath) + "\r\n"
	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		writeFileStatement(scriptPath, []byte(script)),
		writeFileStatement(wrapperPath, []byte(wrapperContent)),
		fmt.Sprintf(`reg.exe add "%s" /v UnattendSystemSounds /d "%s" /f`, defaultUserRunOnceKey, wrapperPath),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}

// contentDeliveryManagerValues are the registry values that gate
// Windows' silent suggested-apps download/installation (Content Delivery
// Manager). Sourced from the reference implementation, which cites
// https://skanthak.hier-im-netz.de/ten.html#eight.
var contentDeliveryManagerValues = []string{
	"ContentDeliveryAllowed",
	"FeatureManagementEnabled",
	"OEMPreInstalledAppsEnabled",
	"PreInstalledAppsEnabled",
	"PreInstalledAppsEverEnabled",
	"SilentInstalledAppsEnabled",
	"SoftLandingEnabled",
	"SubscribedContentEnabled",
	"SubscribedContent-310093Enabled",
	"SubscribedContent-338387Enabled",
	"SubscribedContent-338388Enabled",
	"SubscribedContent-338389Enabled",
	"SubscribedContent-338393Enabled",
	"SubscribedContent-353694Enabled",
	"SubscribedContent-353696Enabled",
	"SubscribedContent-353698Enabled",
	"SystemPaneSuggestionsEnabled",
}

// DisableAppSuggestionsDefaultUserCommand returns one specialize-pass
// command that mounts the default user hive and zeroes every Content
// Delivery Manager value, so future accounts don't get suggested apps
// silently installed. The specialize-only DisableWindowsConsumerFeatures
// policy (folded into NewDeployment's simple command list) covers the
// system-wide half of this tweak.
func DisableAppSuggestionsDefaultUserCommand() string {
	key := defaultUserHiveKey + `\Software\Microsoft\Windows\CurrentVersion\ContentDeliveryManager`
	statements := []string{fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath)}
	for _, name := range contentDeliveryManagerValues {
		statements = append(statements, fmt.Sprintf(`reg.exe add "%s" /v %s /t REG_DWORD /d 0 /f`, key, name))
	}
	statements = append(statements, fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey))
	return wrapCommand(statements)
}

// DisablePointerPrecisionCommand returns one specialize-pass command that
// mounts the default user hive and zeroes the mouse-acceleration curve
// (MouseSpeed/MouseThreshold1/MouseThreshold2), disabling "Enhance pointer
// precision" for future accounts. Values are REG_SZ "0", matching the
// reference implementation's -Type 'String' -Value 0.
func DisablePointerPrecisionCommand() string {
	key := defaultUserHiveKey + `\Control Panel\Mouse`
	return wrapCommand([]string{
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		fmt.Sprintf(`reg.exe add "%s" /v MouseSpeed /t REG_SZ /d 0 /f`, key),
		fmt.Sprintf(`reg.exe add "%s" /v MouseThreshold1 /t REG_SZ /d 0 /f`, key),
		fmt.Sprintf(`reg.exe add "%s" /v MouseThreshold2 /t REG_SZ /d 0 /f`, key),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}
