package components

import (
	"strings"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// optionalFeatureNames maps a RemovableOptionalFeature to its
// Get-WindowsOptionalFeature FeatureName — an exact match, unlike
// RemovableApp's DisplayName substrings or RemovableFeature's capability
// Name prefixes: optional feature names are stable across Windows
// versions, they don't carry a trailing "~~~lang~version" suffix.
// Sourced from github.com/cschneegans/unattend-generator's
// resource/Bloatware.json, not invented from memory.
var optionalFeatureNames = map[profile.RemovableOptionalFeature]string{
	profile.OptionalFeatureRecall:              "Recall",
	profile.OptionalFeatureMediaFeatures:       "MediaPlayback",
	profile.OptionalFeatureRemoteDesktopClient: "Microsoft-RemoteDesktopConnection",
}

// RemoveOptionalFeaturesFirstLogonCommand returns the single command line
// that removes every feature in features via
// Disable-WindowsOptionalFeature, or "" if features is empty. This is a
// third, separate removal mechanism from RemoveAppsFirstLogonCommand
// (Appx) and RemoveFeaturesFirstLogonCommand (DISM capability) — same
// oobeSystem/FirstLogonCommands family, different cmdlets.
func RemoveOptionalFeaturesFirstLogonCommand(features []profile.RemovableOptionalFeature) string {
	var names []string
	for _, f := range features {
		if n, ok := optionalFeatureNames[f]; ok {
			names = append(names, n)
		}
	}
	if len(names) == 0 {
		return ""
	}

	quoted := make([]string, len(names))
	for i, n := range names {
		quoted[i] = "'" + n + "'"
	}
	list := strings.Join(quoted, ",")

	return `powershell -NoProfile -Command "$names=@(` + list + `);foreach($n in $names){Get-WindowsOptionalFeature -Online | Where-Object {$_.FeatureName -eq $n} | Disable-WindowsOptionalFeature -Online -Remove -NoRestart -ErrorAction SilentlyContinue}"`
}
