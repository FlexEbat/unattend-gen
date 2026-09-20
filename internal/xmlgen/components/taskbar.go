package components

import (
	"fmt"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// Slice 26 (Start menu/taskbar, outside the original 6 audit groups).
// Mechanisms sourced from the reference implementation
// (github.com/cschneegans/unattend-generator, modifier/Optimizations.cs
// and resource/ShowAllTrayIcons.*), not invented from memory. The most
// elaborate sub-feature of this section — custom pinned taskbar icons via
// a locked Start layout XML + scheduled-task-based unlock flow — is
// deliberately NOT included here; see tech.md backlog.

// Simple specialize/DefaultUser-hive tweaks, folded into the same
// enabledCommands-style lists their SystemTweaks siblings already use.
const (
	disableWidgetsCommand = `cmd.exe /c reg add "HKLM\SOFTWARE\Policies\Microsoft\Dsh" /v AllowNewsAndInterests /t REG_DWORD /d 0 /f`
)

// leftTaskbarCommand, hideTaskViewButtonCommand, disableBingResultsCommand
// all touch the default-user hive (per-account Explorer\Advanced values,
// not a machine-wide policy like DisableWidgets), so each gets its own
// load/unload cycle — same pattern as the rest of optimizations.go.
func leftTaskbarCommand() string {
	key := defaultUserHiveKey + `\Software\Microsoft\Windows\CurrentVersion\Explorer\Advanced`
	return wrapCommand([]string{
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		fmt.Sprintf(`reg.exe add "%s" /v TaskbarAl /t REG_DWORD /d 0 /f`, key),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}

func hideTaskViewButtonCommand() string {
	key := defaultUserHiveKey + `\Software\Microsoft\Windows\CurrentVersion\Explorer\Advanced`
	return wrapCommand([]string{
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		fmt.Sprintf(`reg.exe add "%s" /v ShowTaskViewButton /t REG_DWORD /d 0 /f`, key),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}

func disableBingResultsCommand() string {
	key := defaultUserHiveKey + `\Software\Policies\Microsoft\Windows\Explorer`
	return wrapCommand([]string{
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		fmt.Sprintf(`reg.exe add "%s" /v DisableSearchBoxSuggestions /t REG_DWORD /d 1 /f`, key),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}

// showAllTrayIconsKeyWin10Statement disables the "hide inactive icons"
// auto-tray behavior directly via the default-user hive (Windows 10 only
// — this registry value has no effect on Windows 11's redesigned tray).
const showAllTrayIconsKeyWin10Statement = `reg.exe add "` + defaultUserHiveKey + `\Software\Microsoft\Windows\CurrentVersion\Explorer" /v EnableAutoTray /t REG_DWORD /d 0 /f`

// showAllTrayIconsTaskXML registers a scheduled task that runs at every
// logon (Windows 11 only — Windows 10 doesn't need it, see above) and
// promotes every notification-area icon to always-visible. Copied
// verbatim from resource/ShowAllTrayIcons.xml.
const showAllTrayIconsTaskXML = `<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
	<Triggers>
		<LogonTrigger>
			<Repetition>
				<Interval>PT1M</Interval>
				<StopAtDurationEnd>false</StopAtDurationEnd>
			</Repetition>
			<Enabled>true</Enabled>
		</LogonTrigger>
	</Triggers>
	<Principals>
		<Principal id="Author">
			<GroupId>S-1-5-32-545</GroupId>
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
			<Arguments>--headless %windir%\System32\WindowsPowerShell\v1.0\powershell.exe -WindowStyle Hidden -NoProfile -NonInteractive -Command "Set-ItemProperty -Path 'Registry::HKCU\Control Panel\NotifyIconSettings\*' -Name 'IsPromoted' -Value 1 -Type 'DWord';"</Arguments>
		</Exec>
	</Actions>
</Task>`

// ShowAllTrayIconsCommand returns one specialize-pass command that mounts
// the default user hive and branches by OS build (the reference's own
// approach — Windows 10 and 11 need genuinely different mechanisms here):
// on Windows 10, a direct registry value; on Windows 11, a per-logon
// scheduled task registered via a small bootstrap script. Applies to every
// future account.
func ShowAllTrayIconsCommand(hidePowerShellWindows bool) string {
	xmlPath := scriptsDir + `\unattend-tray-icons-task.xml`
	scriptPath := scriptsDir + `\unattend-show-all-tray-icons.ps1`
	script := `if( [System.Environment]::OSVersion.Version.Build -lt 20000 ) {
	` + showAllTrayIconsKeyWin10Statement + `;
} else {
	Register-ScheduledTask -TaskName 'ShowAllTrayIcons' -Xml $( Get-Content -LiteralPath "` + xmlPath + `" -Raw );
}
`
	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		writeFileStatement(xmlPath, []byte(showAllTrayIconsTaskXML)),
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		writeFileStatement(scriptPath, []byte(script)),
		invokeCommand(profile.ScriptPs1, scriptPath, hidePowerShellWindows),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}

// TaskbarSearchUserOnceCommand returns one specialize-pass command that
// mounts the default user hive just long enough to register a RunOnce
// entry; that entry sets the live account's own SearchboxTaskbarMode and
// restarts Explorer at first logon — same RunOnce-to-live-HKCU mechanism
// DesktopIconsUserOnceCommand uses. Returns "" for
// TaskbarSearchModeBox/"" (Windows' own default).
func TaskbarSearchUserOnceCommand(mode profile.TaskbarSearchMode, hidePowerShellWindows bool) string {
	if mode == "" || mode == profile.TaskbarSearchModeBox {
		return ""
	}
	values := map[profile.TaskbarSearchMode]int{
		profile.TaskbarSearchModeHide:  0,
		profile.TaskbarSearchModeIcon:  1,
		profile.TaskbarSearchModeBox:   2,
		profile.TaskbarSearchModeLabel: 3,
	}
	script := fmt.Sprintf(`Set-ItemProperty -LiteralPath 'Registry::HKCU\Software\Microsoft\Windows\CurrentVersion\Search' -Name 'SearchboxTaskbarMode' -Type 'DWord' -Value %d -Force;
Stop-Process -Name explorer -Force -ErrorAction SilentlyContinue;
Start-Process explorer.exe;`, values[mode])
	scriptPath := scriptsDir + `\unattend-taskbar-search-uo.ps1`
	wrapperPath := scriptsDir + `\unattend-taskbar-search-uo-run.cmd`
	wrapperContent := "@echo off\r\n" + invokeCommand(profile.ScriptPs1, scriptPath, hidePowerShellWindows) + "\r\n"
	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		writeFileStatement(scriptPath, []byte(script)),
		writeFileStatement(wrapperPath, []byte(wrapperContent)),
		fmt.Sprintf(`reg.exe add "%s" /v UnattendTaskbarSearch /d "%s" /f`, defaultUserRunOnceKey, wrapperPath),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}

const setStartPinsScriptTemplate = `if( [System.Environment]::OSVersion.Version.Build -lt 20000 ) {
	return;
}
$json = '%s';
$key = 'Registry::HKLM\SOFTWARE\Microsoft\PolicyManager\current\device\Start';
New-Item -Path $key -ItemType 'Directory' -ErrorAction 'SilentlyContinue';
Set-ItemProperty -LiteralPath $key -Name 'ConfigureStartPins' -Value $json -Type 'String';
`

// StartPinsCommand returns one specialize-pass command that sets the
// Windows 11 Start menu's pinned-apps policy (no effect on Windows 10).
// Returns "" for StartPinsModeDefault/"".
func StartPinsCommand(s profile.StartPinsSettings, hidePowerShellWindows bool) string {
	var jsonValue string
	switch s.Mode {
	case profile.StartPinsModeEmpty:
		jsonValue = `{"pinnedList":[]}`
	case profile.StartPinsModeCustom:
		if s.JSON == nil {
			return ""
		}
		jsonValue = *s.JSON
	default:
		return ""
	}
	path := scriptsDir + `\unattend-set-start-pins.ps1`
	script := fmt.Sprintf(setStartPinsScriptTemplate, psSingleQuoteEscape(jsonValue))
	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		writeFileStatement(path, []byte(script)),
		invokeCommand(profile.ScriptPs1, path, hidePowerShellWindows),
	})
}

// psSingleQuoteEscape doubles single quotes, the PowerShell single-quoted
// string escaping convention — needed because the JSON payload is embedded
// inside a PowerShell '...' literal in setStartPinsScriptTemplate.
func psSingleQuoteEscape(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\'' {
			out = append(out, '\'', '\'')
		} else {
			out = append(out, s[i])
		}
	}
	return string(out)
}

const startTilesEmptyXML = `<LayoutModificationTemplate Version='1' xmlns='http://schemas.microsoft.com/Start/2014/LayoutModification'>
	<LayoutOptions StartTileGroupCellWidth='6' />
	<DefaultLayoutOverride>
		<StartLayoutCollection>
			<StartLayout GroupCellWidth='6' xmlns='http://schemas.microsoft.com/Start/2014/FullDefaultLayout' />
		</StartLayoutCollection>
	</DefaultLayoutOverride>
</LayoutModificationTemplate>
`

// StartTilesCommand returns one specialize-pass command that writes
// LayoutModification.xml straight to the default profile's AppData (a
// plain file write, no registry hive mount needed — the path exists as
// soon as the image is applied). Affects the Windows 10 Start menu tile
// layout for every future account (no effect on Windows 11, which uses
// StartPinsSettings instead). Returns "" for StartTilesModeDefault/"".
func StartTilesCommand(s profile.StartTilesSettings) string {
	var xmlContent string
	switch s.Mode {
	case profile.StartTilesModeEmpty:
		xmlContent = startTilesEmptyXML
	case profile.StartTilesModeCustom:
		if s.XML == nil {
			return ""
		}
		xmlContent = *s.XML
	default:
		return ""
	}
	path := `C:\Users\Default\AppData\Local\Microsoft\Windows\Shell\LayoutModification.xml`
	return wrapCommand([]string{
		writeFileStatement(path, []byte(xmlContent)),
	})
}
