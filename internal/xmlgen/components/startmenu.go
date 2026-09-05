package components

import (
	"encoding/hex"
	"fmt"

	"github.com/FlexEbat/unattend-gen/internal/profile"
)

// Slice 20 (tech.md backlog group C): Desktop Icons and Folders on Start —
// two more entirely-missing sections. Both target the live account's own
// HKCU (not the default-user hive's static content), so both go through
// the UserOnce/RunOnce mechanism: mount the default user hive just long
// enough to register a RunOnce entry, which then runs in the new account's
// own session at first logon and writes straight to that live HKCU.
// Mechanisms sourced from the reference implementation
// (github.com/cschneegans/unattend-generator, modifier/Optimizations.cs
// and resource/DesktopIcon.json), not invented from memory.

// desktopIconGUIDs maps a DesktopIcon to the CLSID Explorer uses for it
// under HideDesktopIcons. Copied from resource/DesktopIcon.json.
var desktopIconGUIDs = map[profile.DesktopIcon]string{
	profile.DesktopIconMusic:        "{1CF1260C-4DD0-4ebb-811F-33C572699FDE}",
	profile.DesktopIconThisPC:       "{20D04FE0-3AEA-1069-A2D8-08002B30309D}",
	profile.DesktopIconDownloads:    "{374DE290-123F-4565-9164-39C4925E467B}",
	profile.DesktopIconPictures:     "{3ADD1653-EB32-4cb0-BBD7-DFA0ABB5ACCA}",
	profile.DesktopIconControlPanel: "{5399E694-6CE5-4D6C-8FCE-1D8870FDCBA0}",
	profile.DesktopIconUserFiles:    "{59031a47-3f72-44a7-89c5-5595fe6b30ee}",
	profile.DesktopIconRecycleBin:   "{645FF040-5081-101B-9F08-00AA002F954E}",
	profile.DesktopIconVideos:       "{A0953C92-50DC-43bf-BE83-3742FED03C9C}",
	profile.DesktopIconDocuments:    "{A8CDFF1C-4878-43be-B5FD-F8091C1C60D0}",
	profile.DesktopIconDesktop:      "{B4BFCC3A-DB2C-424C-B029-7FE99A87C641}",
	profile.DesktopIconGallery:      "{e88865ea-0e1c-4e20-9aa6-edcd0212c87c}",
	profile.DesktopIconNetwork:      "{F02C1A0D-BE21-4350-88B0-7367FC96EF3C}",
	profile.DesktopIconHome:         "{f874310e-b6b7-47dc-bc84-b9e6b38f5903}",
}

// desktopIconRegistryKeys are the two HideDesktopIcons subkeys Windows
// reads depending on which Start menu style is active (classic vs the
// modern "New Start Panel", used since Vista) — the reference
// implementation writes both so the setting takes effect regardless.
var desktopIconRegistryKeys = []string{"ClassicStartMenu", "NewStartPanel"}

// DesktopIconsUserOnceCommand returns one specialize-pass command that
// mounts the default user hive just long enough to register a RunOnce
// entry; that entry runs at the live account's first logon and writes the
// icon visibility values (0 = shown, 1 = hidden) directly into that
// account's own HKCU, then restarts Explorer so the change is visible
// immediately. Applies to every future account, since it goes through
// RunOnce on the default-user template — not just the account created
// during setup. Returns "" if icons is empty.
func DesktopIconsUserOnceCommand(icons map[profile.DesktopIcon]bool) string {
	if len(icons) == 0 {
		return ""
	}
	lines := []string{"@echo off"}
	for _, key := range desktopIconRegistryKeys {
		regPath := `HKCU\Software\Microsoft\Windows\CurrentVersion\Explorer\HideDesktopIcons\` + key
		lines = append(lines, fmt.Sprintf(`reg.exe add "%s" /f`, regPath))
		for _, icon := range profile.DesktopIcons {
			show, ok := icons[icon]
			if !ok {
				continue
			}
			hidden := 1
			if show {
				hidden = 0
			}
			lines = append(lines, fmt.Sprintf(`reg.exe add "%s" /v "%s" /t REG_DWORD /d %d /f`, regPath, desktopIconGUIDs[icon], hidden))
		}
	}
	lines = append(lines, `taskkill /f /im explorer.exe`, `start explorer.exe`)
	content := ""
	for _, l := range lines {
		content += l + "\r\n"
	}

	path := scriptsDir + `\unattend-desktop-icons-uo.cmd`
	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		writeFileStatement(path, []byte(content)),
		fmt.Sprintf(`reg.exe add "%s" /v UnattendDesktopIcons /d "%s" /f`, defaultUserRunOnceKey, path),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}

// startFolderGUIDBytes are the raw 16-byte GUIDs Windows concatenates into
// the Start menu's VisiblePlaces value, one per pinned folder. Decoded
// from the base64 strings in resource/StartFolder.json.
var startFolderGUIDBytes = map[profile.StartFolder][]byte{
	profile.StartFolderSettings:       mustHexDecode("86087352aa5143429f7b2776584659d4"),
	profile.StartFolderFileExplorer:   mustHexDecode("bc248a140cd68942a0806ed9bba24882"),
	profile.StartFolderDocuments:      mustHexDecode("ced5342d5afa434582f222e6eaf7773c"),
	profile.StartFolderDownloads:      mustHexDecode("2fb367e3de895543bfce61f37b18a937"),
	profile.StartFolderMusic:          mustHexDecode("20060bb0517f324caa1e34cc547f7315"),
	profile.StartFolderPictures:       mustHexDecode("a0073f380ae8804cb05a86db845dbc4d"),
	profile.StartFolderVideos:         mustHexDecode("c5a5b342867df44280a493faca7a88b5"),
	profile.StartFolderNetwork:        mustHexDecode("448175fe0d08ae428bda34ed97b66394"),
	profile.StartFolderPersonalFolder: mustHexDecode("4ab0bd744af9684f8bd64398071da8bc"),
}

func mustHexDecode(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err) // unreachable: s is always one of the fixed literals above
	}
	return b
}

// StartFoldersUserOnceCommand returns one specialize-pass command that
// mounts the default user hive just long enough to register a RunOnce
// entry; that entry runs at the live account's first logon and writes the
// concatenated GUIDs of folders, in order, as the Start menu's
// VisiblePlaces value in that account's own HKCU. Applies to every future
// account, same RunOnce mechanism as DesktopIconsUserOnceCommand. Returns
// "" if folders is empty — note this means "pin zero folders" can't be
// expressed, only "leave Windows' own default set alone".
func StartFoldersUserOnceCommand(folders []profile.StartFolder) string {
	if len(folders) == 0 {
		return ""
	}
	var buf []byte
	for _, f := range folders {
		buf = append(buf, startFolderGUIDBytes[f]...)
	}
	content := "@echo off\r\n" + fmt.Sprintf(`reg.exe add "HKCU\Software\Microsoft\Windows\CurrentVersion\Start" /v VisiblePlaces /t REG_BINARY /d %s /f`, hex.EncodeToString(buf)) + "\r\n"

	path := scriptsDir + `\unattend-start-folders-uo.cmd`
	return wrapCommand([]string{
		ensureScriptsDirStatement(),
		fmt.Sprintf(`reg.exe load %s "%s"`, defaultUserHiveKey, defaultUserHivePath),
		writeFileStatement(path, []byte(content)),
		fmt.Sprintf(`reg.exe add "%s" /v UnattendStartFolders /d "%s" /f`, defaultUserRunOnceKey, path),
		fmt.Sprintf(`reg.exe unload %s`, defaultUserHiveKey),
	})
}
