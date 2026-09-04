package components

import (
	"fmt"
	"strings"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// appDisplayNamePatterns maps a RemovableApp to the substring(s) matched
// against Get-AppxProvisionedPackage's DisplayName. DisplayName substrings
// are used instead of exact PackageFamilyNames because Microsoft changes
// the full package names across Windows versions; DisplayName substrings
// are the same resilient pattern used throughout the Windows admin
// community for this reason. One RemovableApp can map to several patterns
// (a "family" of related packages).
var appDisplayNamePatterns = map[profile.RemovableApp][]string{
	profile.App3DViewer:            {"3DViewer"},
	profile.AppCalculator:          {"WindowsCalculator"},
	profile.AppCamera:              {"WindowsCamera"},
	profile.AppClipchamp:           {"Clipchamp"},
	profile.AppClock:               {"WindowsAlarms"},
	profile.AppCopilot:             {"Copilot"},
	profile.AppCortana:             {"549981C3F5F10", "Cortana"},
	profile.AppFamily:              {"MicrosoftFamily"},
	profile.AppFeedbackHub:         {"WindowsFeedbackHub"},
	profile.AppGetHelp:             {"GetHelp"},
	profile.AppMailAndCalendar:     {"windowscommunicationsapps"},
	profile.AppMaps:                {"WindowsMaps"},
	profile.AppMixedReality:        {"MixedReality"},
	profile.AppMoviesAndTV:         {"ZuneVideo"},
	profile.AppNews:                {"BingNews"},
	profile.AppOffice:              {"MicrosoftOfficeHub"},
	profile.AppOneNote:             {"OneNote"},
	profile.AppPaint3D:             {"Paint3D"},
	profile.AppPeople:              {"People"},
	profile.AppPhoneLink:           {"YourPhone"},
	profile.AppPowerAutomate:       {"PowerAutomateDesktop"},
	profile.AppQuickAssist:         {"QuickAssist"},
	profile.AppSkype:               {"SkypeApp"},
	profile.AppSnippingTool:        {"ScreenSketch"},
	profile.AppSolitaireCollection: {"MicrosoftSolitaireCollection"},
	profile.AppStickyNotes:         {"MicrosoftStickyNotes"},
	profile.AppTeams:               {"MicrosoftTeams", "Teams"},
	profile.AppTips:                {"Getstarted"},
	profile.AppToDo:                {"Todos"},
	profile.AppVoiceRecorder:       {"WindowsSoundRecorder"},
	profile.AppWeather:             {"BingWeather"},
	profile.AppXboxApps:            {"Xbox"},
	// Slice 17 additions (tech.md backlog group A). Selectors sourced from
	// github.com/cschneegans/unattend-generator's resource/Bloatware.json,
	// not invented from memory.
	profile.AppBingSearch:  {"BingSearch"},
	profile.AppDevHome:     {"DevHome"},
	profile.AppGameAssist:  {"Edge.GameAssist"},
	profile.AppStore:       {"WindowsStore", "StorePurchaseApp"},
	profile.AppNotepad:     {"WindowsNotepad"},
	profile.AppOutlook:     {"OutlookForWindows"},
	profile.AppPaint:       {"Microsoft.Paint"}, // distinct from AppPaint3D's "Paint3D" pattern
	profile.AppWallet:      {"Wallet"},
	profile.AppMediaPlayer: {"ZuneMusic"},
	profile.AppTerminal:    {"WindowsTerminal"},
	// AppOneDrive is deliberately absent here: it's not distributed as a
	// provisioned Appx package, see RemoveOneDrive*Command below.
}

// RemoveAppsFirstLogonCommand returns the single command line that removes
// every app in apps, or "" if apps is empty. It runs
// Get-AppxProvisionedPackage/Remove-AppxProvisionedPackage, which works on
// every Windows 11 edition — unlike the "Remove default Microsoft Store
// packages" Intune/Group Policy added in 2026, which is Enterprise/Education
// only.
func RemoveAppsFirstLogonCommand(apps []profile.RemovableApp) string {
	var patterns []string
	for _, app := range apps {
		patterns = append(patterns, appDisplayNamePatterns[app]...)
	}
	if len(patterns) == 0 {
		return ""
	}

	quoted := make([]string, len(patterns))
	for i, p := range patterns {
		quoted[i] = "'" + p + "'"
	}
	list := strings.Join(quoted, ",")

	return `powershell -NoProfile -Command "$patterns=@(` + list + `);foreach($p in $patterns){Get-AppxProvisionedPackage -Online | Where-Object {$_.DisplayName -like \"*$p*\"} | Remove-AppxProvisionedPackage -Online -ErrorAction SilentlyContinue}"`
}

// removeOneDriveFilesScript deletes the leftover OneDrive Start Menu
// shortcut and setup executables. OneDrive isn't a provisioned Appx
// package (it ships as a separate installer), so it can't go through
// appDisplayNamePatterns/Remove-AppxProvisionedPackage like the other
// RemovableApp entries. Sourced from
// github.com/cschneegans/unattend-generator, modifier/Bloatware.cs
// (bw.Id == "RemoveOneDrive"), not invented from memory.
const removeOneDriveFilesScript = `@(
	'C:\Users\Default\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\OneDrive.lnk';
	'C:\Windows\System32\OneDriveSetup.exe';
	'C:\Windows\SysWOW64\OneDriveSetup.exe';
) | Where-Object -FilterScript { [System.IO.File]::Exists( $_ ); } | Remove-Item -Verbose -ErrorAction 'Continue';
`

// RemoveOneDriveFilesCommand returns one specialize-pass command deleting
// OneDrive's leftover shortcut and setup executables.
func RemoveOneDriveFilesCommand() string {
	path := scriptsDir + `\unattend-remove-onedrive.ps1`
	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		writeFileStatement(path, []byte(removeOneDriveFilesScript)),
		invokeCommand(profile.ScriptPs1, path),
	})
}

// RemoveOneDriveDefaultUserCommand returns one specialize-pass command
// that mounts the default user hive and removes the OneDriveSetup autorun
// entry, so OneDrive doesn't reinstall itself for future accounts either.
func RemoveOneDriveDefaultUserCommand() string {
	key := defaultUserHiveKey + `\Software\Microsoft\Windows\CurrentVersion\Run`
	return wrapCommand([]string{
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		fmt.Sprintf(`reg.exe delete "%s" /v OneDriveSetup /f`, key),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}

// ContainsApp reports whether apps contains target.
func ContainsApp(apps []profile.RemovableApp, target profile.RemovableApp) bool {
	for _, a := range apps {
		if a == target {
			return true
		}
	}
	return false
}
