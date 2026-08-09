package components

import (
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
