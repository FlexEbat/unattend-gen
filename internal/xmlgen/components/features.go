package components

import (
	"strings"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// featureCapabilityNamePrefixes maps a RemovableFeature to the prefix of
// its DISM capability Name (the part before the trailing "~~~lang~version",
// which changes across Windows releases — matching on the prefix is the
// resilient approach, same reasoning as apps.go's DisplayName matching).
var featureCapabilityNamePrefixes = map[profile.RemovableFeature]string{
	profile.FeatureInternetExplorer: "Browser.InternetExplorer",
	profile.FeatureWordPad:          "Microsoft.Windows.WordPad",
	profile.FeaturePowerShellISE:    "Microsoft.Windows.PowerShell.ISE",
	profile.FeatureOpenSSHClient:    "OpenSSH.Client",
	profile.FeatureMediaPlayer:      "Media.WindowsMediaPlayer",
	profile.FeatureSpeech:           "Language.Speech",
	profile.FeatureHandwriting:      "Language.Handwriting",
}

// RemoveFeaturesFirstLogonCommand returns the single command line that
// removes every feature in features via Remove-WindowsCapability, or "" if
// features is empty.
func RemoveFeaturesFirstLogonCommand(features []profile.RemovableFeature) string {
	var prefixes []string
	for _, f := range features {
		if p, ok := featureCapabilityNamePrefixes[f]; ok {
			prefixes = append(prefixes, p)
		}
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
