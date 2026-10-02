package components

import (
	"fmt"
	"strings"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// Mechanisms sourced from the reference implementation
// (github.com/cschneegans/unattend-generator, modifier/Optimizations.cs
// and resource/ShowAllTrayIcons.*), not invented from memory. The most
// elaborate sub-feature of this section - custom pinned taskbar icons via
// a locked Start layout XML + scheduled-task-based unlock flow - was added
// separately in .

// Simple specialize/DefaultUser-hive tweaks, folded into the same
// enabledCommands-style lists their SystemTweaks siblings already use.
const (
	disableWidgetsCommand = `cmd.exe /c reg add "HKLM\SOFTWARE\Policies\Microsoft\Dsh" /v AllowNewsAndInterests /t REG_DWORD /d 0 /f`
)

// leftTaskbarCommand, hideTaskViewButtonCommand, disableBingResultsCommand
// all touch the default-user hive (per-account Explorer\Advanced values,
// not a machine-wide policy like DisableWidgets), so each gets its own
// load/unload cycle - same pattern as the rest of optimizations.go.
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
// - this registry value has no effect on Windows 11's redesigned tray).
const showAllTrayIconsKeyWin10Statement = `reg.exe add "` + defaultUserHiveKey + `\Software\Microsoft\Windows\CurrentVersion\Explorer" /v EnableAutoTray /t REG_DWORD /d 0 /f`

// showAllTrayIconsTaskXML registers a scheduled task that runs at every
// logon (Windows 11 only - Windows 10 doesn't need it, see above) and
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
// approach - Windows 10 and 11 need genuinely different mechanisms here):
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
// restarts Explorer at first logon - same RunOnce-to-live-HKCU mechanism
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
// string escaping convention - needed because the JSON payload is embedded
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
// plain file write, no registry hive mount needed - the path exists as
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

// Custom pinned taskbar icons. Mechanism sourced from the
// reference implementation (github.com/cschneegans/unattend-generator,
// modifier/Optimizations.cs SetTaskbarIcons and resource/
// UnlockStartLayout.{vbs,xml}, resource/TaskbarLayout.xsd), not invented
// from memory. A taskbar layout can only be applied through a *locked*
// Start layout (LockedStartLayout=1 + StartLayoutFile policy in the
// default user hive). A locked layout would stay locked forever, so the
// reference adds an unlock flow: at first logon each account writes an
// Application event-log entry (source UnattendGenerator, event ID 1); a
// SYSTEM scheduled task subscribed to that event runs a VBScript that
// sets LockedStartLayout=0 for every loaded user hive, after which the
// user can rearrange the taskbar freely.

const (
	taskbarLayoutPath    = scriptsDir + `\unattend-taskbar-layout.xml`
	unlockVbsPath        = scriptsDir + `\unattend-unlock-start-layout.vbs`
	unlockTaskXMLPath    = scriptsDir + `\unattend-unlock-start-layout.xml`
	taskbarEventSource   = "UnattendGenerator"
	taskbarEventLogName  = "Application"
	taskbarLayoutKeyPath = `\Software\Policies\Microsoft\Windows\Explorer`
)

// taskbarIconsEmptyXML is the reference's "empty taskbar" layout: a
// single DesktopApp pin pointing at the "#leaveempty" sentinel replaces
// the default pin list with nothing.
const taskbarIconsEmptyXML = `<LayoutModificationTemplate xmlns="http://schemas.microsoft.com/Start/2014/LayoutModification" xmlns:defaultlayout="http://schemas.microsoft.com/Start/2014/FullDefaultLayout" xmlns:start="http://schemas.microsoft.com/Start/2014/StartLayout" xmlns:taskbar="http://schemas.microsoft.com/Start/2014/TaskbarLayout" Version="1">
  <CustomTaskbarLayoutCollection PinListPlacement="Replace">
    <defaultlayout:TaskbarLayout>
      <taskbar:TaskbarPinList>
        <taskbar:DesktopApp DesktopApplicationLinkPath="#leaveempty" />
      </taskbar:TaskbarPinList>
    </defaultlayout:TaskbarLayout>
  </CustomTaskbarLayoutCollection>
</LayoutModificationTemplate>
`

// unlockStartLayoutVBS is copied verbatim from resource/UnlockStartLayout.vbs
// (line endings normalised to CRLF when written).
const unlockStartLayoutVBS = `HKU = &H80000003
Set reg = GetObject("winmgmts://./root/default:StdRegProv")
Set fso = CreateObject("Scripting.FileSystemObject")

If reg.EnumKey(HKU, "", sids) = 0 Then
	If Not IsNull(sids) Then
		For Each sid In sids
			key = sid + "\Software\Policies\Microsoft\Windows\Explorer"
			name = "LockedStartLayout"
			If reg.GetDWORDValue(HKU, key, name, existing) = 0 Then
				reg.SetDWORDValue HKU, key, name, 0
			End If
		Next
	End If
End If`

// unlockStartLayoutTaskXML is copied from resource/UnlockStartLayout.xml;
// only the script path in <Arguments> differs (our own file name).
const unlockStartLayoutTaskXML = `<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
	<Triggers>
		<EventTrigger>
			<Enabled>true</Enabled>
			<Subscription>&lt;QueryList&gt;&lt;Query Id="0" Path="Application"&gt;&lt;Select Path="Application"&gt;*[System[Provider[@Name='UnattendGenerator'] and EventID=1]]&lt;/Select&gt;&lt;/Query&gt;&lt;/QueryList&gt;</Subscription>
		</EventTrigger>
	</Triggers>
	<Principals>
		<Principal id="Author">
			<UserId>S-1-5-18</UserId>
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
			<Command>C:\Windows\System32\wscript.exe</Command>
			<Arguments>C:\Windows\Setup\Scripts\unattend-unlock-start-layout.vbs</Arguments>
		</Exec>
	</Actions>
</Task>`

// taskbarIconsSpecializeScript runs in the specialize pass: policy that
// stops Windows from replacing the layout with cloud-optimized content,
// the event source the unlock flow keys on, and the unlock task itself.
func taskbarIconsSpecializeScript() string {
	return `reg.exe add "HKLM\Software\Policies\Microsoft\Windows\CloudContent" /v "DisableCloudOptimizedContent" /t REG_DWORD /d 1 /f;
[System.Diagnostics.EventLog]::CreateEventSource( '` + taskbarEventSource + `', '` + taskbarEventLogName + `' );
Register-ScheduledTask -TaskName 'UnlockStartLayout' -Xml $( Get-Content -LiteralPath "` + unlockTaskXMLPath + `" -Raw );
`
}

// taskbarIconsUserOnceScript runs once at first logon of every account:
// it raises the event that triggers the unlock task.
func taskbarIconsUserOnceScript() string {
	return `[System.Diagnostics.EventLog]::WriteEntry( '` + taskbarEventSource + `', "User '$env:USERNAME' has requested to unlock the Start menu layout.", [System.Diagnostics.EventLogEntryType]::Information, 1 );
`
}

// TaskbarIconsCommand returns one specialize-pass command that pins the
// requested taskbar icons for every future account and registers the
// unlock flow described above. Returns "" for TaskbarIconsModeDefault/""
// and for mode=custom without XML (validation rejects that earlier).
func TaskbarIconsCommand(s profile.TaskbarIconsSettings, hidePowerShellWindows bool) string {
	var layout string
	switch s.Mode {
	case profile.TaskbarIconsModeEmpty:
		layout = taskbarIconsEmptyXML
	case profile.TaskbarIconsModeCustom:
		if s.XML == nil {
			return ""
		}
		layout = *s.XML
	default:
		return ""
	}

	specializePath := scriptsDir + `\unattend-taskbar-icons.ps1`
	userOncePath := scriptsDir + `\unattend-taskbar-icons-uo.ps1`
	wrapperPath := scriptsDir + `\unattend-taskbar-icons-uo-run.cmd`
	wrapper := "@echo off\r\n" + invokeCommand(profile.ScriptPs1, userOncePath, hidePowerShellWindows) + "\r\n"
	layoutKey := defaultUserHiveKey + taskbarLayoutKeyPath

	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		writeFileStatement(taskbarLayoutPath, []byte(layout)),
		writeFileStatement(unlockVbsPath, []byte(strings.ReplaceAll(unlockStartLayoutVBS, "\n", "\r\n"))),
		writeFileStatement(unlockTaskXMLPath, []byte(unlockStartLayoutTaskXML)),
		writeFileStatement(specializePath, []byte(taskbarIconsSpecializeScript())),
		invokeCommand(profile.ScriptPs1, specializePath, hidePowerShellWindows),
		writeFileStatement(userOncePath, []byte(taskbarIconsUserOnceScript())),
		writeFileStatement(wrapperPath, []byte(wrapper)),
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		fmt.Sprintf(`reg.exe add "%s" /v StartLayoutFile /t REG_SZ /d "%s" /f`, layoutKey, taskbarLayoutPath),
		fmt.Sprintf(`reg.exe add "%s" /v LockedStartLayout /t REG_DWORD /d 1 /f`, layoutKey),
		fmt.Sprintf(`reg.exe add "%s" /v UnattendTaskbarIconsUnlock /d "%s" /f`, defaultUserRunOnceKey, wrapperPath),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}
