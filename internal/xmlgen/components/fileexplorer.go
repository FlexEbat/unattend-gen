package components

import (
	"fmt"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

const (
	explorerAdvancedKey   = defaultUserHiveKey + `\Software\Microsoft\Windows\CurrentVersion\Explorer\Advanced`
	classicContextMenuKey = defaultUserHiveKey + `\Software\Classes\CLSID\{86ca1aa0-34aa-4e8b-a509-50c905bae2a2}\InprocServer32`
)

// FileExplorerCommand returns one command that mounts the default user
// hive, applies s, and unmounts it — the same DefaultUser mechanism as
// DefaultUserScripts, so every account (including future ones) gets these
// File Explorer settings. Returns "" when s is the zero value: nothing to
// change.
func FileExplorerCommand(s profile.FileExplorerSettings) string {
	var ops []string

	if s.HiddenFiles == profile.HiddenFilesShowHidden || s.HiddenFiles == profile.HiddenFilesShowAll {
		superHidden := 0
		if s.HiddenFiles == profile.HiddenFilesShowAll {
			superHidden = 1
		}
		ops = append(ops,
			fmt.Sprintf(`reg.exe add "%s" /v Hidden /t REG_DWORD /d 1 /f`, explorerAdvancedKey),
			fmt.Sprintf(`reg.exe add "%s" /v ShowSuperHidden /t REG_DWORD /d %d /f`, explorerAdvancedKey, superHidden),
		)
	}
	if s.ShowFileExtensions {
		ops = append(ops, fmt.Sprintf(`reg.exe add "%s" /v HideFileExt /t REG_DWORD /d 0 /f`, explorerAdvancedKey))
	}
	if s.ClassicContextMenu {
		ops = append(ops, fmt.Sprintf(`reg.exe add "%s" /f /ve`, classicContextMenuKey))
	}
	if s.HideFolderTooltips {
		ops = append(ops, fmt.Sprintf(`reg.exe add "%s" /v ShowInfoTip /t REG_DWORD /d 0 /f`, explorerAdvancedKey))
	}
	if s.OpenToThisPC {
		ops = append(ops, fmt.Sprintf(`reg.exe add "%s" /v LaunchTo /t REG_DWORD /d 1 /f`, explorerAdvancedKey))
	}
	if s.ShowEndTaskInTaskbar {
		ops = append(ops, fmt.Sprintf(`reg.exe add "%s" /v TaskbarEndTask /t REG_DWORD /d 1 /f`, explorerAdvancedKey))
	}

	if len(ops) == 0 {
		return ""
	}

	statements := append([]string{fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath)}, ops...)
	statements = append(statements, fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey))
	return wrapCommand(statements)
}
