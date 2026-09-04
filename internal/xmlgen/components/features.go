package components

import (
	"strings"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// featureCapabilityNamePrefixes maps a RemovableFeature to the prefix(es)
// of its DISM capability Name (the part before the trailing
// "~~~lang~version", which changes across Windows releases — matching on
// the prefix is the resilient approach, same reasoning as apps.go's
// DisplayName matching). Most features have exactly one prefix; a few
// (Windows Hello) need several capabilities removed together.
var featureCapabilityNamePrefixes = map[profile.RemovableFeature][]string{
	profile.FeatureInternetExplorer: {"Browser.InternetExplorer"},
	profile.FeatureWordPad:          {"Microsoft.Windows.WordPad"},
	profile.FeaturePowerShellISE:    {"Microsoft.Windows.PowerShell.ISE"},
	profile.FeatureOpenSSHClient:    {"OpenSSH.Client"},
	profile.FeatureMediaPlayer:      {"Media.WindowsMediaPlayer"},
	profile.FeatureSpeech:           {"Language.Speech", "Language.TextToSpeech"},
	profile.FeatureHandwriting:      {"Language.Handwriting"},
	// Slice 17 additions (tech.md backlog group A). Selectors sourced from
	// github.com/cschneegans/unattend-generator's resource/Bloatware.json,
	// not invented from memory.
	profile.FeatureWindowsHello:   {"Hello.Face.18967", "Hello.Face.Migration.18967", "Hello.Face.20134"},
	profile.FeatureMathInputPanel: {"MathRecognizer"},
	profile.FeatureOneSync:        {"OneCoreUAP.OneSync"},
	profile.FeatureStepsRecorder:  {"App.StepsRecorder"},
}

// RemoveFeaturesFirstLogonCommand returns the single command line that
// removes every feature in features via Remove-WindowsCapability, or "" if
// features is empty.
func RemoveFeaturesFirstLogonCommand(features []profile.RemovableFeature) string {
	var prefixes []string
	for _, f := range features {
		prefixes = append(prefixes, featureCapabilityNamePrefixes[f]...)
	}
	if len(prefixes) == 0 {
		return ""
	}

	quoted := make([]string, len(prefixes))
	for i, p := range prefixes {
		quoted[i] = "'" + p + "'"
	}
	list := strings.Join(quoted, ",")

	return `powershell -NoProfile -Command "$patterns=@(` + list + `);foreach($p in $patterns){Get-WindowsCapability -Online | Where-Object {$_.Name -like \"$p*\"} | Remove-WindowsCapability -Online -ErrorAction SilentlyContinue}"`
}
